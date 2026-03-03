package ports

import (
	"context"
	"time"

	"mini-workflow/history/internal/domain"

	"github.com/google/uuid"
)

type ExecutionRepository interface {
	CreateWorkflowExecution(ctx context.Context, exec *domain.WorkflowExecution, initialEvent *domain.HistoryEvent) error
	GetWorkflowExecution(ctx context.Context, namespace, workflowID string, runID uuid.UUID) (*domain.WorkflowExecution, error)
	UpdateWorkflowExecution(
		ctx context.Context,
		exec *domain.WorkflowExecution,
		expectedVersion int64,
		events []domain.HistoryEvent,
		activitiesToUpsert []domain.ActivityState,
		timersToInsert []domain.Timer,
	) error
	GetHistoryEvents(ctx context.Context, namespace, workflowID string, runID uuid.UUID) ([]domain.HistoryEvent, error)

	GetActivityState(ctx context.Context, namespace, workflowID string, runID uuid.UUID, activityID string) (*domain.ActivityState, error)
	MarkTimerFired(ctx context.Context, timerID uuid.UUID) error
	GetTimerByID(ctx context.Context, timerID uuid.UUID) (*domain.Timer, error)
	GetNonFiredTimers(ctx context.Context, upTo time.Time) ([]domain.Timer, error)
	FindRunningWorkflow(ctx context.Context, namespace, workflowID string) (*domain.WorkflowExecution, error)
	GetActivityTimeoutTimer(ctx context.Context, namespace, workflowID string, runID uuid.UUID, activityID string) (*domain.Timer, error)
	UpdateTimerFireTime(ctx context.Context, timerID uuid.UUID, newFireTime time.Time) error
}

type TimerStore interface {
	ScheduleTimer(ctx context.Context, timerID string, fireTimeMs int64) error
	ClaimExpired(ctx context.Context, nowMs int64, batchSize int) ([]string, error)
	Subscribe(ctx context.Context) (<-chan struct{}, func())
	RebuildFromDB(ctx context.Context, timers []domain.Timer) error
}
