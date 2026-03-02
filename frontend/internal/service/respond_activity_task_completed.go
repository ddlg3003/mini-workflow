package service

import (
	"context"

	pb "mini-workflow/api"
	"mini-workflow/frontend/internal/token"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *FrontendService) RespondActivityTaskCompleted(ctx context.Context, req *pb.RespondActivityTaskCompletedRequest) (*pb.RespondActivityTaskCompletedResponse, error) {
	tok, err := token.Decode(req.TaskToken)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid task token: %v", err)
	}

	_, err = s.history(tok.WorkflowID).RecordActivityTaskCompleted(ctx, &pb.RecordActivityTaskCompletedRequest{
		WorkflowId: tok.WorkflowID,
		RunId:      tok.RunID,
		ActivityId: tok.ActivityID,
		Result:     req.Result,
	})
	if err != nil {
		return nil, mapDownstreamError(err)
	}
	return &pb.RespondActivityTaskCompletedResponse{}, nil
}
