package ports

import (
	"context"

	pb "mini-workflow/api"
)

type HistoryService interface {
	RecordWorkflowExecutionStarted(ctx context.Context, req *pb.RecordWorkflowExecutionStartedRequest) (*pb.RecordWorkflowExecutionStartedResponse, error)
	RecordWorkflowTaskStarted(ctx context.Context, req *pb.RecordWorkflowTaskStartedRequest) (*pb.RecordWorkflowTaskStartedResponse, error)
	RecordWorkflowTaskCompleted(ctx context.Context, req *pb.RecordWorkflowTaskCompletedHistoryRequest) (*pb.RecordWorkflowTaskCompletedHistoryResponse, error)
	RecordWorkflowTaskFailed(ctx context.Context, req *pb.RecordWorkflowTaskFailedHistoryRequest) (*pb.RecordWorkflowTaskFailedHistoryResponse, error)
	RecordActivityTaskStarted(ctx context.Context, req *pb.RecordActivityTaskStartedRequest) (*pb.RecordActivityTaskStartedResponse, error)
	RecordActivityTaskCompleted(ctx context.Context, req *pb.RecordActivityTaskCompletedRequest) (*pb.RecordActivityTaskCompletedResponse, error)
	RecordActivityTaskFailed(ctx context.Context, req *pb.RecordActivityTaskFailedRequest) (*pb.RecordActivityTaskFailedResponse, error)
	RecordActivityTaskHeartbeat(ctx context.Context, req *pb.RecordActivityTaskHeartbeatRequest) (*pb.RecordActivityTaskHeartbeatResponse, error)
	SignalWorkflowExecution(ctx context.Context, req *pb.SignalWorkflowRequest) (*pb.SignalWorkflowResponse, error)
	QueryWorkflowExecution(ctx context.Context, req *pb.QueryWorkflowRequest) (*pb.QueryWorkflowResponse, error)
	GetWorkflowExecutionHistory(ctx context.Context, req *pb.GetHistoryRequest) (*pb.GetHistoryResponse, error)
}

type MatchingClient interface {
	AddWorkflowTask(ctx context.Context, req *pb.AddWorkflowTaskRequest) (*pb.AddWorkflowTaskResponse, error)
	AddActivityTask(ctx context.Context, req *pb.AddActivityTaskRequest) (*pb.AddActivityTaskResponse, error)
}
