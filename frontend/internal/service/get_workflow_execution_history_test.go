package service

import (
	"context"
	"testing"

	pb "mini-workflow/api"
	"mini-workflow/frontend/internal/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGetWorkflowExecutionHistory(t *testing.T) {
	tests := []struct {
		name     string
		req      *pb.GetHistoryRequest
		setup    func(hist *mocks.MockHistoryClient, match *mocks.MockMatchingClient)
		wantResp *pb.GetHistoryResponse
		wantCode codes.Code
	}{
		{
			name: "success",
			req:  &pb.GetHistoryRequest{WorkflowId: "wf1", RunId: "run-1"},
			setup: func(hist *mocks.MockHistoryClient, _ *mocks.MockMatchingClient) {
				hist.EXPECT().
					GetWorkflowExecutionHistory(context.Background(), &pb.GetHistoryRequest{
						WorkflowId: "wf1",
						RunId:      "run-1",
					}).
					Return(&pb.GetHistoryResponse{
						History: []*pb.HistoryEvent{{EventId: 1}},
					}, nil).
					Once()
			},
			wantResp: &pb.GetHistoryResponse{
				History: []*pb.HistoryEvent{{EventId: 1}},
			},
			wantCode: codes.OK,
		},
		{
			name: "history unavailable",
			req:  &pb.GetHistoryRequest{WorkflowId: "wf1", RunId: "run-1"},
			setup: func(hist *mocks.MockHistoryClient, _ *mocks.MockMatchingClient) {
				hist.EXPECT().
					GetWorkflowExecutionHistory(context.Background(), &pb.GetHistoryRequest{
						WorkflowId: "wf1",
						RunId:      "run-1",
					}).
					Return(nil, status.Error(codes.Unavailable, "down")).
					Once()
			},
			wantCode: codes.Unavailable,
		},
		{
			name: "workflow not found",
			req:  &pb.GetHistoryRequest{WorkflowId: "wf2", RunId: "run-2"},
			setup: func(hist *mocks.MockHistoryClient, _ *mocks.MockMatchingClient) {
				hist.EXPECT().
					GetWorkflowExecutionHistory(context.Background(), &pb.GetHistoryRequest{
						WorkflowId: "wf2",
						RunId:      "run-2",
					}).
					Return(nil, status.Error(codes.NotFound, "not found")).
					Once()
			},
			wantCode: codes.NotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, hist, match := newTestService(t)
			if tc.setup != nil {
				tc.setup(hist, match)
			}

			resp, err := svc.GetWorkflowExecutionHistory(context.Background(), tc.req)

			if tc.wantCode == codes.OK {
				require.NoError(t, err)
				assert.Equal(t, tc.wantResp, resp)
			} else {
				s, _ := status.FromError(err)
				assert.Equal(t, tc.wantCode, s.Code())
			}
		})
	}
}
