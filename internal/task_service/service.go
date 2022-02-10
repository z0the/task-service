package task_service

import "context"

type Service interface {
	ScheduleTask(ctx context.Context, task *Task) error
}
