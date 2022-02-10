package task_service

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type taskResult struct {
	text []string
}

func (tr *taskResult) Write(p []byte) (n int, err error) {
	tr.text = append(tr.text, string(p))
	return len(p), err
}

func TestQueueSuiteTestSuite(t *testing.T) {
	suite.Run(t, new(QueueSuite))
}

type QueueSuite struct {
	suite.Suite
	lg *log.Logger
	tr *taskResult
}

func (s *QueueSuite) SetupSuite() {
	s.tr = new(taskResult)
	s.lg = log.New(s.tr, "", 0)
}

func (s *QueueSuite) SetupTest() {
	*s.tr = taskResult{}
}

func (s *QueueSuite) TestExecuteOneTask_Success() {
	tr := new(taskResult)
	lg := log.New(tr, "", 0)
	queue := newQueue(lg, defaultRunnerLimitPerQueue)
	task := &Task{
		QueueID: "1",
		Action:  "action",
		Time:    time.Now().Add(time.Second * 3).Unix(),
	}
	queue.AddTask(task)

	time.Sleep(time.Second * 4)

	// Проверяем, что все ресурсы утилизированны корректно
	// и что в логах отображается информация о задаче
	s.Len(queue.activeTaskWorkers, 0)
	s.Len(queue.passiveTaskList, 0)
	s.Equal(Task{}, *queue.latestTask)
	s.Equal(Task{}, *queue.earliestTask)
	s.Equal(getLogString(task), tr.text[0])
}

func (s *QueueSuite) TestExecute100OneTimeTask_Success() {
	tr := new(taskResult)
	lg := log.New(tr, "", 0)
	queue := newQueue(lg, defaultRunnerLimitPerQueue)
	for i := 0; i < 100; i++ {
		task := &Task{
			QueueID: "1",
			Action:  fmt.Sprintf("action #%d", i),
			Time:    time.Now().Unix(),
		}
		queue.AddTask(task)
	}

	time.Sleep(time.Second * 2)

	// Проверяем, что все ресурсы утилизированны корректно
	// и что в логах есть записи обо всех задачах
	s.Len(queue.activeTaskWorkers, 0)
	s.Len(queue.passiveTaskList, 0)
	s.Len(tr.text, 100)
}

func (s *QueueSuite) TestExecute1000OneTimeTask_Success() {
	tr := new(taskResult)
	lg := log.New(tr, "", 0)
	queue := newQueue(lg, defaultRunnerLimitPerQueue)
	for i := 0; i < 1000; i++ {
		task := &Task{
			QueueID: "1",
			Action:  fmt.Sprintf("action #%d", i),
			Time:    time.Now().Unix(),
		}
		queue.AddTask(task)
	}

	// Спим, чтобы все задачи успели отработать
	time.Sleep(time.Second * 2)

	// Проверяем, что все ресурсы утилизированны корректно

	// и что в логах есть записи обо всех задачах
	s.Len(queue.activeTaskWorkers, 0)
	s.Len(queue.passiveTaskList, 0)
	s.Len(tr.text, 1000)
}

func (s *QueueSuite) TestActiveAndPassiveTask_Success() {
	tr := new(taskResult)
	lg := log.New(tr, "", 0)
	queue := newQueue(lg, defaultRunnerLimitPerQueue)
	for i := 0; i < 1000; i++ {
		task := &Task{
			QueueID: "1",
			Action:  fmt.Sprintf("action #%d", i),
			Time:    time.Now().Add(time.Second * time.Duration(i+2)).Unix(),
		}
		queue.AddTask(task)
	}

	// Проверяем, что при одновременном добавлении 1000 задач
	// количество активных и пассивных задач - корректно
	s.Len(queue.activeTaskWorkers, 100)
	s.Len(queue.passiveTaskList, 900)
	s.Len(tr.text, 0)
}

func (s *QueueSuite) TestTaskTimePriority_Success() {
	tr := new(taskResult)
	lg := log.New(tr, "", 0)
	queue := newQueue(lg, defaultRunnerLimitPerQueue)
	var lastTask *Task
	for i := 0; i < 100; i++ {
		task := &Task{
			QueueID: "1",
			Action:  fmt.Sprintf("action #%d", i),
			Time:    time.Now().Add(time.Minute * time.Duration(i+6)).Unix(),
		}
		queue.AddTask(task)
		if i == 99 {
			lastTask = task
		}
	}

	taskWithEarlierTime := &Task{
		QueueID: "1",
		Action:  "Priority",
		Time:    time.Now().Add(time.Minute).Unix(),
	}
	queue.AddTask(taskWithEarlierTime)

	// Задержка, чтобы обработчик в очереди успел отработать
	time.Sleep(time.Second * 5)

	// Проверяем, что предыдущая последняя задача переместилась в пассивный список
	// А задача с более ранним временем попала в активную очередь
	s.Len(queue.passiveTaskList, 1)
	s.Equal(queue.passiveTaskList[0], lastTask)
	s.Len(queue.activeTaskWorkers, 100)
	_, isNewTaskSwappedWithOld := queue.activeTaskWorkers[taskWithEarlierTime]
	s.True(isNewTaskSwappedWithOld)
}
