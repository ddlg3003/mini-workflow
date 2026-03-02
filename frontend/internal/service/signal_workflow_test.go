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

func TestSignalWorkflow(t *testing.T) {
	tests := []struct {
		name     string
		req      *pb.SignalWorkflowRequest
		setup    func(hist *mocks.MockHistoryClient, match *mocks.MockMatchingClient)
		wantCode codes.Code
	}{
		{
			name: "success",
			req:  &pb.SignalWorkflowRequest{WorkflowId: "wf1", SignalName: "cancel"},
			setup: func(hist *mocks.MockHistoryClient, _ *mocks.MockMatchingClient) {
				hist.EXPECT().
					SignalWorkflowExecution(context.Background(), &pb.SignalWorkflowRequest{WorkflowId: "wf1", SignalName: "cancel"}).
					Return(&pb.SignalWorkflowResponse{}, nil).
					Once()
			},
			wantCode: codes.OK,
		},
		{
			name: "history unavailable",
			req:  &pb.SignalWorkflowRequest{WorkflowId: "wf1", SignalName: "cancel"},
			setup: func(hist *mocks.MockHistoryClient, _ *mocks.MockMatchingClient) {
				hist.EXPECT().
					SignalWorkflowExecution(context.Background(), &pb.SignalWorkflowRequest{WorkflowId: "wf1", SignalName: "cancel"}).
					Return(nil, status.Error(codes.Unavailable, "down")).
					Once()
			},
			wantCode: codes.Unavailable,
		},
		{
			name: "workflow not found",
			req:  &pb.SignalWorkflowRequest{WorkflowId: "missing", SignalName: "cancel"},
			setup: func(hist *mocks.MockHistoryClient, _ *mocks.MockMatchingClient) {
				hist.EXPECT().
					SignalWorkflowExecution(context.Background(), &pb.SignalWorkflowRequest{WorkflowId: "missing", SignalName: "cancel"}).
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

			resp, err := svc.SignalWorkflow(context.Background(), tc.req)

			if tc.wantCode == codes.OK {
				require.NoError(t, err)
				assert.NotNil(t, resp)
			} else {
				s, _ := status.FromError(err)
				assert.Equal(t, tc.wantCode, s.Code())
			}
		})
	}
}
