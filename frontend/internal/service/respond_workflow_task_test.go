package service

import (
	"context"
	"testing"

	pb "mini-workflow/api"
	"mini-workflow/frontend/internal/mocks"
	"mini-workflow/frontend/internal/token"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func makeToken(t *testing.T, wfID, runID, actID string) []byte {
	t.Helper()
	tok, err := token.Encode(token.TaskToken{WorkflowID: wfID, RunID: runID, ActivityID: actID})
	require.NoError(t, err)
	return tok
}

func TestRespondWorkflowTaskCompleted(t *testing.T) {
	validTok := func(t *testing.T) []byte { return makeToken(t, "wf1", "run-1", "") }

	tests := []struct {
		name     string
		req      func(t *testing.T) *pb.RespondWorkflowTaskCompletedRequest
		setup    func(hist *mocks.MockHistoryClient, match *mocks.MockMatchingClient)
		wantCode codes.Code
	}{
		{
			name: "success",
			req: func(t *testing.T) *pb.RespondWorkflowTaskCompletedRequest {
				return &pb.RespondWorkflowTaskCompletedRequest{
					TaskToken: validTok(t),
					Commands:  []*pb.Command{{CommandType: "Noop"}},
				}
			},
			setup: func(hist *mocks.MockHistoryClient, _ *mocks.MockMatchingClient) {
				hist.EXPECT().
					RecordWorkflowTaskCompleted(context.Background(), &pb.RecordWorkflowTaskCompletedHistoryRequest{
						WorkflowId: "wf1", RunId: "run-1",
						Commands: []*pb.Command{{CommandType: "Noop"}},
					}).
					Return(&pb.RecordWorkflowTaskCompletedHistoryResponse{}, nil).
					Once()
			},
			wantCode: codes.OK,
		},
		{
			name: "bad token",
			req: func(_ *testing.T) *pb.RespondWorkflowTaskCompletedRequest {
				return &pb.RespondWorkflowTaskCompletedRequest{TaskToken: []byte("bad!!!")}
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "history unavailable",
			req: func(t *testing.T) *pb.RespondWorkflowTaskCompletedRequest {
				return &pb.RespondWorkflowTaskCompletedRequest{TaskToken: validTok(t)}
			},
			setup: func(hist *mocks.MockHistoryClient, _ *mocks.MockMatchingClient) {
				hist.EXPECT().
					RecordWorkflowTaskCompleted(context.Background(), &pb.RecordWorkflowTaskCompletedHistoryRequest{
						WorkflowId: "wf1", RunId: "run-1",
					}).
					Return(nil, status.Error(codes.Unavailable, "down")).
					Once()
			},
			wantCode: codes.Unavailable,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, hist, match := newTestService(t)
			if tc.setup != nil {
				tc.setup(hist, match)
			}

			resp, err := svc.RespondWorkflowTaskCompleted(context.Background(), tc.req(t))

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

func TestRespondWorkflowTaskFailed(t *testing.T) {
	validTok := func(t *testing.T) []byte { return makeToken(t, "wf1", "run-1", "") }

	tests := []struct {
		name     string
		req      func(t *testing.T) *pb.RespondWorkflowTaskFailedRequest
		setup    func(hist *mocks.MockHistoryClient, match *mocks.MockMatchingClient)
		wantCode codes.Code
	}{
		{
			name: "success",
			req: func(t *testing.T) *pb.RespondWorkflowTaskFailedRequest {
				return &pb.RespondWorkflowTaskFailedRequest{TaskToken: validTok(t), Cause: "non-determinism"}
			},
			setup: func(hist *mocks.MockHistoryClient, _ *mocks.MockMatchingClient) {
				hist.EXPECT().
					RecordWorkflowTaskFailed(context.Background(), &pb.RecordWorkflowTaskFailedHistoryRequest{
						WorkflowId: "wf1", RunId: "run-1", Cause: "non-determinism",
					}).
					Return(&pb.RecordWorkflowTaskFailedHistoryResponse{}, nil).
					Once()
			},
			wantCode: codes.OK,
		},
		{
			name: "bad token",
			req: func(_ *testing.T) *pb.RespondWorkflowTaskFailedRequest {
				return &pb.RespondWorkflowTaskFailedRequest{TaskToken: []byte("!!!")}
			},
			wantCode: codes.InvalidArgument,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, hist, match := newTestService(t)
			if tc.setup != nil {
				tc.setup(hist, match)
			}

			resp, err := svc.RespondWorkflowTaskFailed(context.Background(), tc.req(t))

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
