package task_service

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

const defaultRunnerLimitPerQueue = 100

type taskResult struct {
	text []string
}

func (tr *taskResult) Write(p []byte) (n int, err error) {
	tr.text = append(tr.text, string(p))
	return len(p), err
}

func TestQueueTestSuite(t *testing.T) {
	suite.Run(t, new(QueueTestSuite))
}

type QueueTestSuite struct {
	suite.Suite
}

func (s *QueueTestSuite) SetupSuite() {}

func (s *QueueTestSuite) SetupTest() {}

func (s *QueueTestSuite) TestExecuteOneTask_Success() {
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
	s.Equal(*task, *queue.latestTask)
	s.Equal(*task, *queue.earliestTask)
	s.Equal(getLogString(task), tr.text[0])
}

func (s *QueueTestSuite) TestExecute100DifferentTimeTask_Success() {
	tr := new(taskResult)
	lg := log.New(tr, "", 0)
	queue := newQueue(lg, defaultRunnerLimitPerQueue)
	expectedTextRes := make([]string, 0, 100)
	for i := 0; i < 100; i++ {
		task := &Task{
			QueueID: "1",
			Action:  fmt.Sprintf("action #%d", i),
			Time:    time.Now().Add(time.Second * time.Duration(i)).Unix(),
		}
		queue.AddTask(task)
		expectedTextRes = append(expectedTextRes, getLogString(task))
	}

	time.Sleep(time.Second * 101)

	// Проверяем, что все ресурсы утилизированны корректно
	// и что в логах есть записи обо всех задачах
	s.Len(queue.activeTaskWorkers, 0)
	s.Len(queue.passiveTaskList, 0)
	s.Len(tr.text, 100)
	s.ElementsMatch(expectedTextRes, tr.text)
}

func (s *QueueTestSuite) TestExecute100OneTimeTask_Success() {
	tr := new(taskResult)
	lg := log.New(tr, "", 0)
	queue := newQueue(lg, defaultRunnerLimitPerQueue)

	expectedTextRes := make([]string, 0, 100)
	for i := 0; i < 100; i++ {
		task := &Task{
			QueueID: "1",
			Action:  fmt.Sprintf("action #%d", i),
			Time:    time.Now().Unix(),
		}
		queue.AddTask(task)
		expectedTextRes = append(expectedTextRes, getLogString(task))
	}

	time.Sleep(time.Second * 2)

	// Проверяем, что все ресурсы утилизированны корректно
	// и что в логах есть записи обо всех задачах
	s.Len(queue.activeTaskWorkers, 0)
	s.Len(queue.passiveTaskList, 0)
	s.Len(tr.text, 100)
	s.ElementsMatch(expectedTextRes, tr.text)
}

func (s *QueueTestSuite) TestExecute1000OneTimeTask_Success() {
	tr := new(taskResult)
	lg := log.New(tr, "", 0)
	queue := newQueue(lg, defaultRunnerLimitPerQueue)

	expectedTextRes := make([]string, 0, 1000)
	for i := 0; i < 1000; i++ {
		task := &Task{
			QueueID: "1",
			Action:  fmt.Sprintf("action #%d", i),
			Time:    time.Now().Unix(),
		}
		queue.AddTask(task)
		expectedTextRes = append(expectedTextRes, getLogString(task))
	}

	// Спим, чтобы все задачи успели отработать
	time.Sleep(time.Second * 2)

	// Проверяем, что все ресурсы утилизированны корректно

	// и что в логах есть записи обо всех задачах
	s.Len(queue.activeTaskWorkers, 0)
	s.Len(queue.passiveTaskList, 0)
	s.Len(tr.text, 1000)
	s.ElementsMatch(expectedTextRes, tr.text)
}

func (s *QueueTestSuite) TestActiveAndPassiveTask_Success() {
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
	s.Len(queue.activeTaskWorkers, defaultRunnerLimitPerQueue)
	s.Len(queue.passiveTaskList, 1000-defaultRunnerLimitPerQueue)
	s.Len(tr.text, 0)
}

func (s *QueueTestSuite) TestTaskTimePriority_Success() {
	// Тест не имеет смысла, если ограничение на воркеров менее 2
	if defaultRunnerLimitPerQueue < 2 {
		return
	}

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
	s.Len(queue.passiveTaskList, 100-defaultRunnerLimitPerQueue+1)
	s.Equal(lastTask, queue.passiveTaskList[len(queue.passiveTaskList)-1])
	s.Len(queue.activeTaskWorkers, defaultRunnerLimitPerQueue)
	_, isNewTaskSwappedWithOld := queue.activeTaskWorkers[taskWithEarlierTime]
	s.True(isNewTaskSwappedWithOld)
}
