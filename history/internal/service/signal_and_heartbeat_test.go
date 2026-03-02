package service

import (
	"context"
	"encoding/json"
	"testing"

	pb "mini-workflow/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestSignalWorkflowExecution(t *testing.T) {
	t.Run("invalid run_id", func(t *testing.T) {
		svc, _, _, _ := newTestService(t)
		_, err := svc.SignalWorkflowExecution(context.Background(), &pb.SignalWorkflowRequest{WorkflowId: "wf1", RunId: "bad"})
		s, _ := status.FromError(err)
		assert.Equal(t, codes.InvalidArgument, s.Code())
	})

	t.Run("success — wakes brain", func(t *testing.T) {
		svc, repo, matching, _ := newTestService(t)
		exec := runningExec("wf1")
		payload, _ := json.Marshal([]byte(`"pong"`))

		repo.EXPECT().GetWorkflowExecution(mock.Anything, "default", exec.WorkflowID, exec.RunID).Return(exec, nil).Once()
		repo.EXPECT().UpdateWorkflowExecution(mock.Anything, mock.Anything, exec.CurrentVersion, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
		matching.EXPECT().AddWorkflowTask(mock.Anything, mock.AnythingOfType("*workflow.AddWorkflowTaskRequest")).Return(&pb.AddWorkflowTaskResponse{}, nil).Once()

		resp, err := svc.SignalWorkflowExecution(context.Background(), &pb.SignalWorkflowRequest{
			WorkflowId: exec.WorkflowID,
			RunId:      exec.RunID.String(),
			SignalName: "ping",
			Input:      payload,
		})
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})
}

func TestRecordActivityTaskHeartbeat(t *testing.T) {
	makeToken := func(workflowID, runID, activityID string) []byte {
		tok := taskToken{WorkflowID: workflowID, RunID: runID, ActivityID: activityID, Attempt: 1}
		b, _ := json.Marshal(tok)
		return b
	}

	t.Run("empty task token", func(t *testing.T) {
		svc, _, _, _ := newTestService(t)
		_, err := svc.RecordActivityTaskHeartbeat(context.Background(), &pb.RecordActivityTaskHeartbeatRequest{})
		s, _ := status.FromError(err)
		assert.Equal(t, codes.InvalidArgument, s.Code())
	})

	t.Run("bad task token", func(t *testing.T) {
		svc, _, _, _ := newTestService(t)
		_, err := svc.RecordActivityTaskHeartbeat(context.Background(), &pb.RecordActivityTaskHeartbeatRequest{TaskToken: []byte("!!!")})
		s, _ := status.FromError(err)
		assert.Equal(t, codes.InvalidArgument, s.Code())
	})

	t.Run("success — no cancellation", func(t *testing.T) {
		svc, repo, _, ts := newTestService(t)
		exec := runningExec("wf1")
		actState := buildActivityState(exec, "a1", 1)
		actState.HeartbeatTimeoutSeconds = 30

		repo.EXPECT().GetWorkflowExecution(mock.Anything, "default", exec.WorkflowID, exec.RunID).Return(exec, nil).Once()
		repo.EXPECT().GetActivityState(mock.Anything, "default", exec.WorkflowID, exec.RunID, "a1").Return(actState, nil).Once()
		repo.EXPECT().UpdateWorkflowExecution(mock.Anything, mock.Anything, exec.CurrentVersion, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
		repo.EXPECT().GetActivityTimeoutTimer(mock.Anything, "default", exec.WorkflowID, exec.RunID, "a1").Return(nil, nil).Once()
		_ = ts

		resp, err := svc.RecordActivityTaskHeartbeat(context.Background(), &pb.RecordActivityTaskHeartbeatRequest{
			TaskToken: makeToken(exec.WorkflowID, exec.RunID.String(), "a1"),
		})
		require.NoError(t, err)
		assert.False(t, resp.CancelRequested)
	})
}
