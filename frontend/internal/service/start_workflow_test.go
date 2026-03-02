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

func TestStartWorkflow(t *testing.T) {
	tests := []struct {
		name      string
		req       *pb.StartWorkflowRequest
		setup     func(hist *mocks.MockHistoryClient, match *mocks.MockMatchingClient)
		wantRunID string
		wantCode  codes.Code
	}{
		{
			name:     "missing workflow_id",
			req:      &pb.StartWorkflowRequest{WorkflowType: "MyWF", TaskQueue: "q"},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "missing workflow_type",
			req:      &pb.StartWorkflowRequest{WorkflowId: "wf1", TaskQueue: "q"},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "history unavailable",
			req:  &pb.StartWorkflowRequest{WorkflowId: "wf1", WorkflowType: "MyWF", TaskQueue: "q"},
			setup: func(hist *mocks.MockHistoryClient, match *mocks.MockMatchingClient) {
				hist.EXPECT().
					RecordWorkflowExecutionStarted(context.Background(), &pb.RecordWorkflowExecutionStartedRequest{
						StartRequest: &pb.StartWorkflowRequest{WorkflowId: "wf1", WorkflowType: "MyWF", TaskQueue: "q"},
					}).
					Return(nil, status.Error(codes.Unavailable, "down")).
					Once()
			},
			wantCode: codes.Unavailable,
		},

		{
			name: "success",
			req:  &pb.StartWorkflowRequest{WorkflowId: "wf1", WorkflowType: "MyWF", TaskQueue: "q"},
			setup: func(hist *mocks.MockHistoryClient, match *mocks.MockMatchingClient) {
				hist.EXPECT().
					RecordWorkflowExecutionStarted(context.Background(), &pb.RecordWorkflowExecutionStartedRequest{
						StartRequest: &pb.StartWorkflowRequest{WorkflowId: "wf1", WorkflowType: "MyWF", TaskQueue: "q"},
					}).
					Return(&pb.RecordWorkflowExecutionStartedResponse{RunId: "run-1"}, nil).
					Once()
			},
			wantRunID: "run-1",
			wantCode:  codes.OK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, hist, match := newTestService(t)
			if tc.setup != nil {
				tc.setup(hist, match)
			}

			resp, err := svc.StartWorkflow(context.Background(), tc.req)

			if tc.wantCode == codes.OK {
				require.NoError(t, err)
				assert.Equal(t, tc.wantRunID, resp.RunId)
			} else {
				s, _ := status.FromError(err)
				assert.Equal(t, tc.wantCode, s.Code())
			}
		})
	}
}
