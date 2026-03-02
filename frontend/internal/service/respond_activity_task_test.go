package service

import (
	"context"
	"testing"

	pb "mini-workflow/api"
	"mini-workflow/frontend/internal/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRespondActivityTaskCompleted(t *testing.T) {
	validTok := func(t *testing.T) []byte { return makeToken(t, "wf1", "run-1", "act-1") }

	tests := []struct {
		name     string
		req      func(t *testing.T) *pb.RespondActivityTaskCompletedRequest
		setup    func(hist *mocks.MockHistoryClient, match *mocks.MockMatchingClient)
		wantCode codes.Code
	}{
		{
			name: "success",
			req: func(t *testing.T) *pb.RespondActivityTaskCompletedRequest {
				return &pb.RespondActivityTaskCompletedRequest{TaskToken: validTok(t), Result: []byte(`"ok"`)}
			},
			setup: func(hist *mocks.MockHistoryClient, _ *mocks.MockMatchingClient) {
				hist.EXPECT().
					RecordActivityTaskCompleted(context.Background(), &pb.RecordActivityTaskCompletedRequest{
						WorkflowId: "wf1", RunId: "run-1", ActivityId: "act-1", Result: []byte(`"ok"`),
					}).
					Return(&pb.RecordActivityTaskCompletedResponse{}, nil).
					Once()
			},
			wantCode: codes.OK,
		},
		{
			name: "bad token",
			req: func(_ *testing.T) *pb.RespondActivityTaskCompletedRequest {
				return &pb.RespondActivityTaskCompletedRequest{TaskToken: []byte("bad")}
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
			resp, err := svc.RespondActivityTaskCompleted(context.Background(), tc.req(t))
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

func TestRespondActivityTaskFailed(t *testing.T) {
	validTok := func(t *testing.T) []byte { return makeToken(t, "wf1", "run-1", "act-1") }

	tests := []struct {
		name     string
		req      func(t *testing.T) *pb.RespondActivityTaskFailedRequest
		setup    func(hist *mocks.MockHistoryClient, match *mocks.MockMatchingClient)
		wantCode codes.Code
	}{
		{
			name: "success",
			req: func(t *testing.T) *pb.RespondActivityTaskFailedRequest {
				return &pb.RespondActivityTaskFailedRequest{TaskToken: validTok(t), Reason: "timeout"}
			},
			setup: func(hist *mocks.MockHistoryClient, _ *mocks.MockMatchingClient) {
				hist.EXPECT().
					RecordActivityTaskFailed(context.Background(), &pb.RecordActivityTaskFailedRequest{
						WorkflowId: "wf1", RunId: "run-1", ActivityId: "act-1", Reason: "timeout",
					}).
					Return(&pb.RecordActivityTaskFailedResponse{}, nil).
					Once()
			},
			wantCode: codes.OK,
		},
		{
			name: "bad token",
			req: func(_ *testing.T) *pb.RespondActivityTaskFailedRequest {
				return &pb.RespondActivityTaskFailedRequest{TaskToken: []byte("!!!")}
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
			resp, err := svc.RespondActivityTaskFailed(context.Background(), tc.req(t))
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

func TestRecordActivityTaskHeartbeat(t *testing.T) {
	validTok := func(t *testing.T) []byte { return makeToken(t, "wf1", "run-1", "act-1") }

	tests := []struct {
		name          string
		req           func(t *testing.T) *pb.RecordActivityTaskHeartbeatRequest
		setup         func(hist *mocks.MockHistoryClient, match *mocks.MockMatchingClient)
		wantCancelReq bool
		wantCode      codes.Code
	}{
		{
			name: "success — no cancellation",
			req: func(t *testing.T) *pb.RecordActivityTaskHeartbeatRequest {
				return &pb.RecordActivityTaskHeartbeatRequest{TaskToken: validTok(t), Details: []byte(`"50%"`)}
			},
			setup: func(hist *mocks.MockHistoryClient, _ *mocks.MockMatchingClient) {
				hist.EXPECT().
					RecordActivityTaskHeartbeat(context.Background(), mock.AnythingOfType("*workflow.RecordActivityTaskHeartbeatRequest")).
					Return(&pb.RecordActivityTaskHeartbeatResponse{CancelRequested: false}, nil).
					Once()
			},
			wantCode: codes.OK,
		},
		{
			name: "bad token",
			req: func(_ *testing.T) *pb.RecordActivityTaskHeartbeatRequest {
				return &pb.RecordActivityTaskHeartbeatRequest{TaskToken: []byte("!!!")}
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
			resp, err := svc.RecordActivityTaskHeartbeat(context.Background(), tc.req(t))
			if tc.wantCode == codes.OK {
				require.NoError(t, err)
				assert.Equal(t, tc.wantCancelReq, resp.CancelRequested)
			} else {
				s, _ := status.FromError(err)
				assert.Equal(t, tc.wantCode, s.Code())
			}
		})
	}
}
