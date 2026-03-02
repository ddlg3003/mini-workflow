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

func TestPollActivityTaskQueue(t *testing.T) {
	tests := []struct {
		name      string
		req       *pb.PollActivityTaskQueueRequest
		setup     func(hist *mocks.MockHistoryClient, match *mocks.MockMatchingClient)
		wantToken token.TaskToken
		wantCode  codes.Code
		wantEmpty bool
	}{
		{
			name: "poll timeout returns empty response",
			req:  &pb.PollActivityTaskQueueRequest{TaskQueue: "q1"},
			setup: func(hist *mocks.MockHistoryClient, match *mocks.MockMatchingClient) {
				match.EXPECT().
					PollActivityTaskQueue(mock.Anything, &pb.PollActivityTaskQueueRequest{TaskQueue: "q1"}).
					Return(nil, status.Error(codes.DeadlineExceeded, "timeout")).
					Once()
			},
			wantEmpty: true,
		},
		{
			name: "matching unavailable",
			req:  &pb.PollActivityTaskQueueRequest{TaskQueue: "q1"},
			setup: func(hist *mocks.MockHistoryClient, match *mocks.MockMatchingClient) {
				match.EXPECT().
					PollActivityTaskQueue(mock.Anything, &pb.PollActivityTaskQueueRequest{TaskQueue: "q1"}).
					Return(nil, status.Error(codes.Unavailable, "down")).
					Once()
			},
			wantCode: codes.Unavailable,
		},
		{
			name: "success with valid token",
			req:  &pb.PollActivityTaskQueueRequest{TaskQueue: "q1"},
			setup: func(hist *mocks.MockHistoryClient, match *mocks.MockMatchingClient) {
				match.EXPECT().
					PollActivityTaskQueue(mock.Anything, &pb.PollActivityTaskQueueRequest{TaskQueue: "q1"}).
					Return(&pb.PollActivityTaskQueueResponse{
						TaskToken:    []byte(`{"workflow_id":"wf-1","run_id":"run-2","activity_id":"act-1"}`),
						ActivityId:   "act-1",
						ActivityType: "MyAct",
						Input:        []byte(`{}`),
					}, nil).
					Once()
			},
			wantToken: token.TaskToken{
				WorkflowID: "wf-1",
				RunID:      "run-2",
				ActivityID: "act-1",
				Attempt:    1,
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

			resp, err := svc.PollActivityTaskQueue(context.Background(), tc.req)

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
