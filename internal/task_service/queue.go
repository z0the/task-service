package task_service

import (
	"context"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"
)

func newQueue(lg *log.Logger, taskRunnerLimit int) *Queue {
	q := &Queue{
		lg:                  lg,
		taskWorkerLimit:     taskRunnerLimit,
		activeTaskWorkers:   make(map[*Task]context.CancelFunc),
		workerDoneChan:      make(chan *Task, taskRunnerLimit),
		workerInterruptChan: make(chan *Task, taskRunnerLimit),
		passiveTaskList:     make([]*Task, 0, 0),
		latestTask:          new(Task),
		earliestTask:        new(Task),
	}

	q.runTaskDoneHandler()
	q.runTaskInterruptHandler()

	return q
}

type Queue struct {
	sync.Mutex
	id                  string
	lg                  *log.Logger
	taskWorkerLimit     int
	passiveTaskList     []*Task
	activeTaskWorkers   map[*Task]context.CancelFunc
	workerDoneChan      chan *Task
	workerInterruptChan chan *Task
	latestTask          *Task
	earliestTask        *Task
}

func (q *Queue) GetLatestTaskTime() int64 {
	return q.latestTask.Time
}

// AddTask синхронно добавляет новую задачу в обработку
func (q *Queue) AddTask(task *Task) {
	q.Lock()
	defer q.Unlock()

	q.addTask(task)
}

// addTask НЕБЕЗОПАСНО, для корректной работы - перед вызовом необходимо заблокировать очередь
// Добавляет новую задачу в обработку
func (q *Queue) addTask(task *Task) {
	// Если достигнут лимит по активным воркерам
	if len(q.activeTaskWorkers) >= q.taskWorkerLimit {
		// Если новая задача должна запуститься перед самой ранней запущенной,
		// то прерываем воркера с самым большим временем и запускаем новую задачу
		if task.Time <= q.earliestTask.Time {
			if q.latestTask == nil {
				fmt.Println("WARNING nil")
			}
			// Остановка воркера
			q.activeTaskWorkers[q.latestTask]()

			q.startTaskWorker(task)

			// Обновляем время самой поздней и самой ранней запущенной задачи
			q.updateLatestAndEarliestTask()
		} else {
			// Иначе добавляем задачу в пассивный список
			q.passiveTaskList = append(q.passiveTaskList, task)
			// Сразу сортируем, чтобы долго не искать следующую задачу, когда активная очередь освободиться
			sort.Slice(q.passiveTaskList, func(i, j int) bool {
				return q.passiveTaskList[i].Time > q.passiveTaskList[j].Time
			})
		}
	} else {
		if q.latestTask.Time < task.Time {
			q.latestTask = task
		}
		if q.earliestTask.Time > task.Time || q.earliestTask.Time == 0 {
			q.earliestTask = task
		}
		q.startTaskWorker(task)
	}
}

func (q *Queue) startTaskWorker(task *Task) {
	ctx, cancel := context.WithCancel(context.Background())
	q.activeTaskWorkers[task] = cancel

	go func() {
		for {
			select {
			case <-ctx.Done():
				q.workerInterruptChan <- task
				return
			default:
				break
			}

			if time.Now().Unix() == task.Time {
				q.lg.Print(getLogString(task))
				q.workerDoneChan <- task
				return
			}
		}
	}()
}

func (q *Queue) runTaskInterruptHandler() {
	go func() {
		for task := range q.workerInterruptChan {
			q.Lock()

			delete(q.activeTaskWorkers, task)
			q.addTask(task)

			q.Unlock()
		}
	}()
}

func (q *Queue) runTaskDoneHandler() {
	go func() {
		for task := range q.workerDoneChan {
			q.Lock()

			delete(q.activeTaskWorkers, task)

			// Переносим задачи из пассивного списка в активную обработку
			freeWorkersNum := q.taskWorkerLimit - len(q.activeTaskWorkers)
			for i, task := range q.passiveTaskList {
				if freeWorkersNum <= 0 {
					break
				}
				q.addTask(task)
				q.passiveTaskList = q.passiveTaskList[i+1:]
				i--
				freeWorkersNum--
			}

			// Если выполнилась самая последняя или самая первая задача,
			// то обновляем соответсвующие значения
			if task == q.latestTask || task == q.earliestTask {
				q.updateLatestAndEarliestTask()
			}

			q.Unlock()
		}
	}()
}

func (q *Queue) updateLatestAndEarliestTask() {
	q.latestTask = new(Task)
	q.earliestTask = new(Task)
	for task := range q.activeTaskWorkers {
		if task.Time > q.latestTask.Time {
			q.latestTask = task
		}
		if task.Time < q.earliestTask.Time || q.earliestTask.Time == 0 {
			q.earliestTask = task
		}
	}
}

func getLogString(task *Task) string {
	return fmt.Sprintf("Task %s is processed at time: %d\n", task.Action, task.Time)
}
