-- +migrate Up
ALTER TABLE activity_states ADD COLUMN task_queue VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE activity_states ADD COLUMN activity_type VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE activity_states ADD COLUMN input JSONB;

-- +migrate Down
ALTER TABLE activity_states DROP COLUMN IF EXISTS input;
ALTER TABLE activity_states DROP COLUMN IF EXISTS activity_type;
ALTER TABLE activity_states DROP COLUMN IF EXISTS task_queue;
