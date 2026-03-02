package service

import (
	"mini-workflow/history/internal/ports"

	"go.uber.org/zap"
)

type historyService struct {
	repo       ports.ExecutionRepository
	matching   ports.MatchingClient
	timerStore ports.TimerStore
	log        *zap.Logger
}

func New(repo ports.ExecutionRepository, matching ports.MatchingClient, timerStore ports.TimerStore, log *zap.Logger) *historyService {
	return &historyService{
		repo:       repo,
		matching:   matching,
		timerStore: timerStore,
		log:        log,
	}
}
