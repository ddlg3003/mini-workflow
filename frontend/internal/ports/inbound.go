package ports

import (
	"context"

	pb "mini-workflow/api"
)

type FrontendService interface {
	StartWorkflow(ctx context.Context, req *pb.StartWorkflowRequest) (*pb.StartWorkflowResponse, error)
	SignalWorkflow(ctx context.Context, req *pb.SignalWorkflowRequest) (*pb.SignalWorkflowResponse, error)
	QueryWorkflow(ctx context.Context, req *pb.QueryWorkflowRequest) (*pb.QueryWorkflowResponse, error)
	GetWorkflowExecutionHistory(ctx context.Context, req *pb.GetHistoryRequest) (*pb.GetHistoryResponse, error)
	PollWorkflowTaskQueue(ctx context.Context, req *pb.PollWorkflowTaskQueueRequest) (*pb.PollWorkflowTaskQueueResponse, error)
	RespondWorkflowTaskCompleted(ctx context.Context, req *pb.RespondWorkflowTaskCompletedRequest) (*pb.RespondWorkflowTaskCompletedResponse, error)
	RespondWorkflowTaskFailed(ctx context.Context, req *pb.RespondWorkflowTaskFailedRequest) (*pb.RespondWorkflowTaskFailedResponse, error)
	PollActivityTaskQueue(ctx context.Context, req *pb.PollActivityTaskQueueRequest) (*pb.PollActivityTaskQueueResponse, error)
	RespondActivityTaskCompleted(ctx context.Context, req *pb.RespondActivityTaskCompletedRequest) (*pb.RespondActivityTaskCompletedResponse, error)
	RespondActivityTaskFailed(ctx context.Context, req *pb.RespondActivityTaskFailedRequest) (*pb.RespondActivityTaskFailedResponse, error)
	RecordActivityTaskHeartbeat(ctx context.Context, req *pb.RecordActivityTaskHeartbeatRequest) (*pb.RecordActivityTaskHeartbeatResponse, error)
}
