package grpc

import (
	"context"

	pb "mini-workflow/api"
	"mini-workflow/config"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type matchingClient struct {
	client pb.MatchingServiceClient
}

func NewMatchingClient(cfg config.ServiceConfig) (*matchingClient, error) {
	conn, err := grpc.NewClient(cfg.Addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
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
