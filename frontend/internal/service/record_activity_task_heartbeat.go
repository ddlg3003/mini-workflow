package service

import (
	"context"

	pb "mini-workflow/api"
	"mini-workflow/frontend/internal/token"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *FrontendService) RecordActivityTaskHeartbeat(ctx context.Context, req *pb.RecordActivityTaskHeartbeatRequest) (*pb.RecordActivityTaskHeartbeatResponse, error) {
	tok, err := token.Decode(req.TaskToken)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid task token: %v", err)
	}

	resp, err := s.history(tok.WorkflowID).RecordActivityTaskHeartbeat(ctx, req)
	if err != nil {
		return nil, mapDownstreamError(err)
	}
	return resp, nil
}
