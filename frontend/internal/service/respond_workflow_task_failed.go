package service

import (
	"context"

	pb "mini-workflow/api"
	"mini-workflow/frontend/internal/token"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *FrontendService) RespondWorkflowTaskFailed(ctx context.Context, req *pb.RespondWorkflowTaskFailedRequest) (*pb.RespondWorkflowTaskFailedResponse, error) {
	tok, err := token.Decode(req.TaskToken)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid task token: %v", err)
	}

	_, err = s.history(tok.WorkflowID).RecordWorkflowTaskFailed(ctx, &pb.RecordWorkflowTaskFailedHistoryRequest{
		WorkflowId: tok.WorkflowID,
		RunId:      tok.RunID,
		Cause:      req.Cause,
	})
	if err != nil {
		return nil, mapDownstreamError(err)
	}
	return &pb.RespondWorkflowTaskFailedResponse{}, nil
}
