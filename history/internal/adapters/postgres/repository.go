package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"mini-workflow/history/internal/domain"
	"mini-workflow/history/internal/ports"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var ErrConcurrencyConflict = errors.New("concurrency conflict: version mismatch")

type postgresRepo struct {
	db *sqlx.DB
}

func NewExecutionRepository(db *sqlx.DB) ports.ExecutionRepository {
	return &postgresRepo{db: db}
}

func (r *postgresRepo) CreateWorkflowExecution(ctx context.Context, exec *domain.WorkflowExecution, initialEvent *domain.HistoryEvent) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.NamedExecContext(ctx, `
		INSERT INTO workflow_executions (namespace, workflow_id, run_id, workflow_type, task_queue, status, current_version, next_event_id, input)
		VALUES (:namespace, :workflow_id, :run_id, :workflow_type, :task_queue, :status, :current_version, :next_event_id, :input)
	`, exec)
	if err != nil {
		return fmt.Errorf("failed to insert workflow execution: %w", err)
	}

	_, err = tx.NamedExecContext(ctx, `
		INSERT INTO history_events (namespace, workflow_id, run_id, event_id, event_type, payload)
		VALUES (:namespace, :workflow_id, :run_id, :event_id, :event_type, :payload)
	`, initialEvent)
	if err != nil {
		return fmt.Errorf("failed to insert initial event: %w", err)
	}

	return tx.Commit()
}

func (r *postgresRepo) GetWorkflowExecution(ctx context.Context, namespace, workflowID string, runID uuid.UUID) (*domain.WorkflowExecution, error) {
	var exec domain.WorkflowExecution
	err := r.db.GetContext(ctx, &exec, `
		SELECT namespace, workflow_id, run_id, workflow_type, task_queue, status, current_version, next_event_id, input, result, start_time, last_updated_time
		FROM workflow_executions
		WHERE namespace = $1 AND workflow_id = $2 AND run_id = $3
	`, namespace, workflowID, runID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("workflow execution not found")
		}
		return nil, err
	}
	return &exec, nil
}

func (r *postgresRepo) FindRunningWorkflow(ctx context.Context, namespace, workflowID string) (*domain.WorkflowExecution, error) {
	var exec domain.WorkflowExecution
	err := r.db.GetContext(ctx, &exec, `
		SELECT namespace, workflow_id, run_id, workflow_type, task_queue, status, current_version, next_event_id, input, result, start_time, last_updated_time
		FROM workflow_executions
		WHERE namespace = $1 AND workflow_id = $2 AND status = $3
		LIMIT 1
	`, namespace, workflowID, domain.WorkflowStatusRunning)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &exec, nil
}

func (r *postgresRepo) UpdateWorkflowExecution(
	ctx context.Context,
	exec *domain.WorkflowExecution,
	expectedVersion int64,
	events []domain.HistoryEvent,
	activitiesToUpsert []domain.ActivityState,
	timersToInsert []domain.Timer,
) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `
		UPDATE workflow_executions
		SET status = $1, current_version = current_version + 1, next_event_id = $2, result = $3, last_updated_time = CURRENT_TIMESTAMP
		WHERE namespace = $4 AND workflow_id = $5 AND run_id = $6 AND current_version = $7
	`, exec.Status, exec.NextEventID, exec.Result, exec.Namespace, exec.WorkflowID, exec.RunID, expectedVersion)
	if err != nil {
		return fmt.Errorf("failed to update execution state: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrConcurrencyConflict
	}

	if len(events) > 0 {
		_, err = tx.NamedExecContext(ctx, `
			INSERT INTO history_events (namespace, workflow_id, run_id, event_id, event_type, payload)
			VALUES (:namespace, :workflow_id, :run_id, :event_id, :event_type, :payload)
		`, events)
		if err != nil {
			return fmt.Errorf("failed to insert history events: %w", err)
		}
	}

	for _, act := range activitiesToUpsert {
		_, err = tx.NamedExecContext(ctx, `
			INSERT INTO activity_states (namespace, workflow_id, run_id, activity_id, status, attempt, last_heartbeat, last_failure_reason, task_token, heartbeat_timeout_seconds, task_queue, activity_type, input)
			VALUES (:namespace, :workflow_id, :run_id, :activity_id, :status, :attempt, :last_heartbeat, :last_failure_reason, :task_token, :heartbeat_timeout_seconds, :task_queue, :activity_type, :input)
			ON CONFLICT (namespace, workflow_id, run_id, activity_id) DO UPDATE SET
				status = EXCLUDED.status,
				attempt = EXCLUDED.attempt,
				last_heartbeat = EXCLUDED.last_heartbeat,
				last_failure_reason = EXCLUDED.last_failure_reason,
				task_token = EXCLUDED.task_token,
				heartbeat_timeout_seconds = EXCLUDED.heartbeat_timeout_seconds,
				task_queue = EXCLUDED.task_queue,
				activity_type = EXCLUDED.activity_type,
				input = EXCLUDED.input
		`, act)
		if err != nil {
			return fmt.Errorf("failed to upsert activity state %s: %w", act.ActivityID, err)
		}
	}

	if len(timersToInsert) > 0 {
		_, err = tx.NamedExecContext(ctx, `
			INSERT INTO timers (timer_id, namespace, workflow_id, run_id, fire_time, timer_type, task_token, is_fired)
			VALUES (:timer_id, :namespace, :workflow_id, :run_id, :fire_time, :timer_type, :task_token, :is_fired)
		`, timersToInsert)
		if err != nil {
			return fmt.Errorf("failed to insert timers: %w", err)
		}
	}

	return tx.Commit()
}

