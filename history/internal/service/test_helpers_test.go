package service

import (
	"encoding/json"

	"mini-workflow/history/internal/domain"

	pb "mini-workflow/api"

	"github.com/google/uuid"
)

func runningExec(workflowID string) *domain.WorkflowExecution {
	return &domain.WorkflowExecution{
		Namespace:      "default",
		WorkflowID:     workflowID,
		RunID:          uuid.New(),
		TaskQueue:      "q",
		Status:         domain.WorkflowStatusRunning,
		CurrentVersion: 1,
		NextEventID:    2,
	}
}

func buildActivityState(exec *domain.WorkflowExecution, activityID string, attempt int) *domain.ActivityState {
	return &domain.ActivityState{
		Namespace:  exec.Namespace,
		WorkflowID: exec.WorkflowID,
		RunID:      exec.RunID,
		ActivityID: activityID,
		Status:     domain.ActivityStatusStarted,
		Attempt:    attempt,
	}
}

func scheduleActivityCmd(activityID, taskQueue string) *pb.Command {
	attrs, _ := json.Marshal(map[string]any{
		"activity_id":   activityID,
		"activity_type": "MyActivity",
		"input":         []byte(`"input"`),
		"task_queue":    taskQueue,
	})
	return &pb.Command{CommandType: "ScheduleActivityTask", Attributes: attrs}
}

func completeWorkflowCmd(result []byte) *pb.Command {
	attrs, _ := json.Marshal(map[string]any{"result": result})
	return &pb.Command{CommandType: "CompleteWorkflowExecution", Attributes: attrs}
}

func startTimerCmd(fireAfter int64) *pb.Command {
	attrs, _ := json.Marshal(map[string]any{"timer_id": "t1", "fire_after_seconds": fireAfter})
	return &pb.Command{CommandType: "StartTimer", Attributes: attrs}
}
