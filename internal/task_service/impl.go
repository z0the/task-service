package task_service

import (
	"context"
	"log"
	"time"

	"kroog-test/config"
)

func NewService(cfg config.Config, lg *log.Logger) *service {
	s := &service{
		cfg:          cfg,
		lg:           lg,
		taskQueueMap: make(map[string]*Queue),
	}

	s.startQueueCleanupWorker()

	return s
}

type service struct {
	cfg          config.Config
	lg           *log.Logger
	taskQueueMap map[string]*Queue
}

func (s service) ScheduleTask(_ context.Context, task *Task) error {

	taskQueue, ok := s.taskQueueMap[task.QueueID]
	if !ok {
		taskQueue = newQueue(s.lg, s.cfg.WorkerLimitPerQueue)
		s.taskQueueMap[task.QueueID] = taskQueue
	}
	taskQueue.AddTask(task)

	return nil
}

// startQueueCleanupWorker удаляет очереди, если они не используются час и более
func (s service) startQueueCleanupWorker() {
	go func() {
		for queueID, queue := range s.taskQueueMap {
			if time.Now().Unix()-queue.GetLatestTaskTime() >= int64(time.Hour.Seconds()) {
				delete(s.taskQueueMap, queueID)
				s.lg.Printf("queue with ID: %s is utilized\n", queueID)
			}
		}
		time.Sleep(time.Hour)
	}()
}
