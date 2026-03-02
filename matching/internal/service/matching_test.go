package service

import (
	"context"
	"encoding/json"
	"testing"

	pb "mini-workflow/api"
	"mini-workflow/config"
	"mini-workflow/matching/internal/domain"
	"mini-workflow/matching/internal/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newSvc(t *testing.T) (*matchingService, *mocks.MockTaskQueue) {
	q := mocks.NewMockTaskQueue(t)
	return New(q, config.ServiceConfig{PollTimeoutSeconds: 60}), q
}

// ---------- AddWorkflowTask ----------

func TestAddWorkflowTask(t *testing.T) {
	t.Run("missing workflow_id → InvalidArgument", func(t *testing.T) {
		svc, _ := newSvc(t)
		_, err := svc.AddWorkflowTask(context.Background(), &pb.AddWorkflowTaskRequest{RunId: "r1"})
		s, _ := status.FromError(err)
		assert.Equal(t, codes.InvalidArgument, s.Code())
	})

	t.Run("success — pushes to workflow key", func(t *testing.T) {
		svc, q := newSvc(t)
		expectedKey := "task_queue:default:q1:workflow"

		q.EXPECT().Push(mock.Anything, expectedKey, mock.MatchedBy(func(b []byte) bool {
			var p domain.WorkflowTask
			return json.Unmarshal(b, &p) == nil && p.WorkflowID == "wf1" && p.RunID == "r1"
		})).Return(nil).Once()

		resp, err := svc.AddWorkflowTask(context.Background(), &pb.AddWorkflowTaskRequest{
			WorkflowId: "wf1",
			RunId:      "r1",
			TaskQueue:  "q1",
		})
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})
}

// ---------- AddActivityTask ----------

func TestAddActivityTask(t *testing.T) {
	t.Run("missing activity_id → InvalidArgument", func(t *testing.T) {
		svc, _ := newSvc(t)
		_, err := svc.AddActivityTask(context.Background(), &pb.AddActivityTaskRequest{WorkflowId: "wf1", RunId: "r1"})
		s, _ := status.FromError(err)
		assert.Equal(t, codes.InvalidArgument, s.Code())
	})

	t.Run("success — pushes to activity key", func(t *testing.T) {
		svc, q := newSvc(t)
		expectedKey := "task_queue:default:q1:activity"

		q.EXPECT().Push(mock.Anything, expectedKey, mock.MatchedBy(func(b []byte) bool {
			var p domain.ActivityTask
			return json.Unmarshal(b, &p) == nil && p.ActivityID == "a1" && p.WorkflowID == "wf1"
		})).Return(nil).Once()

		resp, err := svc.AddActivityTask(context.Background(), &pb.AddActivityTaskRequest{
			WorkflowId: "wf1",
			RunId:      "r1",
			ActivityId: "a1",
			TaskQueue:  "q1",
			Input:      []byte(`"hello"`),
		})
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})
}

// ---------- PollWorkflowTaskQueue ----------

func TestPollWorkflowTaskQueue(t *testing.T) {
	t.Run("task_queue required → InvalidArgument", func(t *testing.T) {
		svc, _ := newSvc(t)
		_, err := svc.PollWorkflowTaskQueue(context.Background(), &pb.PollWorkflowTaskQueueRequest{})
		s, _ := status.FromError(err)
		assert.Equal(t, codes.InvalidArgument, s.Code())
	})

	t.Run("task available → returns task token", func(t *testing.T) {
		svc, q := newSvc(t)
		payload, _ := json.Marshal(domain.WorkflowTask{WorkflowID: "wf1", RunID: "r1", TaskQueue: "q1"})

		q.EXPECT().BlockPop(mock.Anything, "task_queue:default:q1:workflow", 60).Return(payload, nil).Once()

		resp, err := svc.PollWorkflowTaskQueue(context.Background(), &pb.PollWorkflowTaskQueueRequest{TaskQueue: "q1"})
		require.NoError(t, err)

		var tok map[string]string
		require.NoError(t, json.Unmarshal(resp.TaskToken, &tok))
		assert.Equal(t, "wf1", tok["workflow_id"])
		assert.Equal(t, "r1", tok["run_id"])
	})

	t.Run("timeout → NotFound", func(t *testing.T) {
		svc, q := newSvc(t)
		q.EXPECT().BlockPop(mock.Anything, "task_queue:default:q1:workflow", 60).Return(nil, nil).Once()

		_, err := svc.PollWorkflowTaskQueue(context.Background(), &pb.PollWorkflowTaskQueueRequest{TaskQueue: "q1"})
		s, _ := status.FromError(err)
		assert.Equal(t, codes.NotFound, s.Code())
	})
}

// ---------- PollActivityTaskQueue ----------

func TestPollActivityTaskQueue(t *testing.T) {
	t.Run("task_queue required → InvalidArgument", func(t *testing.T) {
		svc, _ := newSvc(t)
		_, err := svc.PollActivityTaskQueue(context.Background(), &pb.PollActivityTaskQueueRequest{})
		s, _ := status.FromError(err)
		assert.Equal(t, codes.InvalidArgument, s.Code())
	})

	t.Run("task available → returns full payload", func(t *testing.T) {
		svc, q := newSvc(t)
		payload, _ := json.Marshal(domain.ActivityTask{
			WorkflowID:              "wf1",
			RunID:                   "r1",
			ActivityID:              "a1",
			ActivityType:            "MyActivity",
			Input:                   []byte(`"input"`),
			HeartbeatTimeoutSeconds: 30,
		})

		q.EXPECT().BlockPop(mock.Anything, "task_queue:default:q1:activity", 60).Return(payload, nil).Once()

		resp, err := svc.PollActivityTaskQueue(context.Background(), &pb.PollActivityTaskQueueRequest{TaskQueue: "q1"})
		require.NoError(t, err)
		assert.Equal(t, "a1", resp.ActivityId)
		assert.Equal(t, "MyActivity", resp.ActivityType)
		assert.Equal(t, int32(30), resp.HeartbeatTimeoutSeconds)
	})

	t.Run("timeout → NotFound", func(t *testing.T) {
		svc, q := newSvc(t)
		q.EXPECT().BlockPop(mock.Anything, "task_queue:default:q1:activity", 60).Return(nil, nil).Once()

		_, err := svc.PollActivityTaskQueue(context.Background(), &pb.PollActivityTaskQueueRequest{TaskQueue: "q1"})
		s, _ := status.FromError(err)
		assert.Equal(t, codes.NotFound, s.Code())
	})
}
