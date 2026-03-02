package service

import (
	"context"

	pb "mini-workflow/api"
	"mini-workflow/frontend/internal/token"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *FrontendService) RespondActivityTaskFailed(ctx context.Context, req *pb.RespondActivityTaskFailedRequest) (*pb.RespondActivityTaskFailedResponse, error) {
	tok, err := token.Decode(req.TaskToken)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid task token: %v", err)
	}

	_, err = s.history(tok.WorkflowID).RecordActivityTaskFailed(ctx, &pb.RecordActivityTaskFailedRequest{
		WorkflowId: tok.WorkflowID,
		RunId:      tok.RunID,
		ActivityId: tok.ActivityID,
		Reason:     req.Reason,
	})
	if err != nil {
		return nil, mapDownstreamError(err)
	}
	return &pb.RespondActivityTaskFailedResponse{}, nil
}
