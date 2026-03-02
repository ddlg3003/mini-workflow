package service

import (
	"context"

	pb "mini-workflow/api"
)

func (s *FrontendService) QueryWorkflow(ctx context.Context, req *pb.QueryWorkflowRequest) (*pb.QueryWorkflowResponse, error) {
	resp, err := s.history(req.WorkflowId).QueryWorkflowExecution(ctx, req)
	if err != nil {
		return nil, mapDownstreamError(err)
	}
	return resp, nil
}
