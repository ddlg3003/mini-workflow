package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	pb "mini-workflow/api"
	"mini-workflow/history/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRecordWorkflowTaskStarted(t *testing.T) {
	svc, repo, _, timerStore := newTestService(t)

	wfID := "wf-1"
	runID := uuid.New()

	request := &pb.RecordWorkflowTaskStartedRequest{
		WorkflowId: wfID,
		RunId:      runID.String(),
	}

	exec := &domain.WorkflowExecution{
		Namespace:      "default",
		WorkflowID:     wfID,
		RunID:          runID,
		Status:         domain.WorkflowStatusRunning,
		CurrentVersion: 1,
		NextEventID:    5,
	}

	repo.EXPECT().GetWorkflowExecution(mock.Anything, "default", wfID, runID).Return(exec, nil).Once()

	var timersCaptured []domain.Timer
	repo.EXPECT().UpdateWorkflowExecution(mock.Anything, mock.Anything, int64(1), mock.Anything, []domain.ActivityState(nil), mock.Anything).
		Run(func(ctx context.Context, exec *domain.WorkflowExecution, expectedVersion int64, events []domain.HistoryEvent, activitiesToUpsert []domain.ActivityState, timersToInsert []domain.Timer) {
			timersCaptured = timersToInsert
		}).
		Return(nil).Once()

	timerStore.EXPECT().ScheduleTimer(mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	res, err := svc.RecordWorkflowTaskStarted(context.Background(), request)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	assert.Len(t, timersCaptured, 1)
	assert.Equal(t, domain.TimerTypeWorkflowTaskTimeout, timersCaptured[0].TimerType)

	// Ensure the token has the correct event_id
	var tok struct {
		EventID int64 `json:"event_id"`
	}
	err = json.Unmarshal(timersCaptured[0].TaskToken, &tok)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), tok.EventID)

	// Ensure timer fires roughly 10s from now
	expectedFireTime := time.Now().Add(10 * time.Second)
	assert.WithinDuration(t, expectedFireTime, timersCaptured[0].FireTime, 1*time.Second)
}
