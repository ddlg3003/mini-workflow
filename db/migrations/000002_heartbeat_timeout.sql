-- +migrate Up
ALTER TABLE activity_states ADD COLUMN heartbeat_timeout_seconds INT NOT NULL DEFAULT 0;
CREATE INDEX idx_timers_activity ON timers(workflow_id, run_id, timer_type) WHERE NOT is_fired;

-- +migrate Down
DROP INDEX IF EXISTS idx_timers_activity;
ALTER TABLE activity_states DROP COLUMN IF EXISTS heartbeat_timeout_seconds;
