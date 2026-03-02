package service

import (
	"context"
	"encoding/json"

	pb "mini-workflow/api"
	"mini-workflow/frontend/internal/token"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *FrontendService) PollWorkflowTaskQueue(ctx context.Context, req *pb.PollWorkflowTaskQueueRequest) (*pb.PollWorkflowTaskQueueResponse, error) {
	pollResp, err := s.matching.PollWorkflowTaskQueue(ctx, req)
	if err != nil {
		if isTimeout(err) {
			return &pb.PollWorkflowTaskQueueResponse{}, nil
		}
		return nil, mapDownstreamError(err)
	}

	var matchingToken map[string]string
	if err := json.Unmarshal(pollResp.TaskToken, &matchingToken); err != nil {
		return nil, status.Errorf(codes.Internal, "decode matching token: %v", err)
	}

	histResp, err := s.history(matchingToken["workflow_id"]).GetWorkflowExecutionHistory(ctx, &pb.GetHistoryRequest{
		WorkflowId: matchingToken["workflow_id"],
		RunId:      matchingToken["run_id"],
	})

	if err != nil {
		return nil, mapDownstreamError(err)
	}

	tok, err := token.Encode(token.TaskToken{
		WorkflowID: matchingToken["workflow_id"],
		RunID:      matchingToken["run_id"],
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "encode token: %v", err)
	}

	return &pb.PollWorkflowTaskQueueResponse{
		TaskToken:    tok,
		WorkflowType: pollResp.WorkflowType,
		History:      histResp.History,
	}, nil
}
