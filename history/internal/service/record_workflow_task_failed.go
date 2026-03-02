package service

import (
	"context"
	"encoding/json"

	pb "mini-workflow/api"
	"mini-workflow/history/internal/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *historyService) RecordWorkflowTaskFailed(ctx context.Context, req *pb.RecordWorkflowTaskFailedHistoryRequest) (*pb.RecordWorkflowTaskFailedHistoryResponse, error) {
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

	payload, _ := json.Marshal(map[string]any{"cause": req.Cause})
	event1 := domain.HistoryEvent{
		Namespace:  exec.Namespace,
		WorkflowID: exec.WorkflowID,
		RunID:      exec.RunID,
		EventID:    eventID,
		EventType:  "WorkflowTaskFailed",
		Payload:    payload,
	}

	event2 := domain.HistoryEvent{
		Namespace:  exec.Namespace,
		WorkflowID: exec.WorkflowID,
		RunID:      exec.RunID,
		EventID:    eventID + 1,
		EventType:  "WorkflowExecutionFailed",
		Payload:    payload,
	}

	exec.NextEventID = eventID + 2
	exec.Status = domain.WorkflowStatusFailed

	if err := s.repo.UpdateWorkflowExecution(ctx, exec, expectedVersion, []domain.HistoryEvent{event1, event2}, nil, nil); err != nil {
		return nil, status.Errorf(codes.Internal, "update execution: %v", err)
	}

	return &pb.RecordWorkflowTaskFailedHistoryResponse{}, nil
}
