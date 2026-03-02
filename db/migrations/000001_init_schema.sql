-- +migrate Up
-- 1. Snapshot of the current state for each Workflow
CREATE TABLE workflow_executions (
    namespace          VARCHAR(255) NOT NULL DEFAULT 'default',
    workflow_id        VARCHAR(255) NOT NULL,
    run_id             UUID NOT NULL,
    workflow_type      VARCHAR(255) NOT NULL,
    task_queue         VARCHAR(255) NOT NULL,
    status             VARCHAR(50) NOT NULL, -- Running, Completed, Failed, TimedOut
    current_version    BIGINT NOT NULL DEFAULT 1, -- Optimistic Concurrency Control
    next_event_id      BIGINT NOT NULL DEFAULT 1, -- Incremented for every history event
    input              JSONB,
    result             JSONB,
    start_time         TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_updated_time  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (namespace, workflow_id, run_id)
);

-- 2. Append-Only Immutable Log (Heart of Replay)
CREATE TABLE history_events (
    id                 BIGSERIAL PRIMARY KEY,
    namespace          VARCHAR(255) NOT NULL DEFAULT 'default',
    workflow_id        VARCHAR(255) NOT NULL,
    run_id             UUID NOT NULL,
    event_id           BIGINT NOT NULL, -- Sequential ID within the workflow
    event_type         VARCHAR(100) NOT NULL, -- e.g., 'ActivityScheduled', 'TimerStarted'
    payload            JSONB, -- Context data for the event
    created_at         TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (namespace, workflow_id, run_id, event_id)
    -- No foreign key for scalability reasons
);

-- 3. Activity State Management (Retry & Heartbeat)
CREATE TABLE activity_states (
    namespace          VARCHAR(255) NOT NULL DEFAULT 'default',
    workflow_id        VARCHAR(255) NOT NULL,
    run_id             UUID NOT NULL,
    activity_id        VARCHAR(255) NOT NULL,
    status             VARCHAR(50) NOT NULL, -- Scheduled, Started, Completed, Failed
    attempt            INT NOT NULL DEFAULT 1,
    last_heartbeat     TIMESTAMP WITH TIME ZONE,
    last_failure_reason TEXT,
    task_token         BYTEA, 
    PRIMARY KEY (namespace, workflow_id, run_id, activity_id)
    -- No foreign key for scalability reasons
);

-- 4. Timer Registry (Integrated in History Service)
CREATE TABLE timers (
    timer_id           UUID PRIMARY KEY,
    namespace          VARCHAR(255) NOT NULL DEFAULT 'default',
    workflow_id        VARCHAR(255) NOT NULL,
    run_id             UUID NOT NULL,
    fire_time          TIMESTAMP WITH TIME ZONE NOT NULL,
    timer_type         VARCHAR(50) NOT NULL, -- UserTimer, ActivityTimeout, WorkflowTimeout
    task_token         BYTEA, 
    is_fired           BOOLEAN DEFAULT FALSE
    -- No foreign key for scalability reasons
);

-- Critical Index for Timer Polling
CREATE INDEX idx_timers_scanning ON timers(fire_time) WHERE NOT is_fired;
CREATE INDEX idx_wf_status ON workflow_executions(status);

-- +migrate Down
DROP INDEX IF EXISTS idx_wf_status;
DROP INDEX IF EXISTS idx_timers_scanning;
DROP TABLE IF EXISTS timers;
DROP TABLE IF EXISTS activity_states;
DROP TABLE IF EXISTS history_events;
DROP TABLE IF EXISTS workflow_executions;
