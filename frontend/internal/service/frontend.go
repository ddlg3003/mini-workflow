package service

import (
	"context"
	"errors"

	"mini-workflow/frontend/internal/ports"
	"mini-workflow/frontend/internal/router"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FrontendService struct {
	router   *router.Router
	matching ports.MatchingClient
	log      *zap.Logger
}

func New(r *router.Router, matching ports.MatchingClient, log *zap.Logger) *FrontendService {
	return &FrontendService{router: r, matching: matching, log: log}
}

func (s *FrontendService) history(workflowID string) ports.HistoryClient {
	return s.router.GetHistoryClient(workflowID)
}

func mapDownstreamError(err error) error {
	s, ok := status.FromError(err)
	if !ok {
		return status.Errorf(codes.Unavailable, "downstream error: %v", err)
	}
	switch s.Code() {
	case codes.NotFound:
		return err
	case codes.Unavailable, codes.DeadlineExceeded:
		return status.Errorf(codes.Unavailable, "downstream unavailable: %v", s.Message())
	default:
		return err
	}
}

func isTimeout(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	s, ok := status.FromError(err)
	if ok && s.Code() == codes.DeadlineExceeded {
		return true
	}
	return false
}
