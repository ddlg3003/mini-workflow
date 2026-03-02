package service

import (
	"context"
	"testing"

	pb "mini-workflow/api"
	"mini-workflow/history/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRecordWorkflowTaskCompleted(t *testing.T) {
	t.Run("invalid run_id", func(t *testing.T) {
		svc, _, _, _ := newTestService(t)
		_, err := svc.RecordWorkflowTaskCompleted(context.Background(), &pb.RecordWorkflowTaskCompletedHistoryRequest{WorkflowId: "wf1", RunId: "bad"})
		s, _ := status.FromError(err)
		assert.Equal(t, codes.InvalidArgument, s.Code())
	})

	t.Run("schedule activity — enqueues activity task", func(t *testing.T) {
		svc, repo, matching, ts := newTestService(t)
		exec := runningExec("wf1")

		repo.EXPECT().GetWorkflowExecution(mock.Anything, "default", exec.WorkflowID, exec.RunID).Return(exec, nil).Once()
		repo.EXPECT().UpdateWorkflowExecution(mock.Anything, mock.Anything, exec.CurrentVersion, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
		ts.EXPECT().ScheduleTimer(mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("int64")).Return(nil).Once()
		matching.EXPECT().AddActivityTask(mock.Anything, mock.AnythingOfType("*workflow.AddActivityTaskRequest")).Return(&pb.AddActivityTaskResponse{}, nil).Once()

		resp, err := svc.RecordWorkflowTaskCompleted(context.Background(), &pb.RecordWorkflowTaskCompletedHistoryRequest{
			WorkflowId: exec.WorkflowID,
			RunId:      exec.RunID.String(),
			Commands:   []*pb.Command{scheduleActivityCmd("act-1", "q")},
		})
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("complete workflow — sets status Completed", func(t *testing.T) {
		svc, repo, _, _ := newTestService(t)
		exec := runningExec("wf1")

		repo.EXPECT().GetWorkflowExecution(mock.Anything, "default", exec.WorkflowID, exec.RunID).Return(exec, nil).Once()
		repo.EXPECT().UpdateWorkflowExecution(mock.Anything, mock.MatchedBy(func(e *domain.WorkflowExecution) bool {
			return e.Status == domain.WorkflowStatusCompleted
		}), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		resp, err := svc.RecordWorkflowTaskCompleted(context.Background(), &pb.RecordWorkflowTaskCompletedHistoryRequest{
			WorkflowId: exec.WorkflowID,
			RunId:      exec.RunID.String(),
			Commands:   []*pb.Command{completeWorkflowCmd([]byte(`"done"`))},
		})
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("start timer — schedules timer in redis", func(t *testing.T) {
		svc, repo, _, ts := newTestService(t)
		exec := runningExec("wf1")

		repo.EXPECT().GetWorkflowExecution(mock.Anything, "default", exec.WorkflowID, exec.RunID).Return(exec, nil).Once()
		repo.EXPECT().UpdateWorkflowExecution(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
		ts.EXPECT().ScheduleTimer(mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("int64")).Return(nil).Once()

		resp, err := svc.RecordWorkflowTaskCompleted(context.Background(), &pb.RecordWorkflowTaskCompletedHistoryRequest{
			WorkflowId: exec.WorkflowID,
			RunId:      exec.RunID.String(),
			Commands:   []*pb.Command{startTimerCmd(10)},
		})
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})
}
