package service

import (
	"context"
	"encoding/json"

	pb "mini-workflow/api"
	"mini-workflow/history/internal/domain"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *historyService) RecordActivityTaskCompleted(ctx context.Context, req *pb.RecordActivityTaskCompletedRequest) (*pb.RecordActivityTaskCompletedResponse, error) {
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

	payload, _ := json.Marshal(map[string]any{
		"activity_id": req.ActivityId,
		"result":      req.Result,
	})
	event := domain.HistoryEvent{
		Namespace:  exec.Namespace,
		WorkflowID: exec.WorkflowID,
		RunID:      exec.RunID,
		EventID:    eventID,
		EventType:  "ActivityTaskCompleted",
		Payload:    payload,
	}
	exec.NextEventID = eventID + 1

	actState, err := s.repo.GetActivityState(ctx, "default", req.WorkflowId, runID, req.ActivityId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get activity state: %v", err)
	}

	actState.Status = domain.ActivityStatusCompleted
	actState.Attempt = 1

	if err := s.repo.UpdateWorkflowExecution(ctx, exec, expectedVersion, []domain.HistoryEvent{event}, []domain.ActivityState{*actState}, nil); err != nil {
		return nil, status.Errorf(codes.Internal, "update execution: %v", err)
	}

	if _, err := s.matching.AddWorkflowTask(ctx, &pb.AddWorkflowTaskRequest{
		TaskQueue:    exec.TaskQueue,
		WorkflowId:   exec.WorkflowID,
		RunId:        exec.RunID.String(),
		WorkflowType: exec.WorkflowType,
	}); err != nil {
		s.log.Warn("failed to enqueue workflow task after activity completion",
			zap.String("activity_id", req.ActivityId), zap.Error(err))
	}

	return &pb.RecordActivityTaskCompletedResponse{}, nil
}
