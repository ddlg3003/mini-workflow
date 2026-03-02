package service

import (
	"context"
	"testing"

	pb "mini-workflow/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRecordActivityTaskCompleted(t *testing.T) {
	t.Run("invalid run_id", func(t *testing.T) {
		svc, _, _, _ := newTestService(t)
		_, err := svc.RecordActivityTaskCompleted(context.Background(), &pb.RecordActivityTaskCompletedRequest{WorkflowId: "wf1", RunId: "bad", ActivityId: "a1"})
		s, _ := status.FromError(err)
		assert.Equal(t, codes.InvalidArgument, s.Code())
	})

	t.Run("success — wakes brain", func(t *testing.T) {
		svc, repo, matching, _ := newTestService(t)
		exec := runningExec("wf1")

		repo.EXPECT().GetWorkflowExecution(mock.Anything, "default", exec.WorkflowID, exec.RunID).Return(exec, nil).Once()
		repo.EXPECT().UpdateWorkflowExecution(mock.Anything, mock.Anything, exec.CurrentVersion, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
		matching.EXPECT().AddWorkflowTask(mock.Anything, mock.AnythingOfType("*workflow.AddWorkflowTaskRequest")).Return(&pb.AddWorkflowTaskResponse{}, nil).Once()

		resp, err := svc.RecordActivityTaskCompleted(context.Background(), &pb.RecordActivityTaskCompletedRequest{
			WorkflowId: exec.WorkflowID,
			RunId:      exec.RunID.String(),
			ActivityId: "act-1",
			Result:     []byte(`"ok"`),
		})
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})
}

func TestRecordActivityTaskFailed(t *testing.T) {
	t.Run("invalid run_id", func(t *testing.T) {
		svc, _, _, _ := newTestService(t)
		_, err := svc.RecordActivityTaskFailed(context.Background(), &pb.RecordActivityTaskFailedRequest{WorkflowId: "wf1", RunId: "bad"})
		s, _ := status.FromError(err)
		assert.Equal(t, codes.InvalidArgument, s.Code())
	})

	t.Run("retryable — schedules timer", func(t *testing.T) {
		svc, repo, _, ts := newTestService(t)
		exec := runningExec("wf1")
		actState := buildActivityState(exec, "a1", 1)

		repo.EXPECT().GetWorkflowExecution(mock.Anything, "default", exec.WorkflowID, exec.RunID).Return(exec, nil).Once()
		repo.EXPECT().GetActivityState(mock.Anything, "default", exec.WorkflowID, exec.RunID, "a1").Return(actState, nil).Once()
		repo.EXPECT().UpdateWorkflowExecution(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
		ts.EXPECT().ScheduleTimer(mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("int64")).Return(nil).Once()

		resp, err := svc.RecordActivityTaskFailed(context.Background(), &pb.RecordActivityTaskFailedRequest{
			WorkflowId: exec.WorkflowID,
			RunId:      exec.RunID.String(),
			ActivityId: "a1",
			Reason:     "timeout",
		})
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("final failure — wakes brain", func(t *testing.T) {
		svc, repo, matching, _ := newTestService(t)
		exec := runningExec("wf1")
		actState := buildActivityState(exec, "a1", maxActivityAttempts)

		repo.EXPECT().GetWorkflowExecution(mock.Anything, "default", exec.WorkflowID, exec.RunID).Return(exec, nil).Once()
		repo.EXPECT().GetActivityState(mock.Anything, "default", exec.WorkflowID, exec.RunID, "a1").Return(actState, nil).Once()
		repo.EXPECT().UpdateWorkflowExecution(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
		matching.EXPECT().AddWorkflowTask(mock.Anything, mock.AnythingOfType("*workflow.AddWorkflowTaskRequest")).Return(&pb.AddWorkflowTaskResponse{}, nil).Once()

		resp, err := svc.RecordActivityTaskFailed(context.Background(), &pb.RecordActivityTaskFailedRequest{
			WorkflowId: exec.WorkflowID,
			RunId:      exec.RunID.String(),
			ActivityId: "a1",
			Reason:     "giving up",
		})
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})
}
