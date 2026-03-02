package service

import (
	"context"
	"encoding/json"
	"fmt"

	pb "mini-workflow/api"
	"mini-workflow/config"
	"mini-workflow/matching/internal/domain"
	"mini-workflow/matching/internal/ports"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type matchingService struct {
	queue ports.TaskQueue
	cfg   config.ServiceConfig
}

func New(queue ports.TaskQueue, cfg config.ServiceConfig) *matchingService {
	return &matchingService{queue: queue, cfg: cfg}
}

func queueKey(namespace, taskQueue, kind string) string {
	return fmt.Sprintf("task_queue:%s:%s:%s", namespace, taskQueue, kind)
}

const defaultNamespace = "default"

func (s *matchingService) AddWorkflowTask(ctx context.Context, req *pb.AddWorkflowTaskRequest) (*pb.AddWorkflowTaskResponse, error) {
	if req.WorkflowId == "" || req.RunId == "" {
		return nil, status.Error(codes.InvalidArgument, "workflow_id and run_id are required")
	}
	tq := req.TaskQueue
	if tq == "" {
		return nil, status.Error(codes.InvalidArgument, "task_queue is required")
	}

	payload, err := json.Marshal(domain.WorkflowTask{
		WorkflowID:   req.WorkflowId,
		RunID:        req.RunId,
		WorkflowType: req.WorkflowType,
		TaskQueue:    tq,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "marshal workflow task: %v", err)
	}

	key := queueKey(defaultNamespace, tq, "workflow")
	if err := s.queue.Push(ctx, key, payload); err != nil {
		return nil, status.Errorf(codes.Internal, "push workflow task: %v", err)
	}
	return &pb.AddWorkflowTaskResponse{}, nil
}

func (s *matchingService) AddActivityTask(ctx context.Context, req *pb.AddActivityTaskRequest) (*pb.AddActivityTaskResponse, error) {
	if req.WorkflowId == "" || req.RunId == "" || req.ActivityId == "" {
		return nil, status.Error(codes.InvalidArgument, "workflow_id, run_id, and activity_id are required")
	}

	tq := req.TaskQueue
	if tq == "" {
		return nil, status.Error(codes.InvalidArgument, "task_queue is required")
	}

	payload, err := json.Marshal(domain.ActivityTask{
		WorkflowID:              req.WorkflowId,
		RunID:                   req.RunId,
		ActivityID:              req.ActivityId,
		ActivityType:            req.ActivityType,
		Input:                   req.Input,
		TaskQueue:               tq,
		HeartbeatTimeoutSeconds: req.HeartbeatTimeoutSeconds,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "marshal activity task: %v", err)
	}

	key := queueKey(defaultNamespace, tq, "activity")
	if err := s.queue.Push(ctx, key, payload); err != nil {
		return nil, status.Errorf(codes.Internal, "push activity task: %v", err)
	}
	return &pb.AddActivityTaskResponse{}, nil
}

func (s *matchingService) PollWorkflowTaskQueue(ctx context.Context, req *pb.PollWorkflowTaskQueueRequest) (*pb.PollWorkflowTaskQueueResponse, error) {
	if req.TaskQueue == "" {
		return nil, status.Error(codes.InvalidArgument, "task_queue is required")
	}

	key := queueKey(defaultNamespace, req.TaskQueue, "workflow")
	raw, err := s.queue.BlockPop(ctx, key, s.cfg.PollTimeoutSeconds)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "poll workflow task queue: %v", err)
	}
	if raw == nil {
		return nil, status.Error(codes.NotFound, "no workflow task available")
	}

	var t domain.WorkflowTask
	if err := json.Unmarshal(raw, &t); err != nil {
		return nil, status.Errorf(codes.Internal, "unmarshal workflow task: %v", err)
	}

	taskToken, _ := json.Marshal(map[string]string{
		"workflow_id": t.WorkflowID,
		"run_id":      t.RunID,
	})
	fmt.Println("===================PollWorkflowTaskQueue", string(taskToken), key)

	return &pb.PollWorkflowTaskQueueResponse{
		TaskToken:    taskToken,
		WorkflowType: t.WorkflowType,
	}, nil
}

func (s *matchingService) PollActivityTaskQueue(ctx context.Context, req *pb.PollActivityTaskQueueRequest) (*pb.PollActivityTaskQueueResponse, error) {
	if req.TaskQueue == "" {
		return nil, status.Error(codes.InvalidArgument, "task_queue is required")
	}

	key := queueKey(defaultNamespace, req.TaskQueue, "activity")
	raw, err := s.queue.BlockPop(ctx, key, s.cfg.PollTimeoutSeconds)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "poll activity task queue: %v", err)
	}
	if raw == nil {
		return nil, status.Error(codes.NotFound, "no activity task available")
	}

	var t domain.ActivityTask
	if err := json.Unmarshal(raw, &t); err != nil {
		return nil, status.Errorf(codes.Internal, "unmarshal activity task: %v", err)
	}

	taskToken, _ := json.Marshal(map[string]string{
		"workflow_id": t.WorkflowID,
		"run_id":      t.RunID,
		"activity_id": t.ActivityID,
	})

	return &pb.PollActivityTaskQueueResponse{
		TaskToken:               taskToken,
		ActivityId:              t.ActivityID,
		ActivityType:            t.ActivityType,
		Input:                   t.Input,
		HeartbeatTimeoutSeconds: t.HeartbeatTimeoutSeconds,
	}, nil
}
