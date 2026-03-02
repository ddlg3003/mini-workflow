package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// WorkflowStatus represents the current state of a workflow execution
type WorkflowStatus string

const (
	WorkflowStatusRunning   WorkflowStatus = "Running"
	WorkflowStatusCompleted WorkflowStatus = "Completed"
	WorkflowStatusFailed    WorkflowStatus = "Failed"
	WorkflowStatusTimedOut  WorkflowStatus = "TimedOut"
)

// ActivityStatus represents the state of an activity task
type ActivityStatus string

const (
	ActivityStatusScheduled ActivityStatus = "Scheduled"
	ActivityStatusStarted   ActivityStatus = "Started"
	ActivityStatusCompleted ActivityStatus = "Completed"
	ActivityStatusFailed    ActivityStatus = "Failed"
)

// TimerType represents the type of timer
type TimerType string

const (
	TimerTypeUser                TimerType = "UserTimer"
	TimerTypeActivityTimeout     TimerType = "ActivityTimeout"
	TimerTypeWorkflowTimeout     TimerType = "WorkflowTimeout"
	TimerTypeWorkflowTaskTimeout TimerType = "WorkflowTaskTimeout"
)

// WorkflowExecution represents a row in the workflow_executions table
type WorkflowExecution struct {
	Namespace       string           `db:"namespace"`
	WorkflowID      string           `db:"workflow_id"`
	RunID           uuid.UUID        `db:"run_id"`
	WorkflowType    string           `db:"workflow_type"`
	TaskQueue       string           `db:"task_queue"`
	Status          WorkflowStatus   `db:"status"`
	CurrentVersion  int64            `db:"current_version"`
	NextEventID     int64            `db:"next_event_id"`
	Input           json.RawMessage  `db:"input"`
	Result          *json.RawMessage `db:"result"`
	StartTime       time.Time        `db:"start_time"`
	LastUpdatedTime time.Time        `db:"last_updated_time"`
}

// HistoryEvent represents a row in the history_events table
type HistoryEvent struct {
	ID         int64           `db:"id"`
	Namespace  string          `db:"namespace"`
	WorkflowID string          `db:"workflow_id"`
	RunID      uuid.UUID       `db:"run_id"`
	EventID    int64           `db:"event_id"`
	EventType  string          `db:"event_type"` // e.g., 'WorkflowExecutionStarted', 'ActivityTaskScheduled'
	Payload    json.RawMessage `db:"payload"`
	CreatedAt  time.Time       `db:"created_at"`
}

// ActivityState represents a row in the activity_states table
type ActivityState struct {
	Namespace               string         `db:"namespace"`
	WorkflowID              string         `db:"workflow_id"`
	RunID                   uuid.UUID      `db:"run_id"`
	ActivityID              string         `db:"activity_id"`
	Status                  ActivityStatus `db:"status"`
	Attempt                 int            `db:"attempt"`
	LastHeartbeat           *time.Time     `db:"last_heartbeat"`
	LastFailureReason       *string        `db:"last_failure_reason"`
	TaskToken               []byte         `db:"task_token"`
	HeartbeatTimeoutSeconds int            `db:"heartbeat_timeout_seconds"`
	TaskQueue               string         `db:"task_queue"`
	ActivityType            string         `db:"activity_type"`
	Input                   []byte         `db:"input"`
}

// Timer represents a row in the timers table
type Timer struct {
	TimerID    uuid.UUID `db:"timer_id"`
	Namespace  string    `db:"namespace"`
	WorkflowID string    `db:"workflow_id"`
	RunID      uuid.UUID `db:"run_id"`
	FireTime   time.Time `db:"fire_time"`
	TimerType  TimerType `db:"timer_type"`
	TaskToken  []byte    `db:"task_token"`
	IsFired    bool      `db:"is_fired"`
}
