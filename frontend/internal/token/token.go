package token

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type TaskToken struct {
	WorkflowID string `json:"wf_id"`
	RunID      string `json:"run_id"`
	ActivityID string `json:"act_id,omitempty"`
	Attempt    int32  `json:"attempt,omitempty"`
}

func Encode(t TaskToken) ([]byte, error) {
	b, err := json.Marshal(t)
	if err != nil {
		return nil, fmt.Errorf("marshal token: %w", err)
	}
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(b)))
	base64.StdEncoding.Encode(dst, b)
	return dst, nil
}

func Decode(raw []byte) (TaskToken, error) {
	dst := make([]byte, base64.StdEncoding.DecodedLen(len(raw)))
	n, err := base64.StdEncoding.Decode(dst, raw)
	if err != nil {
		return TaskToken{}, fmt.Errorf("decode token: %w", err)
	}
	var t TaskToken
	if err := json.Unmarshal(dst[:n], &t); err != nil {
		return TaskToken{}, fmt.Errorf("unmarshal token: %w", err)
	}
	return t, nil
}
