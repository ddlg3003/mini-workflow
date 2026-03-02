package service

import (
	"context"
	"encoding/json"
	"time"

	pb "mini-workflow/api"
	"mini-workflow/history/internal/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *historyService) RecordActivityTaskStarted(ctx context.Context, req *pb.RecordActivityTaskStartedRequest) (*pb.RecordActivityTaskStartedResponse, error) {
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
	now := time.Now().UTC()

	payload, _ := json.Marshal(map[string]any{
		"activity_id": req.ActivityId,
		"started_at":  now,
	})
	event := domain.HistoryEvent{
		Namespace:  exec.Namespace,
		WorkflowID: exec.WorkflowID,
		RunID:      exec.RunID,
		EventID:    eventID,
		EventType:  "ActivityTaskStarted",
		Payload:    payload,
	}
	exec.NextEventID = eventID + 1

	actState := domain.ActivityState{
		Namespace:     exec.Namespace,
		WorkflowID:    exec.WorkflowID,
		RunID:         exec.RunID,
		ActivityID:    req.ActivityId,
		Status:        domain.ActivityStatusStarted,
		Attempt:       1,
		LastHeartbeat: &now,
	}

	if err := s.repo.UpdateWorkflowExecution(ctx, exec, expectedVersion, []domain.HistoryEvent{event}, []domain.ActivityState{actState}, nil); err != nil {
		return nil, status.Errorf(codes.Internal, "update execution: %v", err)
	}

	return &pb.RecordActivityTaskStartedResponse{}, nil
}
