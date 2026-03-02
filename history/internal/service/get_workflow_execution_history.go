package service

import (
	"context"

	pb "mini-workflow/api"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *historyService) GetWorkflowExecutionHistory(ctx context.Context, req *pb.GetHistoryRequest) (*pb.GetHistoryResponse, error) {
	runID, err := runIDFromString(req.RunId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	events, err := s.repo.GetHistoryEvents(ctx, "default", req.WorkflowId, runID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get history events: %v", err)
	}

	protoEvents := make([]*pb.HistoryEvent, 0, len(events))
	for _, e := range events {
		protoEvents = append(protoEvents, &pb.HistoryEvent{
			EventId:   e.EventID,
			EventType: e.EventType,
			Payload:   e.Payload,
			Timestamp: e.CreatedAt.UnixMilli(),
		})
	}

	return &pb.GetHistoryResponse{History: protoEvents}, nil
}
