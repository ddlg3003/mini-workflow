package service

import (
	"context"
	"encoding/json"
	"time"

	pb "mini-workflow/api"
	"mini-workflow/history/internal/domain"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *historyService) RecordActivityTaskHeartbeat(ctx context.Context, req *pb.RecordActivityTaskHeartbeatRequest) (*pb.RecordActivityTaskHeartbeatResponse, error) {
	if len(req.TaskToken) == 0 {
		return nil, status.Error(codes.InvalidArgument, "task_token is required")
	}

	tok, err := decodeTaskToken(req.TaskToken)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid task token: %v", err)
	}

	runID, err := runIDFromString(tok.RunID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid run_id in token: %v", err)
	}

	exec, err := s.repo.GetWorkflowExecution(ctx, "default", tok.WorkflowID, runID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "workflow execution not found: %v", err)
	}

	cancelRequested := exec.Status != domain.WorkflowStatusRunning

	if !cancelRequested {
		now := time.Now().UTC()
		actState, err := s.repo.GetActivityState(ctx, "default", tok.WorkflowID, runID, tok.ActivityID)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "get activity state: %v", err)
		}

		actState.LastHeartbeat = &now

		if err := s.repo.UpdateWorkflowExecution(ctx, exec, exec.CurrentVersion, nil, []domain.ActivityState{*actState}, nil); err != nil {
			return nil, status.Errorf(codes.Internal, "update heartbeat: %v", err)
		}

		// Push the heartbeat deadline forward in Redis so the timer processor
		// sleeps until the new deadline instead of firing early.
		timeoutTimer, err := s.repo.GetActivityTimeoutTimer(ctx, "default", tok.WorkflowID, runID, tok.ActivityID)
		if err != nil {
			s.log.Warn("get activity timeout timer failed", zap.String("activity_id", tok.ActivityID), zap.Error(err))
		} else if timeoutTimer != nil {
			timeoutSecs := actState.HeartbeatTimeoutSeconds
			if timeoutSecs <= 0 {
				timeoutSecs = 30
			}
			newFireTime := now.Add(time.Duration(timeoutSecs) * time.Second)

			if err := s.repo.UpdateTimerFireTime(ctx, timeoutTimer.TimerID, newFireTime); err != nil {
				s.log.Warn("update timer fire time failed", zap.String("timer_id", timeoutTimer.TimerID.String()), zap.Error(err))
			} else if err := s.timerStore.ScheduleTimer(ctx, timeoutTimer.TimerID.String(), newFireTime.UnixMilli()); err != nil {
				s.log.Warn("reschedule timer in redis failed", zap.String("timer_id", timeoutTimer.TimerID.String()), zap.Error(err))
			}
		}
	}

	return &pb.RecordActivityTaskHeartbeatResponse{CancelRequested: cancelRequested}, nil
}

// activityTimerToken is the JSON shape stored in Timer.TaskToken for ActivityTimeout timers.
type activityTimerToken struct {
	ActivityID              string `json:"activity_id"`
	HeartbeatTimeoutSeconds int    `json:"heartbeat_timeout_seconds"`
}

func decodeActivityTimerToken(raw []byte) (*activityTimerToken, error) {
	var tok activityTimerToken
	if err := json.Unmarshal(raw, &tok); err != nil {
		return nil, err
	}
	return &tok, nil
}
