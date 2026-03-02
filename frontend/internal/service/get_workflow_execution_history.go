package service

import (
	"context"

	pb "mini-workflow/api"
)

func (s *FrontendService) GetWorkflowExecutionHistory(ctx context.Context, req *pb.GetHistoryRequest) (*pb.GetHistoryResponse, error) {
	resp, err := s.history(req.WorkflowId).GetWorkflowExecutionHistory(ctx, req)
	if err != nil {
		return nil, mapDownstreamError(err)
	}
	return resp, nil
}
