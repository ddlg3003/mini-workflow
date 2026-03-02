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

func (s *historyService) RecordWorkflowTaskStarted(ctx context.Context, req *pb.RecordWorkflowTaskStartedRequest) (*pb.RecordWorkflowTaskStartedResponse, error) {
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

	payload, _ := json.Marshal(map[string]any{"started_at": time.Now().UTC()})

	event := domain.HistoryEvent{
		Namespace:  exec.Namespace,
		WorkflowID: exec.WorkflowID,
		RunID:      exec.RunID,
		EventID:    eventID,
		EventType:  "WorkflowTaskStarted",
		Payload:    payload,
	}

	exec.NextEventID = eventID + 1

	// Create a Timer to detect if the worker crashes before completing the task.
	timerID := uuid.New()
	// Default to 10 seconds for Task Timeout as specified
	fireTime := time.Now().Add(10 * time.Second)
	// We embed the eventID in the token to ensure we only timeout THIS specific task attempt
	taskToken, _ := json.Marshal(map[string]any{
		"event_id": eventID,
	})

	timer := domain.Timer{
		TimerID:    timerID,
		Namespace:  exec.Namespace,
		WorkflowID: exec.WorkflowID,
		RunID:      exec.RunID,
		FireTime:   fireTime,
		TimerType:  domain.TimerTypeWorkflowTaskTimeout,
		TaskToken:  taskToken,
	}

	if err := s.repo.UpdateWorkflowExecution(ctx, exec, expectedVersion, []domain.HistoryEvent{event}, nil, []domain.Timer{timer}); err != nil {
		return nil, status.Errorf(codes.Internal, "update execution: %v", err)
	}

	if err := s.timerStore.ScheduleTimer(ctx, timer.TimerID.String(), timer.FireTime.UnixMilli()); err != nil {
		s.log.Warn("failed to schedule workflow task timeout timer in redis", zap.String("timer_id", timer.TimerID.String()), zap.Error(err))
	}

	return &pb.RecordWorkflowTaskStartedResponse{}, nil
}
