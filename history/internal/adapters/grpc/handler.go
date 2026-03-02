package grpc

import (
	"context"

	pb "mini-workflow/api"
	"mini-workflow/history/internal/ports"
)

type Handler struct {
	pb.UnimplementedHistoryServiceServer
	svc ports.HistoryService
}

func NewHandler(svc ports.HistoryService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RecordWorkflowExecutionStarted(ctx context.Context, req *pb.RecordWorkflowExecutionStartedRequest) (*pb.RecordWorkflowExecutionStartedResponse, error) {
	return h.svc.RecordWorkflowExecutionStarted(ctx, req)
}

func (h *Handler) RecordWorkflowTaskStarted(ctx context.Context, req *pb.RecordWorkflowTaskStartedRequest) (*pb.RecordWorkflowTaskStartedResponse, error) {
	return h.svc.RecordWorkflowTaskStarted(ctx, req)
}

func (h *Handler) RecordWorkflowTaskCompleted(ctx context.Context, req *pb.RecordWorkflowTaskCompletedHistoryRequest) (*pb.RecordWorkflowTaskCompletedHistoryResponse, error) {
	return h.svc.RecordWorkflowTaskCompleted(ctx, req)
}

func (h *Handler) RecordWorkflowTaskFailed(ctx context.Context, req *pb.RecordWorkflowTaskFailedHistoryRequest) (*pb.RecordWorkflowTaskFailedHistoryResponse, error) {
	return h.svc.RecordWorkflowTaskFailed(ctx, req)
}

func (h *Handler) RecordActivityTaskStarted(ctx context.Context, req *pb.RecordActivityTaskStartedRequest) (*pb.RecordActivityTaskStartedResponse, error) {
	return h.svc.RecordActivityTaskStarted(ctx, req)
}

func (h *Handler) RecordActivityTaskCompleted(ctx context.Context, req *pb.RecordActivityTaskCompletedRequest) (*pb.RecordActivityTaskCompletedResponse, error) {
	return h.svc.RecordActivityTaskCompleted(ctx, req)
}

func (h *Handler) RecordActivityTaskFailed(ctx context.Context, req *pb.RecordActivityTaskFailedRequest) (*pb.RecordActivityTaskFailedResponse, error) {
	return h.svc.RecordActivityTaskFailed(ctx, req)
}

func (h *Handler) RecordActivityTaskHeartbeat(ctx context.Context, req *pb.RecordActivityTaskHeartbeatRequest) (*pb.RecordActivityTaskHeartbeatResponse, error) {
	return h.svc.RecordActivityTaskHeartbeat(ctx, req)
}

func (h *Handler) SignalWorkflowExecution(ctx context.Context, req *pb.SignalWorkflowRequest) (*pb.SignalWorkflowResponse, error) {
	return h.svc.SignalWorkflowExecution(ctx, req)
}

func (h *Handler) QueryWorkflowExecution(ctx context.Context, req *pb.QueryWorkflowRequest) (*pb.QueryWorkflowResponse, error) {
	return h.svc.QueryWorkflowExecution(ctx, req)
}

func (h *Handler) GetWorkflowExecutionHistory(ctx context.Context, req *pb.GetHistoryRequest) (*pb.GetHistoryResponse, error) {
	return h.svc.GetWorkflowExecutionHistory(ctx, req)
}
