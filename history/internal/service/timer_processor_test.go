package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	pb "mini-workflow/api"
	"mini-workflow/history/internal/domain"
	"mini-workflow/history/internal/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestTimerProcessorDispatch(t *testing.T) {
	t.Run("user timer fires workflow task", func(t *testing.T) {
		repo := mocks.NewMockExecutionRepository(t)
		ts := mocks.NewMockTimerStore(t)
		matching := mocks.NewMockMatchingClient(t)

		timerID := uuid.New()
		runID := uuid.New()
		timer := &domain.Timer{
			TimerID:    timerID,
			Namespace:  "default",
			WorkflowID: "wf1",
			RunID:      runID,
			FireTime:   time.Now().Add(-time.Second),
			TimerType:  domain.TimerTypeUser,
		}

		exec := &domain.WorkflowExecution{
			Namespace:    "default",
			WorkflowID:   "wf1",
			RunID:        runID,
			TaskQueue:    "test-queue",
			WorkflowType: "test-wf",
		}

		repo.EXPECT().GetTimerByID(mock.Anything, timerID).Return(timer, nil).Once()
		repo.EXPECT().MarkTimerFired(mock.Anything, timerID).Return(nil).Once()
		repo.EXPECT().GetWorkflowExecution(mock.Anything, "default", "wf1", runID).Return(exec, nil).Once()

		repo.EXPECT().UpdateWorkflowExecution(mock.Anything, mock.Anything, int64(0), mock.MatchedBy(func(events []domain.HistoryEvent) bool {
			return len(events) == 1 && events[0].EventType == "TimerFired"
		}), []domain.ActivityState(nil), []domain.Timer(nil)).Return(nil).Once()

		matching.EXPECT().AddWorkflowTask(mock.Anything, mock.MatchedBy(func(req *pb.AddWorkflowTaskRequest) bool {
			return req.TaskQueue == "test-queue" && req.WorkflowType == "test-wf"
		})).Return(&pb.AddWorkflowTaskResponse{}, nil).Once()

		tp := NewTimerProcessor(repo, ts, matching, zap.NewNop())
		tp.dispatch(context.Background(), timerID)
	})

	t.Run("activity timeout fires activity task", func(t *testing.T) {
		repo := mocks.NewMockExecutionRepository(t)
		ts := mocks.NewMockTimerStore(t)
		matching := mocks.NewMockMatchingClient(t)

		timerID := uuid.New()
		runID := uuid.New()
		hbToken, _ := json.Marshal(map[string]any{
			"activity_id":               "a1",
			"heartbeat_timeout_seconds": 30,
		})
		timer := &domain.Timer{
			TimerID:    timerID,
			Namespace:  "default",
			WorkflowID: "wf1",
			RunID:      runID,
			FireTime:   time.Now().Add(-time.Second),
			TimerType:  domain.TimerTypeActivityTimeout,
			TaskToken:  hbToken,
		}
		actState := &domain.ActivityState{
			Namespace:  "default",
			WorkflowID: "wf1",
			RunID:      runID,
			ActivityID: "a1",
			Status:     domain.ActivityStatusStarted,
			Attempt:    1,
		}

		repo.EXPECT().GetTimerByID(mock.Anything, timerID).Return(timer, nil).Once()
		repo.EXPECT().MarkTimerFired(mock.Anything, timerID).Return(nil).Once()
		repo.EXPECT().GetActivityState(mock.Anything, "default", "wf1", runID, "a1").Return(actState, nil).Once()
		matching.EXPECT().AddActivityTask(mock.Anything, mock.AnythingOfType("*workflow.AddActivityTaskRequest")).Return(&pb.AddActivityTaskResponse{}, nil).Once()
		_ = ts

		tp := NewTimerProcessor(repo, ts, matching, zap.NewNop())
		tp.dispatch(context.Background(), timerID)
	})

	t.Run("scavenger rebuilds timer store", func(t *testing.T) {
		repo := mocks.NewMockExecutionRepository(t)
		ts := mocks.NewMockTimerStore(t)
		matching := mocks.NewMockMatchingClient(t)

		timers := []domain.Timer{
			{TimerID: uuid.New(), FireTime: time.Now().Add(time.Minute)},
			{TimerID: uuid.New(), FireTime: time.Now().Add(2 * time.Minute)},
		}

		repo.EXPECT().GetNonFiredTimers(mock.Anything, mock.Anything).Return(timers, nil).Once()
		ts.EXPECT().RebuildFromDB(mock.Anything, timers).Return(nil).Once()

		tp := NewTimerProcessor(repo, ts, matching, zap.NewNop())
		tp.scavenge(context.Background())
	})

	// --- handleActivityTimeout edge cases ---

	t.Run("stale timer: worker still alive → reschedule, no AddActivityTask", func(t *testing.T) {
		repo := mocks.NewMockExecutionRepository(t)
		ts := mocks.NewMockTimerStore(t)
		matching := mocks.NewMockMatchingClient(t)

		timerID := uuid.New()
		runID := uuid.New()
		hbToken, _ := json.Marshal(map[string]any{
			"activity_id":               "a1",
			"heartbeat_timeout_seconds": 30,
		})
		timer := &domain.Timer{
			TimerID:    timerID,
			Namespace:  "default",
			WorkflowID: "wf1",
			RunID:      runID,
			TimerType:  domain.TimerTypeActivityTimeout,
			TaskToken:  hbToken,
		}
		// last_heartbeat was just 5s ago — well within the 30s window
		recentHeartbeat := time.Now().Add(-5 * time.Second)
		actState := &domain.ActivityState{
			Namespace:     "default",
			WorkflowID:    "wf1",
			RunID:         runID,
			ActivityID:    "a1",
			Status:        domain.ActivityStatusStarted,
			LastHeartbeat: &recentHeartbeat,
		}

		// Expect a reschedule into Redis (new deadline ~25s from now) and NO AddActivityTask.
		repo.EXPECT().GetTimerByID(mock.Anything, timerID).Return(timer, nil).Once()
		repo.EXPECT().MarkTimerFired(mock.Anything, timerID).Return(nil).Once()
		repo.EXPECT().GetActivityState(mock.Anything, "default", "wf1", runID, "a1").Return(actState, nil).Once()
		ts.EXPECT().ScheduleTimer(mock.Anything, timerID.String(), mock.AnythingOfType("int64")).Return(nil).Once()
		// matching.AddActivityTask must NOT be called — mockery enforces this automatically.

		tp := NewTimerProcessor(repo, ts, matching, zap.NewNop())
		tp.dispatch(context.Background(), timerID)
	})

	t.Run("corrupt task token → bail out silently without crashing", func(t *testing.T) {
		repo := mocks.NewMockExecutionRepository(t)
		ts := mocks.NewMockTimerStore(t)
		matching := mocks.NewMockMatchingClient(t)

		timerID := uuid.New()
		timer := &domain.Timer{
			TimerID:   timerID,
			Namespace: "default",
			TimerType: domain.TimerTypeActivityTimeout,
			TaskToken: []byte("!!!not-json!!!"),
		}

		repo.EXPECT().GetTimerByID(mock.Anything, timerID).Return(timer, nil).Once()
		repo.EXPECT().MarkTimerFired(mock.Anything, timerID).Return(nil).Once()
		// Neither GetActivityState nor AddActivityTask should be called.

		tp := NewTimerProcessor(repo, ts, matching, zap.NewNop())
		tp.dispatch(context.Background(), timerID) // must not panic
	})

	t.Run("workflow task timeout recovers stalled task", func(t *testing.T) {
		repo := mocks.NewMockExecutionRepository(t)
		ts := mocks.NewMockTimerStore(t)
		matching := mocks.NewMockMatchingClient(t)

		timerID := uuid.New()
		runID := uuid.New()
		token, _ := json.Marshal(map[string]any{
			"event_id": 5, // The event ID of WorkflowTaskStarted
		})
		timer := &domain.Timer{
			TimerID:    timerID,
			Namespace:  "default",
			WorkflowID: "wf1",
			RunID:      runID,
			TimerType:  domain.TimerTypeWorkflowTaskTimeout,
			TaskToken:  token,
		}

		exec := &domain.WorkflowExecution{
			Namespace:    "default",
			WorkflowID:   "wf1",
			RunID:        runID,
			Status:       domain.WorkflowStatusRunning,
			TaskQueue:    "main-queue",
			WorkflowType: "MyWorkflow",
			NextEventID:  6, // Still 6 (Worker has not replied yet. NextEventID is event_id + 1)
		}

		repo.EXPECT().GetTimerByID(mock.Anything, timerID).Return(timer, nil).Once()
		repo.EXPECT().MarkTimerFired(mock.Anything, timerID).Return(nil).Once()
		repo.EXPECT().GetWorkflowExecution(mock.Anything, "default", "wf1", runID).Return(exec, nil).Once()

		// Expect it to re-queue the task
		matching.EXPECT().AddWorkflowTask(mock.Anything, mock.MatchedBy(func(req *pb.AddWorkflowTaskRequest) bool {
			return req.TaskQueue == "main-queue" && req.WorkflowId == "wf1" && req.RunId == runID.String() && req.WorkflowType == "MyWorkflow"
		})).Return(&pb.AddWorkflowTaskResponse{}, nil).Once()

		tp := NewTimerProcessor(repo, ts, matching, zap.NewNop())
		tp.dispatch(context.Background(), timerID)
	})

	t.Run("workflow task timeout ignored if task already progressed", func(t *testing.T) {
		repo := mocks.NewMockExecutionRepository(t)
		ts := mocks.NewMockTimerStore(t)
		matching := mocks.NewMockMatchingClient(t)

		timerID := uuid.New()
		runID := uuid.New()
		token, _ := json.Marshal(map[string]any{
			"event_id": 5,
		})
		timer := &domain.Timer{
			TimerID:    timerID,
			Namespace:  "default",
			WorkflowID: "wf1",
			RunID:      runID,
			TimerType:  domain.TimerTypeWorkflowTaskTimeout,
			TaskToken:  token,
		}

		exec := &domain.WorkflowExecution{
			Namespace:    "default",
			WorkflowID:   "wf1",
			RunID:        runID,
			Status:       domain.WorkflowStatusRunning,
			TaskQueue:    "main-queue",
			WorkflowType: "MyWorkflow",
			NextEventID:  7, // Moved past 6, meaning the task was completed or failed and progressed
		}

		repo.EXPECT().GetTimerByID(mock.Anything, timerID).Return(timer, nil).Once()
		repo.EXPECT().MarkTimerFired(mock.Anything, timerID).Return(nil).Once()
		repo.EXPECT().GetWorkflowExecution(mock.Anything, "default", "wf1", runID).Return(exec, nil).Once()

		// Should NOT call AddWorkflowTask
		tp := NewTimerProcessor(repo, ts, matching, zap.NewNop())
		tp.dispatch(context.Background(), timerID)
	})
}