func (r *postgresRepo) GetHistoryEvents(ctx context.Context, namespace, workflowID string, runID uuid.UUID) ([]domain.HistoryEvent, error) {
	var events []domain.HistoryEvent
	err := r.db.SelectContext(ctx, &events, `
		SELECT id, namespace, workflow_id, run_id, event_id, event_type, payload, created_at
		FROM history_events
		WHERE namespace = $1 AND workflow_id = $2 AND run_id = $3
		ORDER BY event_id ASC
	`, namespace, workflowID, runID)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (r *postgresRepo) GetActivityState(ctx context.Context, namespace, workflowID string, runID uuid.UUID, activityID string) (*domain.ActivityState, error) {
	var act domain.ActivityState
	err := r.db.GetContext(ctx, &act, `
		SELECT namespace, workflow_id, run_id, activity_id, status, attempt, last_heartbeat, last_failure_reason, task_token, heartbeat_timeout_seconds, task_queue, activity_type, input
		FROM activity_states
		WHERE namespace = $1 AND workflow_id = $2 AND run_id = $3 AND activity_id = $4
	`, namespace, workflowID, runID, activityID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("activity state not found: %s", activityID)
		}
		return nil, err
	}
	return &act, nil
}

func (r *postgresRepo) MarkTimerFired(ctx context.Context, timerID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE timers SET is_fired = true WHERE timer_id = $1
	`, timerID)
	return err
}

func (r *postgresRepo) GetTimerByID(ctx context.Context, timerID uuid.UUID) (*domain.Timer, error) {
	var t domain.Timer
	err := r.db.GetContext(ctx, &t, `
		SELECT timer_id, namespace, workflow_id, run_id, fire_time, timer_type, task_token, is_fired
		FROM timers
		WHERE timer_id = $1
	`, timerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("timer not found: %s", timerID)
		}
		return nil, err
	}
	return &t, nil
}

func (r *postgresRepo) GetNonFiredTimers(ctx context.Context, upTo time.Time) ([]domain.Timer, error) {
	var timers []domain.Timer
	err := r.db.SelectContext(ctx, &timers, `
		SELECT timer_id, namespace, workflow_id, run_id, fire_time, timer_type, task_token, is_fired
		FROM timers
		WHERE is_fired = false AND fire_time <= $1
		ORDER BY fire_time ASC
	`, upTo)
	if err != nil {
		return nil, err
	}
	return timers, nil
}

func (r *postgresRepo) GetActivityTimeoutTimer(ctx context.Context, namespace, workflowID string, runID uuid.UUID, activityID string) (*domain.Timer, error) {
	var t domain.Timer
	err := r.db.GetContext(ctx, &t, `
		SELECT timer_id, namespace, workflow_id, run_id, fire_time, timer_type, task_token, is_fired
		FROM timers
		WHERE namespace = $1 AND workflow_id = $2 AND run_id = $3
		  AND timer_type = $4 AND is_fired = false
		  AND task_token::text LIKE '%' || $5 || '%'
		ORDER BY fire_time ASC
		LIMIT 1
	`, namespace, workflowID, runID, string(domain.TimerTypeActivityTimeout), activityID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (r *postgresRepo) UpdateTimerFireTime(ctx context.Context, timerID uuid.UUID, newFireTime time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE timers SET fire_time = $1 WHERE timer_id = $2
	`, newFireTime, timerID)
	return err
}
