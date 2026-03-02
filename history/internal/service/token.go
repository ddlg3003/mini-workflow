package service

import (
	"encoding/json"
	"fmt"
)

type taskToken struct {
	WorkflowID string `json:"workflow_id"`
	RunID      string `json:"run_id"`
	ActivityID string `json:"activity_id"`
	Attempt    int    `json:"attempt"`
}

func decodeTaskToken(raw []byte) (*taskToken, error) {
	var tok taskToken
	if err := json.Unmarshal(raw, &tok); err != nil {
		return nil, fmt.Errorf("unmarshal task token: %w", err)
	}
	if tok.WorkflowID == "" || tok.RunID == "" {
		return nil, fmt.Errorf("task token missing workflow_id or run_id")
	}
	return &tok, nil
}
