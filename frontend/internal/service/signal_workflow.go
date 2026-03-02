package service

import (
	"context"

	pb "mini-workflow/api"
)

func (s *FrontendService) SignalWorkflow(ctx context.Context, req *pb.SignalWorkflowRequest) (*pb.SignalWorkflowResponse, error) {
	resp, err := s.history(req.WorkflowId).SignalWorkflowExecution(ctx, req)
	if err != nil {
		return nil, mapDownstreamError(err)
	}
	return resp, nil
}
