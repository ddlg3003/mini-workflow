package grpc

import (
	"context"

	pb "mini-workflow/api"
	"mini-workflow/config"
	"mini-workflow/frontend/internal/interceptor"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type matchingClient struct {
	client pb.MatchingServiceClient
}

func NewMatchingClient(cfg config.ServiceConfig) (*matchingClient, error) {
	conn, err := grpc.NewClient(cfg.Addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(interceptor.MatchingTimeout(cfg.ClientTimeout(), cfg.PollTimeout())),
	)
	if err != nil {
		return nil, err
	}
	return &matchingClient{client: pb.NewMatchingServiceClient(conn)}, nil
}

func (c *matchingClient) AddWorkflowTask(ctx context.Context, req *pb.AddWorkflowTaskRequest) (*pb.AddWorkflowTaskResponse, error) {
	return c.client.AddWorkflowTask(ctx, req)
}

func (c *matchingClient) AddActivityTask(ctx context.Context, req *pb.AddActivityTaskRequest) (*pb.AddActivityTaskResponse, error) {
	return c.client.AddActivityTask(ctx, req)
}

func (c *matchingClient) PollWorkflowTaskQueue(ctx context.Context, req *pb.PollWorkflowTaskQueueRequest) (*pb.PollWorkflowTaskQueueResponse, error) {
	return c.client.PollWorkflowTaskQueue(ctx, req)
}

func (c *matchingClient) PollActivityTaskQueue(ctx context.Context, req *pb.PollActivityTaskQueueRequest) (*pb.PollActivityTaskQueueResponse, error) {
	return c.client.PollActivityTaskQueue(ctx, req)
}
