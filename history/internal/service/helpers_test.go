package service

import (
	"testing"

	"mini-workflow/history/internal/mocks"

	"go.uber.org/zap"
)

func newTestService(t *testing.T) (*historyService, *mocks.MockExecutionRepository, *mocks.MockMatchingClient, *mocks.MockTimerStore) {
	t.Helper()
	repo := mocks.NewMockExecutionRepository(t)
	matching := mocks.NewMockMatchingClient(t)
	ts := mocks.NewMockTimerStore(t)
	svc := New(repo, matching, ts, zap.NewNop())
	return svc, repo, matching, ts
}
