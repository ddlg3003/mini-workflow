package service

import (
	"testing"

	"mini-workflow/frontend/internal/mocks"
	"mini-workflow/frontend/internal/router"

	"go.uber.org/zap"
)

func newTestService(t *testing.T) (*FrontendService, *mocks.MockHistoryClient, *mocks.MockMatchingClient) {
	t.Helper()
	hist := mocks.NewMockHistoryClient(t)
	match := mocks.NewMockMatchingClient(t)
	r := router.New(hist)
	svc := New(r, match, zap.NewNop())
	return svc, hist, match
}
