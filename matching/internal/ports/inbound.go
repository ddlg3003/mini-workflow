package ports

import (
	"context"

	pb "mini-workflow/api"
)

type MatchingService interface {
	AddWorkflowTask(ctx context.Context, req *pb.AddWorkflowTaskRequest) (*pb.AddWorkflowTaskResponse, error)
	AddActivityTask(ctx context.Context, req *pb.AddActivityTaskRequest) (*pb.AddActivityTaskResponse, error)
	PollWorkflowTaskQueue(ctx context.Context, req *pb.PollWorkflowTaskQueueRequest) (*pb.PollWorkflowTaskQueueResponse, error)
	PollActivityTaskQueue(ctx context.Context, req *pb.PollActivityTaskQueueRequest) (*pb.PollActivityTaskQueueResponse, error)
}
