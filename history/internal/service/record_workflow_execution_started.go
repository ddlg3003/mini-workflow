package service

import (
	"context"
	"encoding/json"
	"fmt"

	pb "mini-workflow/api"
	"mini-workflow/history/internal/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *historyService) RecordWorkflowExecutionStarted(ctx context.Context, req *pb.RecordWorkflowExecutionStartedRequest) (*pb.RecordWorkflowExecutionStartedResponse, error) {
	sr := req.StartRequest
	if sr == nil || sr.WorkflowId == "" || sr.WorkflowType == "" {
		return nil, status.Error(codes.InvalidArgument, "start_request with workflow_id and workflow_type is required")
	}

	existing, err := s.repo.FindRunningWorkflow(ctx, "default", sr.WorkflowId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "check duplicate: %v", err)
	}
	if existing != nil {
		return nil, status.Errorf(codes.AlreadyExists, "workflow %q is already running (run_id: %s)", sr.WorkflowId, existing.RunID)
	}

	runID := uuid.New()

	payload, err := json.Marshal(map[string]any{
		"workflow_type": sr.WorkflowType,
		"run_id":        runID.String(),
		"task_queue":    sr.TaskQueue,
		"input":         sr.Input,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "marshal event payload: %v", err)
	}

	exec := &domain.WorkflowExecution{
		Namespace:      "default",
		WorkflowID:     sr.WorkflowId,
		RunID:          runID,
		WorkflowType:   sr.WorkflowType,
		TaskQueue:      sr.TaskQueue,
		Status:         domain.WorkflowStatusRunning,
		CurrentVersion: 1,
		NextEventID:    2,
		Input:          sr.Input,
	}

	initialEvent := &domain.HistoryEvent{
		Namespace:  "default",
		WorkflowID: sr.WorkflowId,
		RunID:      runID,
		EventID:    1,
		EventType:  "WorkflowExecutionStarted",
		Payload:    payload,
	}

	if err := s.repo.CreateWorkflowExecution(ctx, exec, initialEvent); err != nil {
		return nil, status.Errorf(codes.Internal, "create workflow execution: %v", err)
	}

	if _, err := s.matching.AddWorkflowTask(ctx, &pb.AddWorkflowTaskRequest{
		TaskQueue:    sr.TaskQueue,
		WorkflowId:   sr.WorkflowId,
		RunId:        runID.String(),
		WorkflowType: sr.WorkflowType,
	}); err != nil {
		s.log.Warn("failed to enqueue initial workflow task", zap.String("workflow_id", sr.WorkflowId), zap.Error(err))
	}

	return &pb.RecordWorkflowExecutionStartedResponse{RunId: runID.String()}, nil
}

func runIDFromString(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("invalid run_id %q: %w", s, err)
	}
	return id, nil
}
