package service

import (
	"context"

	pb "mini-workflow/api"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *FrontendService) StartWorkflow(ctx context.Context, req *pb.StartWorkflowRequest) (*pb.StartWorkflowResponse, error) {
	if req.WorkflowId == "" || req.WorkflowType == "" {
		return nil, status.Error(codes.InvalidArgument, "workflow_id and workflow_type are required")
	}

	histResp, err := s.history(req.WorkflowId).RecordWorkflowExecutionStarted(ctx, &pb.RecordWorkflowExecutionStartedRequest{
		StartRequest: req,
	})
	if err != nil {
		return nil, mapDownstreamError(err)
	}

	return &pb.StartWorkflowResponse{RunId: histResp.RunId}, nil
}
