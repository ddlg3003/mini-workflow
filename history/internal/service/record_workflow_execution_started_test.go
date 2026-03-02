package service

import (
	"context"
	"testing"

	pb "mini-workflow/api"
	"mini-workflow/history/internal/domain"
	"mini-workflow/history/internal/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRecordWorkflowExecutionStarted(t *testing.T) {
	tests := []struct {
		name      string
		req       *pb.RecordWorkflowExecutionStartedRequest
		setup     func(repo *mocks.MockExecutionRepository, matching *mocks.MockMatchingClient, ts *mocks.MockTimerStore)
		wantCode  codes.Code
		wantRunID bool
	}{
		{
			name:     "nil start_request",
			req:      &pb.RecordWorkflowExecutionStartedRequest{},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "missing workflow_id",
			req:      &pb.RecordWorkflowExecutionStartedRequest{StartRequest: &pb.StartWorkflowRequest{WorkflowType: "T"}},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "duplicate running workflow",
			req: &pb.RecordWorkflowExecutionStartedRequest{
				StartRequest: &pb.StartWorkflowRequest{WorkflowId: "wf1", WorkflowType: "T", TaskQueue: "q"},
			},
			setup: func(repo *mocks.MockExecutionRepository, _ *mocks.MockMatchingClient, _ *mocks.MockTimerStore) {
				existing := &domain.WorkflowExecution{WorkflowID: "wf1", RunID: uuid.New(), Status: domain.WorkflowStatusRunning}
				repo.EXPECT().FindRunningWorkflow(mock.Anything, "default", "wf1").Return(existing, nil).Once()
			},
			wantCode: codes.AlreadyExists,
		},
		{
			name: "success",
			req: &pb.RecordWorkflowExecutionStartedRequest{
				StartRequest: &pb.StartWorkflowRequest{WorkflowId: "wf1", WorkflowType: "T", TaskQueue: "q"},
			},
			setup: func(repo *mocks.MockExecutionRepository, matching *mocks.MockMatchingClient, _ *mocks.MockTimerStore) {
				repo.EXPECT().FindRunningWorkflow(mock.Anything, "default", "wf1").Return(nil, nil).Once()
				repo.EXPECT().CreateWorkflowExecution(mock.Anything, mock.AnythingOfType("*domain.WorkflowExecution"), mock.AnythingOfType("*domain.HistoryEvent")).Return(nil).Once()
				matching.EXPECT().AddWorkflowTask(mock.Anything, mock.AnythingOfType("*workflow.AddWorkflowTaskRequest")).Return(&pb.AddWorkflowTaskResponse{}, nil).Once()
			},
			wantCode:  codes.OK,
			wantRunID: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, repo, matching, ts := newTestService(t)
			if tc.setup != nil {
				tc.setup(repo, matching, ts)
			}
			resp, err := svc.RecordWorkflowExecutionStarted(context.Background(), tc.req)
			if tc.wantCode == codes.OK {
				require.NoError(t, err)
				if tc.wantRunID {
					assert.NotEmpty(t, resp.RunId)
					_, parseErr := uuid.Parse(resp.RunId)
					assert.NoError(t, parseErr)
				}
			} else {
				s, _ := status.FromError(err)
				assert.Equal(t, tc.wantCode, s.Code())
			}
		})
	}
}
