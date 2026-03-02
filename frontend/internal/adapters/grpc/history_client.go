package grpc

import (
	"context"

	pb "mini-workflow/api"
	"mini-workflow/config"
	"mini-workflow/frontend/internal/interceptor"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type historyClient struct {
	client pb.HistoryServiceClient
}

func NewHistoryClient(cfg config.ServiceConfig) (*historyClient, error) {
	conn, err := grpc.NewClient(cfg.Addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(interceptor.Timeout(cfg.ClientTimeout())),
	)
	if err != nil {
		return nil, err
	}
	return &historyClient{client: pb.NewHistoryServiceClient(conn)}, nil
}

func (c *historyClient) RecordWorkflowExecutionStarted(ctx context.Context, req *pb.RecordWorkflowExecutionStartedRequest) (*pb.RecordWorkflowExecutionStartedResponse, error) {
	return c.client.RecordWorkflowExecutionStarted(ctx, req)
}

func (c *historyClient) RecordWorkflowTaskStarted(ctx context.Context, req *pb.RecordWorkflowTaskStartedRequest) (*pb.RecordWorkflowTaskStartedResponse, error) {
	return c.client.RecordWorkflowTaskStarted(ctx, req)
}

func (c *historyClient) RecordWorkflowTaskCompleted(ctx context.Context, req *pb.RecordWorkflowTaskCompletedHistoryRequest) (*pb.RecordWorkflowTaskCompletedHistoryResponse, error) {
	return c.client.RecordWorkflowTaskCompleted(ctx, req)
}

func (c *historyClient) RecordWorkflowTaskFailed(ctx context.Context, req *pb.RecordWorkflowTaskFailedHistoryRequest) (*pb.RecordWorkflowTaskFailedHistoryResponse, error) {
	return c.client.RecordWorkflowTaskFailed(ctx, req)
}

func (c *historyClient) RecordActivityTaskStarted(ctx context.Context, req *pb.RecordActivityTaskStartedRequest) (*pb.RecordActivityTaskStartedResponse, error) {
	return c.client.RecordActivityTaskStarted(ctx, req)
}

func (c *historyClient) RecordActivityTaskCompleted(ctx context.Context, req *pb.RecordActivityTaskCompletedRequest) (*pb.RecordActivityTaskCompletedResponse, error) {
	return c.client.RecordActivityTaskCompleted(ctx, req)
}

func (c *historyClient) RecordActivityTaskFailed(ctx context.Context, req *pb.RecordActivityTaskFailedRequest) (*pb.RecordActivityTaskFailedResponse, error) {
	return c.client.RecordActivityTaskFailed(ctx, req)
}

func (c *historyClient) RecordActivityTaskHeartbeat(ctx context.Context, req *pb.RecordActivityTaskHeartbeatRequest) (*pb.RecordActivityTaskHeartbeatResponse, error) {
	return c.client.RecordActivityTaskHeartbeat(ctx, req)
}

func (c *historyClient) SignalWorkflowExecution(ctx context.Context, req *pb.SignalWorkflowRequest) (*pb.SignalWorkflowResponse, error) {
	return c.client.SignalWorkflowExecution(ctx, req)
}

func (c *historyClient) QueryWorkflowExecution(ctx context.Context, req *pb.QueryWorkflowRequest) (*pb.QueryWorkflowResponse, error) {
	return c.client.QueryWorkflowExecution(ctx, req)
}

func (c *historyClient) GetWorkflowExecutionHistory(ctx context.Context, req *pb.GetHistoryRequest) (*pb.GetHistoryResponse, error) {
	return c.client.GetWorkflowExecutionHistory(ctx, req)
}
