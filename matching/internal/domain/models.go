package domain

// WorkflowTask is the JSON payload pushed to and popped from the workflow Redis queue.
type WorkflowTask struct {
	WorkflowID   string `json:"workflow_id"`
	RunID        string `json:"run_id"`
	WorkflowType string `json:"workflow_type"`
	TaskQueue    string `json:"task_queue"`
}

// ActivityTask is the JSON payload pushed to and popped from the activity Redis queue.
type ActivityTask struct {
	WorkflowID              string `json:"workflow_id"`
	RunID                   string `json:"run_id"`
	ActivityID              string `json:"activity_id"`
	ActivityType            string `json:"activity_type"`
	Input                   []byte `json:"input"`
	TaskQueue               string `json:"task_queue"`
	HeartbeatTimeoutSeconds int32  `json:"heartbeat_timeout_seconds"`
}
