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

func TestQueryWorkflow(t *testing.T) {
	tests := []struct {
		name       string
		req        *pb.QueryWorkflowRequest
		setup      func(hist *mocks.MockHistoryClient, match *mocks.MockMatchingClient)
		wantResult []byte
		wantCode   codes.Code
	}{
		{
			name: "success",
			req:  &pb.QueryWorkflowRequest{WorkflowId: "wf1", QueryType: "state"},
			setup: func(hist *mocks.MockHistoryClient, _ *mocks.MockMatchingClient) {
				hist.EXPECT().
					QueryWorkflowExecution(context.Background(), &pb.QueryWorkflowRequest{WorkflowId: "wf1", QueryType: "state"}).
					Return(&pb.QueryWorkflowResponse{QueryResult: []byte(`"running"`)}, nil).
					Once()
			},
			wantResult: []byte(`"running"`),
			wantCode:   codes.OK,
		},
		{
			name: "not found passthrough",
			req:  &pb.QueryWorkflowRequest{WorkflowId: "missing", QueryType: "state"},
			setup: func(hist *mocks.MockHistoryClient, _ *mocks.MockMatchingClient) {
				hist.EXPECT().
					QueryWorkflowExecution(context.Background(), &pb.QueryWorkflowRequest{WorkflowId: "missing", QueryType: "state"}).
					Return(nil, status.Error(codes.NotFound, "not found")).
					Once()
			},
			wantCode: codes.NotFound,
		},
		{
			name: "history unavailable",
			req:  &pb.QueryWorkflowRequest{WorkflowId: "wf1", QueryType: "state"},
			setup: func(hist *mocks.MockHistoryClient, _ *mocks.MockMatchingClient) {
				hist.EXPECT().
					QueryWorkflowExecution(context.Background(), &pb.QueryWorkflowRequest{WorkflowId: "wf1", QueryType: "state"}).
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

			resp, err := svc.QueryWorkflow(context.Background(), tc.req)

			if tc.wantCode == codes.OK {
				require.NoError(t, err)
				assert.Equal(t, tc.wantResult, resp.QueryResult)
			} else {
				s, _ := status.FromError(err)
				assert.Equal(t, tc.wantCode, s.Code())
			}
		})
	}
}
