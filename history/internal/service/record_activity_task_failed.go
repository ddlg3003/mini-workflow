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

func (s *historyService) RecordActivityTaskFailed(ctx context.Context, req *pb.RecordActivityTaskFailedRequest) (*pb.RecordActivityTaskFailedResponse, error) {
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

	actState, err := s.repo.GetActivityState(ctx, "default", req.WorkflowId, runID, req.ActivityId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get activity state: %v", err)
	}

	expectedVersion := exec.CurrentVersion
	eventID := exec.NextEventID
	failureReason := req.Reason

	if actState.Attempt < maxActivityAttempts {
		fireTime := time.Now().Add(time.Duration(retryBackoffSeconds) * time.Second)
		timerID := uuid.New()

		payload, _ := json.Marshal(map[string]any{
			"activity_id": req.ActivityId,
			"attempt":     actState.Attempt,
			"reason":      req.Reason,
		})
		event := domain.HistoryEvent{
			Namespace:  exec.Namespace,
			WorkflowID: exec.WorkflowID,
			RunID:      exec.RunID,
			EventID:    eventID,
			EventType:  "ActivityTaskFailed",
			Payload:    payload,
		}
		eventID++

		actState.Attempt++
		actState.Status = domain.ActivityStatusFailed
		actState.LastFailureReason = &failureReason

		timerToken, _ := json.Marshal(map[string]any{
			"activity_id":               req.ActivityId,
			"heartbeat_timeout_seconds": actState.HeartbeatTimeoutSeconds,
		})

		retryTimer := domain.Timer{
			TimerID:    timerID,
			Namespace:  exec.Namespace,
			WorkflowID: exec.WorkflowID,
			RunID:      exec.RunID,
			FireTime:   fireTime,
			TimerType:  domain.TimerTypeActivityTimeout,
			TaskToken:  timerToken,
		}
		exec.NextEventID = eventID

		if err := s.repo.UpdateWorkflowExecution(ctx, exec, expectedVersion, []domain.HistoryEvent{event}, []domain.ActivityState{*actState}, []domain.Timer{retryTimer}); err != nil {
			return nil, status.Errorf(codes.Internal, "update execution: %v", err)
		}

		if err := s.timerStore.ScheduleTimer(ctx, timerID.String(), fireTime.UnixMilli()); err != nil {
			s.log.Warn("failed to schedule retry timer", zap.String("timer_id", timerID.String()), zap.Error(err))
		}
	} else {
		payload, _ := json.Marshal(map[string]any{
			"activity_id": req.ActivityId,
			"reason":      req.Reason,
			"final":       true,
		})
		event := domain.HistoryEvent{
			Namespace:  exec.Namespace,
			WorkflowID: exec.WorkflowID,
			RunID:      exec.RunID,
			EventID:    eventID,
			EventType:  "ActivityTaskFailed",
			Payload:    payload,
		}
		eventID++

		actState.Status = domain.ActivityStatusFailed
		actState.LastFailureReason = &failureReason
		exec.NextEventID = eventID

		if err := s.repo.UpdateWorkflowExecution(ctx, exec, expectedVersion, []domain.HistoryEvent{event}, []domain.ActivityState{*actState}, nil); err != nil {
			return nil, status.Errorf(codes.Internal, "update execution: %v", err)
		}

		if _, err := s.matching.AddWorkflowTask(ctx, &pb.AddWorkflowTaskRequest{
			TaskQueue:    exec.TaskQueue,
			WorkflowId:   exec.WorkflowID,
			RunId:        exec.RunID.String(),
			WorkflowType: exec.WorkflowType,
		}); err != nil {
			s.log.Warn("failed to enqueue workflow task after final activity failure",
				zap.String("activity_id", req.ActivityId), zap.Error(err))
		}
	}

	return &pb.RecordActivityTaskFailedResponse{}, nil
}
