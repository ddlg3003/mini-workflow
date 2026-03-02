package grpc

import (
	"context"

	pb "mini-workflow/api"
	"mini-workflow/frontend/internal/ports"
	"mini-workflow/frontend/internal/service"
)

type Handler struct {
	pb.UnimplementedFrontendServiceServer
	svc *service.FrontendService
}

func NewHandler(svc *service.FrontendService) *Handler {
	return &Handler{svc: svc}
}

var _ ports.FrontendService = (*service.FrontendService)(nil)

func (h *Handler) StartWorkflow(ctx context.Context, req *pb.StartWorkflowRequest) (*pb.StartWorkflowResponse, error) {
	return h.svc.StartWorkflow(ctx, req)
}

func (h *Handler) SignalWorkflow(ctx context.Context, req *pb.SignalWorkflowRequest) (*pb.SignalWorkflowResponse, error) {
	return h.svc.SignalWorkflow(ctx, req)
}

func (h *Handler) QueryWorkflow(ctx context.Context, req *pb.QueryWorkflowRequest) (*pb.QueryWorkflowResponse, error) {
	return h.svc.QueryWorkflow(ctx, req)
}

func (h *Handler) GetWorkflowExecutionHistory(ctx context.Context, req *pb.GetHistoryRequest) (*pb.GetHistoryResponse, error) {
	return h.svc.GetWorkflowExecutionHistory(ctx, req)
}

func (h *Handler) PollWorkflowTaskQueue(ctx context.Context, req *pb.PollWorkflowTaskQueueRequest) (*pb.PollWorkflowTaskQueueResponse, error) {
	return h.svc.PollWorkflowTaskQueue(ctx, req)
}

func (h *Handler) RespondWorkflowTaskCompleted(ctx context.Context, req *pb.RespondWorkflowTaskCompletedRequest) (*pb.RespondWorkflowTaskCompletedResponse, error) {
	return h.svc.RespondWorkflowTaskCompleted(ctx, req)
}

func (h *Handler) RespondWorkflowTaskFailed(ctx context.Context, req *pb.RespondWorkflowTaskFailedRequest) (*pb.RespondWorkflowTaskFailedResponse, error) {
	return h.svc.RespondWorkflowTaskFailed(ctx, req)
}

func (h *Handler) PollActivityTaskQueue(ctx context.Context, req *pb.PollActivityTaskQueueRequest) (*pb.PollActivityTaskQueueResponse, error) {
	return h.svc.PollActivityTaskQueue(ctx, req)
}

func (h *Handler) RespondActivityTaskCompleted(ctx context.Context, req *pb.RespondActivityTaskCompletedRequest) (*pb.RespondActivityTaskCompletedResponse, error) {
	return h.svc.RespondActivityTaskCompleted(ctx, req)
}

func (h *Handler) RespondActivityTaskFailed(ctx context.Context, req *pb.RespondActivityTaskFailedRequest) (*pb.RespondActivityTaskFailedResponse, error) {
	return h.svc.RespondActivityTaskFailed(ctx, req)
}

func (h *Handler) RecordActivityTaskHeartbeat(ctx context.Context, req *pb.RecordActivityTaskHeartbeatRequest) (*pb.RecordActivityTaskHeartbeatResponse, error) {
	return h.svc.RecordActivityTaskHeartbeat(ctx, req)
}
