package service

import (
	"context"
	"encoding/json"
	"time"

	pb "mini-workflow/api"
	"mini-workflow/history/internal/domain"
	"mini-workflow/history/internal/ports"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	timerBatchSize    = 10
	scavengerInterval = 5 * time.Minute
	maxIdleWait       = 1 * time.Hour
)

type TimerProcessor struct {
	repo       ports.ExecutionRepository
	timerStore ports.TimerStore
	matching   ports.MatchingClient
	log        *zap.Logger
}

func NewTimerProcessor(repo ports.ExecutionRepository, timerStore ports.TimerStore, matching ports.MatchingClient, log *zap.Logger) *TimerProcessor {
	return &TimerProcessor{repo: repo, timerStore: timerStore, matching: matching, log: log}
}

func (p *TimerProcessor) Run(ctx context.Context) {
	wakeupCh, unsubscribe := p.timerStore.Subscribe(ctx)
	defer unsubscribe()

	scavengerTicker := time.NewTicker(scavengerInterval)
	defer scavengerTicker.Stop()

	for {
		p.processExpired(ctx)

		waitDur := p.nextWaitDuration(ctx)

		select {
		case <-time.After(waitDur):
		case <-wakeupCh:
		case <-scavengerTicker.C:
			p.scavenge(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (p *TimerProcessor) processExpired(ctx context.Context) {
	nowMs := time.Now().UnixMilli()
	ids, err := p.timerStore.ClaimExpired(ctx, nowMs, timerBatchSize)
	if err != nil {
		p.log.Error("claim expired timers failed", zap.Error(err))
		return
	}

	for _, id := range ids {
		timerID, err := uuid.Parse(id)
		if err != nil {
			p.log.Error("invalid timer_id in redis set", zap.String("id", id), zap.Error(err))
			continue
		}
		p.dispatch(ctx, timerID)
	}
}

func (p *TimerProcessor) dispatch(ctx context.Context, timerID uuid.UUID) {
	t, err := p.repo.GetTimerByID(ctx, timerID)
	if err != nil {
		p.log.Error("get timer by id failed", zap.String("timer_id", timerID.String()), zap.Error(err))
		return
	}

	if err := p.repo.MarkTimerFired(ctx, timerID); err != nil {
		p.log.Error("mark timer fired failed", zap.String("timer_id", timerID.String()), zap.Error(err))
		return
	}

	switch t.TimerType {
	case domain.TimerTypeUser:
		exec, err := p.processExecutionTimer(ctx, t, "TimerFired", "")
		if err != nil {
			return
		}

		if _, err := p.matching.AddWorkflowTask(ctx, &pb.AddWorkflowTaskRequest{
			TaskQueue:    exec.TaskQueue,
			WorkflowId:   t.WorkflowID,
			RunId:        t.RunID.String(),
			WorkflowType: exec.WorkflowType,
		}); err != nil {
			p.log.Warn("add workflow task for timer failed", zap.String("timer_id", timerID.String()), zap.Error(err))
		}

	case domain.TimerTypeWorkflowTimeout:
		_, _ = p.processExecutionTimer(ctx, t, "WorkflowExecutionTimedOut", domain.WorkflowStatusTimedOut)

	case domain.TimerTypeActivityTimeout:
		p.handleActivityTimeout(ctx, t)

	case domain.TimerTypeWorkflowTaskTimeout:
		p.handleWorkflowTaskTimeout(ctx, t)
	}
}

func (p *TimerProcessor) processExecutionTimer(ctx context.Context, t *domain.Timer, eventType string, newStatus domain.WorkflowStatus) (*domain.WorkflowExecution, error) {
	exec, err := p.repo.GetWorkflowExecution(ctx, t.Namespace, t.WorkflowID, t.RunID)
	if err != nil {
		p.log.Error("get workflow execution for timer failed", zap.String("workflow_id", t.WorkflowID), zap.Error(err))
		return nil, err
	}

	expectedVersion := exec.CurrentVersion
	eventID := exec.NextEventID

	payload, _ := json.Marshal(map[string]any{
		"timer_id": t.TimerID.String(),
	})

	event := domain.HistoryEvent{
		Namespace:  t.Namespace,
		WorkflowID: t.WorkflowID,
		RunID:      t.RunID,
		EventID:    eventID,
		EventType:  eventType,
		Payload:    payload,
	}

	exec.NextEventID = eventID + 1
	if newStatus != "" {
		exec.Status = newStatus
	}

	if err := p.repo.UpdateWorkflowExecution(ctx, exec, expectedVersion, []domain.HistoryEvent{event}, nil, nil); err != nil {
		p.log.Error("update execution with timer event failed", zap.Error(err))
		return nil, err
	}

	return exec, nil
}

func (p *TimerProcessor) handleActivityTimeout(ctx context.Context, t *domain.Timer) {
	// Decode which activity this timer belongs to and what the expected heartbeat window is.
	tok, err := decodeActivityTimerToken(t.TaskToken)
	if err != nil || tok.ActivityID == "" {
		p.log.Error("invalid activity timer token", zap.String("timer_id", t.TimerID.String()), zap.Error(err))
		return
	}

	actState, err := p.repo.GetActivityState(ctx, t.Namespace, t.WorkflowID, t.RunID, tok.ActivityID)
	if err != nil {
		p.log.Error("get activity state for timeout failed", zap.String("activity_id", tok.ActivityID), zap.Error(err))
		return
	}

	// Double-check: the Redis timer may have fired slightly early or been claimed
	// before a concurrent heartbeat could push the deadline forward. By re-reading
	// last_heartbeat from the DB, we give the worker one last benefit of the doubt.
	// If the worker is still within its window, reschedule the timer and walk away.
	// Only if it truly went silent do we declare a timeout and re-enqueue the task.
	timeoutSecs := tok.HeartbeatTimeoutSeconds
	if timeoutSecs <= 0 {
		timeoutSecs = 30
	}
	if actState.LastHeartbeat != nil {
		nextDeadline := actState.LastHeartbeat.Add(time.Duration(timeoutSecs) * time.Second)
		if nextDeadline.After(time.Now()) {
			// Worker is still alive — reschedule the watchdog timer to the real deadline.
			if err := p.timerStore.ScheduleTimer(ctx, t.TimerID.String(), nextDeadline.UnixMilli()); err != nil {
				p.log.Warn("reschedule stale activity timer failed", zap.String("timer_id", t.TimerID.String()), zap.Error(err))
			}
			return
		}
	}

	// The worker is silent — treat it as a failure / retry.
	p.log.Info("activity heartbeat timeout detected",
		zap.String("activity_id", tok.ActivityID),
		zap.String("workflow_id", t.WorkflowID),
	)
	if _, err := p.matching.AddActivityTask(ctx, &pb.AddActivityTaskRequest{
		TaskQueue:               actState.TaskQueue,
		WorkflowId:              t.WorkflowID,
		RunId:                   t.RunID.String(),
		ActivityId:              tok.ActivityID,
		ActivityType:            actState.ActivityType,
		Input:                   actState.Input,
		HeartbeatTimeoutSeconds: int32(actState.HeartbeatTimeoutSeconds),
	}); err != nil {
		p.log.Warn("add activity task after timeout failed", zap.String("activity_id", tok.ActivityID), zap.Error(err))
	}
}

func (p *TimerProcessor) handleWorkflowTaskTimeout(ctx context.Context, t *domain.Timer) {
	// Parse the TaskToken to retrieve the event_id of the WorkflowTaskStarted event
	var tok struct {
		EventID int64 `json:"event_id"`
	}
	if err := json.Unmarshal(t.TaskToken, &tok); err != nil {
		p.log.Error("invalid workflow task timeout token", zap.String("timer_id", t.TimerID.String()), zap.Error(err))
		return
	}

	exec, err := p.repo.GetWorkflowExecution(ctx, t.Namespace, t.WorkflowID, t.RunID)
	if err != nil {
		p.log.Error("get workflow execution for workflow task timeout failed", zap.String("workflow_id", t.WorkflowID), zap.Error(err))
		return
	}

	// Replay Safety Check:
	// If the status is no longer Running, or if NextEventID has moved past the Started event,
	// it means the worker either successfully replied or the task was already re-queued
	// and processed. We can safely ignore this timer.
	if exec.Status != domain.WorkflowStatusRunning || exec.NextEventID > tok.EventID+1 {
		p.log.Debug("workflow task timeout ignored (already progressed or completed)",
			zap.String("workflow_id", t.WorkflowID),
			zap.Int64("expected_event_id", tok.EventID+1),
			zap.Int64("actual_next_event_id", exec.NextEventID),
			zap.String("status", string(exec.Status)),
		)
		return
	}

	// The worker has stall/crashed. Re-queue the task to Matching without adding a new event.
	p.log.Info("workflow task timeout detected. recovery initiated.",
		zap.String("workflow_id", t.WorkflowID),
		zap.String("run_id", t.RunID.String()),
		zap.Int64("event_id", tok.EventID),
	)

	if _, err := p.matching.AddWorkflowTask(ctx, &pb.AddWorkflowTaskRequest{
		TaskQueue:    exec.TaskQueue,
		WorkflowId:   t.WorkflowID,
		RunId:        t.RunID.String(),
		WorkflowType: exec.WorkflowType,
	}); err != nil {
		p.log.Error("failed to recover workflow task (AddWorkflowTask failed)",
			zap.String("workflow_id", t.WorkflowID),
			zap.Error(err),
		)
	}
}

func (p *TimerProcessor) nextWaitDuration(ctx context.Context) time.Duration {
	ids, err := p.timerStore.ClaimExpired(ctx, time.Now().Add(maxIdleWait).UnixMilli()+1, 1)
	if err != nil || len(ids) == 0 {
		return maxIdleWait
	}
	timerID, err := uuid.Parse(ids[0])
	if err != nil {
		return maxIdleWait
	}
	t, err := p.repo.GetTimerByID(ctx, timerID)
	if err != nil {
		return maxIdleWait
	}
	if err := p.timerStore.ScheduleTimer(ctx, ids[0], t.FireTime.UnixMilli()); err != nil {
		p.log.Warn("re-schedule peeked timer failed", zap.Error(err))
	}
	wait := time.Until(t.FireTime)
	if wait < 0 {
		return 0
	}
	return wait
}

func (p *TimerProcessor) scavenge(ctx context.Context) {
	upTo := time.Now().Add(1 * time.Hour)
	timers, err := p.repo.GetNonFiredTimers(ctx, upTo)
	if err != nil {
		p.log.Error("scavenger: get non-fired timers failed", zap.Error(err))
		return
	}
	if err := p.timerStore.RebuildFromDB(ctx, timers); err != nil {
		p.log.Error("scavenger: rebuild timer store failed", zap.Error(err))
	}
}
