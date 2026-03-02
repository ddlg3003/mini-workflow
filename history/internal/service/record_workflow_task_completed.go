package service

import (
	"context"
	"encoding/json"
	"time"

	pb "mini-workflow/api"
	"mini-workflow/history/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	maxActivityAttempts = 3
	retryBackoffSeconds = 5
)

func (s *historyService) RecordWorkflowTaskCompleted(ctx context.Context, req *pb.RecordWorkflowTaskCompletedHistoryRequest) (*pb.RecordWorkflowTaskCompletedHistoryResponse, error) {
	runID, err := runIDFromString(req.RunId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	exec, err := s.repo.GetWorkflowExecution(ctx, "default", req.WorkflowId, runID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "workflow execution not found: %v", err)
	}
	if exec.Status != domain.WorkflowStatusRunning {
		return nil, status.Errorf(codes.FailedPrecondition, "workflow is not running (status: %s)", exec.Status)
	}

	expectedVersion := exec.CurrentVersion
	eventID := exec.NextEventID

	completedPayload, _ := json.Marshal(map[string]any{"command_count": len(req.Commands)})
	events := []domain.HistoryEvent{
		{
			Namespace:  exec.Namespace,
			WorkflowID: exec.WorkflowID,
			RunID:      exec.RunID,
			EventID:    eventID,
			EventType:  "WorkflowTaskCompleted",
			Payload:    completedPayload,
		},
	}
	eventID++

	var activities []domain.ActivityState
	var timers []domain.Timer
	var activityTasksToEnqueue []*pb.AddActivityTaskRequest

	for _, cmd := range req.Commands {
		switch cmd.CommandType {
		case "ScheduleActivityTask":
			var attrs struct {
				ActivityID              string `json:"activity_id"`
				ActivityType            string `json:"activity_type"`
				Input                   []byte `json:"input"`
				TaskQueue               string `json:"task_queue"`
				HeartbeatTimeoutSeconds int    `json:"heartbeat_timeout_seconds"`
			}
			if err := json.Unmarshal(cmd.Attributes, &attrs); err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "parse ScheduleActivityTask attributes: %v", err)
			}
			if attrs.HeartbeatTimeoutSeconds == 0 {
				attrs.HeartbeatTimeoutSeconds = 30
			}

			payload, _ := json.Marshal(map[string]any{
				"activity_id":   attrs.ActivityID,
				"activity_type": attrs.ActivityType,
				"input":         attrs.Input,
			})
			events = append(events, domain.HistoryEvent{
				Namespace:  exec.Namespace,
				WorkflowID: exec.WorkflowID,
				RunID:      exec.RunID,
				EventID:    eventID,
				EventType:  "ActivityTaskScheduled",
				Payload:    payload,
			})
			eventID++

			tq := attrs.TaskQueue
			if tq == "" {
				tq = exec.TaskQueue
			}
			activities = append(activities, domain.ActivityState{
				Namespace:               exec.Namespace,
				WorkflowID:              exec.WorkflowID,
				RunID:                   exec.RunID,
				ActivityID:              attrs.ActivityID,
				Status:                  domain.ActivityStatusScheduled,
				Attempt:                 1,
				HeartbeatTimeoutSeconds: attrs.HeartbeatTimeoutSeconds,
				TaskQueue:               tq,
				ActivityType:            attrs.ActivityType,
				Input:                   attrs.Input,
			})

			// Create a heartbeat-deadline timer so the processor can detect worker silence.
			hbTimerID := uuid.New()
			hbFireTime := time.Now().Add(time.Duration(attrs.HeartbeatTimeoutSeconds) * time.Second)
			hbToken, _ := json.Marshal(map[string]any{
				"activity_id":               attrs.ActivityID,
				"heartbeat_timeout_seconds": attrs.HeartbeatTimeoutSeconds,
			})
			timers = append(timers, domain.Timer{
				TimerID:    hbTimerID,
				Namespace:  exec.Namespace,
				WorkflowID: exec.WorkflowID,
				RunID:      exec.RunID,
				FireTime:   hbFireTime,
				TimerType:  domain.TimerTypeActivityTimeout,
				TaskToken:  hbToken,
			})

			activityTasksToEnqueue = append(activityTasksToEnqueue, &pb.AddActivityTaskRequest{
				TaskQueue:               tq,
				WorkflowId:              exec.WorkflowID,
				RunId:                   exec.RunID.String(),
				ActivityId:              attrs.ActivityID,
				ActivityType:            attrs.ActivityType,
				Input:                   attrs.Input,
				HeartbeatTimeoutSeconds: int32(attrs.HeartbeatTimeoutSeconds),
			})

		case "StartTimer":
			var attrs struct {
				TimerID       string `json:"timer_id"`
				FireAfterSecs int64  `json:"fire_after_seconds"`
			}
			if err := json.Unmarshal(cmd.Attributes, &attrs); err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "parse StartTimer attributes: %v", err)
			}
			timerID := uuid.New()
			fireTime := time.Now().Add(time.Duration(attrs.FireAfterSecs) * time.Second)
			payload, _ := json.Marshal(map[string]any{
				"timer_id":  timerID.String(),
				"fire_time": fireTime.UTC(),
			})
			events = append(events, domain.HistoryEvent{
				Namespace:  exec.Namespace,
				WorkflowID: exec.WorkflowID,
				RunID:      exec.RunID,
				EventID:    eventID,
				EventType:  "TimerStarted",
				Payload:    payload,
			})
			eventID++
			timers = append(timers, domain.Timer{
				TimerID:    timerID,
				Namespace:  exec.Namespace,
				WorkflowID: exec.WorkflowID,
				RunID:      exec.RunID,
				FireTime:   fireTime,
				TimerType:  domain.TimerTypeUser,
			})

		case "CompleteWorkflowExecution":
			var attrs struct {
				Result []byte `json:"result"`
			}
			json.Unmarshal(cmd.Attributes, &attrs)
			payload, _ := json.Marshal(map[string]any{"result": attrs.Result})
			events = append(events, domain.HistoryEvent{
				Namespace:  exec.Namespace,
				WorkflowID: exec.WorkflowID,
				RunID:      exec.RunID,
				EventID:    eventID,
				EventType:  "WorkflowExecutionCompleted",
				Payload:    payload,
			})
			eventID++
			exec.Status = domain.WorkflowStatusCompleted
			resRaw := json.RawMessage(attrs.Result)
			exec.Result = &resRaw
		}
	}

	exec.NextEventID = eventID
	if err := s.repo.UpdateWorkflowExecution(ctx, exec, expectedVersion, events, activities, timers); err != nil {
		return nil, status.Errorf(codes.Internal, "update execution: %v", err)
	}

	for _, t := range timers {
		if err := s.timerStore.ScheduleTimer(ctx, t.TimerID.String(), t.FireTime.UnixMilli()); err != nil {
			s.log.Warn("failed to schedule timer in redis", zap.String("timer_id", t.TimerID.String()), zap.Error(err))
		}
	}

	for _, req := range activityTasksToEnqueue {
		if _, err := s.matching.AddActivityTask(ctx, req); err != nil {
			s.log.Warn("failed to enqueue activity task", zap.String("activity_id", req.ActivityId), zap.Error(err))
		}
	}

	return &pb.RecordWorkflowTaskCompletedHistoryResponse{}, nil
}
