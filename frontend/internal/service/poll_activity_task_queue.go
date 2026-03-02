package service

import (
	"context"
	"encoding/json"
	"fmt"

	pb "mini-workflow/api"
	"mini-workflow/frontend/internal/token"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *FrontendService) PollActivityTaskQueue(ctx context.Context, req *pb.PollActivityTaskQueueRequest) (*pb.PollActivityTaskQueueResponse, error) {
	fmt.Println("=============Polling Activity Task Queue", req)
	pollResp, err := s.matching.PollActivityTaskQueue(ctx, req)
	if err != nil {
		if isTimeout(err) {
			return &pb.PollActivityTaskQueueResponse{}, nil
		}
		return nil, mapDownstreamError(err)
	}

	fmt.Println("=============Polling Activity Task Queue", pollResp)
	var matchingToken map[string]string
	if err := json.Unmarshal(pollResp.TaskToken, &matchingToken); err != nil {
		return nil, status.Errorf(codes.Internal, "decode matching token: %v", err)
	}

	tok, err := token.Encode(token.TaskToken{
		WorkflowID: matchingToken["workflow_id"],
		RunID:      matchingToken["run_id"],
		ActivityID: pollResp.ActivityId,
		Attempt:    1,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "encode token: %v", err)
	}

	fmt.Println("=============Polling Activity Task Queue", tok)
	return &pb.PollActivityTaskQueueResponse{
		TaskToken:               tok,
		ActivityId:              pollResp.ActivityId,
		ActivityType:            pollResp.ActivityType,
		Input:                   pollResp.Input,
		HeartbeatTimeoutSeconds: pollResp.HeartbeatTimeoutSeconds,
	}, nil
}
