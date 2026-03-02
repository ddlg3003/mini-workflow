package grpc

import (
	"context"

	pb "mini-workflow/api"
	"mini-workflow/matching/internal/ports"
)

type Handler struct {
	pb.UnimplementedMatchingServiceServer
	svc ports.MatchingService
}

func NewHandler(svc ports.MatchingService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) AddWorkflowTask(ctx context.Context, req *pb.AddWorkflowTaskRequest) (*pb.AddWorkflowTaskResponse, error) {
	return h.svc.AddWorkflowTask(ctx, req)
}

func (h *Handler) AddActivityTask(ctx context.Context, req *pb.AddActivityTaskRequest) (*pb.AddActivityTaskResponse, error) {
	return h.svc.AddActivityTask(ctx, req)
}

func (h *Handler) PollWorkflowTaskQueue(ctx context.Context, req *pb.PollWorkflowTaskQueueRequest) (*pb.PollWorkflowTaskQueueResponse, error) {
	return h.svc.PollWorkflowTaskQueue(ctx, req)
}

func (h *Handler) PollActivityTaskQueue(ctx context.Context, req *pb.PollActivityTaskQueueRequest) (*pb.PollActivityTaskQueueResponse, error) {
	return h.svc.PollActivityTaskQueue(ctx, req)
}
