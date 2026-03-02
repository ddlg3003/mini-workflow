package service

import (
	"context"
	"testing"

	pb "mini-workflow/api"
	"mini-workflow/frontend/internal/mocks"
	"mini-workflow/frontend/internal/token"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestPollWorkflowTaskQueue(t *testing.T) {
	tests := []struct {
		name      string
		req       *pb.PollWorkflowTaskQueueRequest
		setup     func(hist *mocks.MockHistoryClient, match *mocks.MockMatchingClient)
		wantToken token.TaskToken
		wantCode  codes.Code
		wantEmpty bool
	}{
		{
			name: "poll timeout returns empty response",
			req:  &pb.PollWorkflowTaskQueueRequest{TaskQueue: "q1"},
			setup: func(hist *mocks.MockHistoryClient, match *mocks.MockMatchingClient) {
				match.EXPECT().
					PollWorkflowTaskQueue(mock.Anything, &pb.PollWorkflowTaskQueueRequest{TaskQueue: "q1"}).
					Return(nil, status.Error(codes.DeadlineExceeded, "timeout")).
					Once()
			},
			wantEmpty: true,
		},
		{
			name: "matching unavailable",
			req:  &pb.PollWorkflowTaskQueueRequest{TaskQueue: "q1"},
			setup: func(hist *mocks.MockHistoryClient, match *mocks.MockMatchingClient) {
				match.EXPECT().
					PollWorkflowTaskQueue(mock.Anything, &pb.PollWorkflowTaskQueueRequest{TaskQueue: "q1"}).
					Return(nil, status.Error(codes.Unavailable, "down")).
					Once()
			},
			wantCode: codes.Unavailable,
		},
		{
			name: "success with valid token",
			req:  &pb.PollWorkflowTaskQueueRequest{TaskQueue: "q1"},
			setup: func(hist *mocks.MockHistoryClient, match *mocks.MockMatchingClient) {
				match.EXPECT().
					PollWorkflowTaskQueue(mock.Anything, &pb.PollWorkflowTaskQueueRequest{TaskQueue: "q1"}).
					Return(&pb.PollWorkflowTaskQueueResponse{
						TaskToken:    []byte(`{"workflow_id":"wf1","run_id":"run-1"}`),
						WorkflowType: "wf1",
					}, nil).
					Once()
				hist.EXPECT().
					GetWorkflowExecutionHistory(mock.Anything, &pb.GetHistoryRequest{
						WorkflowId: "wf1",
						RunId:      "run-1",
					}).
					Return(&pb.GetHistoryResponse{
						History: []*pb.HistoryEvent{{EventId: 1}},
					}, nil).
					Once()
			},
			wantToken: token.TaskToken{
				WorkflowID: "wf1",
				RunID:      "run-1",
			},
			wantCode: codes.OK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, hist, match := newTestService(t)
			if tc.setup != nil {
				tc.setup(hist, match)
			}

			resp, err := svc.PollWorkflowTaskQueue(context.Background(), tc.req)

			if tc.wantCode == codes.OK {
				require.NoError(t, err)
				if tc.wantEmpty {
					assert.Empty(t, resp.TaskToken)
				} else {
					decodedArgs, err := token.Decode(resp.TaskToken)
					require.NoError(t, err)
					assert.Equal(t, tc.wantToken, decodedArgs)
				}
			} else {
				s, _ := status.FromError(err)
				assert.Equal(t, tc.wantCode, s.Code())
			}
		})
	}
}
