package service

import (
	"context"

	pb "mini-workflow/api"
	"mini-workflow/frontend/internal/token"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *FrontendService) RespondWorkflowTaskCompleted(ctx context.Context, req *pb.RespondWorkflowTaskCompletedRequest) (*pb.RespondWorkflowTaskCompletedResponse, error) {
	tok, err := token.Decode(req.TaskToken)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid task token: %v", err)
	}

	_, err = s.history(tok.WorkflowID).RecordWorkflowTaskCompleted(ctx, &pb.RecordWorkflowTaskCompletedHistoryRequest{
		WorkflowId: tok.WorkflowID,
		RunId:      tok.RunID,
		Commands:   req.Commands,
	})
	if err != nil {
		return nil, mapDownstreamError(err)
	}
	return &pb.RespondWorkflowTaskCompletedResponse{}, nil
}
