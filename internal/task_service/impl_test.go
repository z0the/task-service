package task_service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"kroog-test/config"
)

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}

type ServiceTestSuite struct {
	suite.Suite
	tr *taskResult
	*service
}

func (s *ServiceTestSuite) SetupSuite() {
	cfg := config.GetConfig()

	s.tr = new(taskResult)
	lg := log.New(s.tr, "", 0)
	s.service = NewService(cfg, lg)
}

func (s *ServiceTestSuite) SetupTest() {
	*s.tr = taskResult{}
}

func (s *ServiceTestSuite) TestScheduleOneTaskInOneQueue_Success() {
	task := &Task{
		Time:    time.Now().Unix(),
		Action:  "one task",
		QueueID: "1",
	}

	s.NoError(s.service.ScheduleTask(context.Background(), task))

	time.Sleep(time.Second)

	s.Len(s.tr.text, 1)
	s.Equal(getLogString(task), s.tr.text[0])
}

func (s *ServiceTestSuite) TestSchedule10TasksInOneQueue_Success() {

	expectedTextRes := make([]string, 0, 10)
	for i := 0; i < 10; i++ {
		task := &Task{
			Action:  fmt.Sprintf("task #%d", i),
			QueueID: "1",
			Time:    time.Now().Add(time.Second * time.Duration(i)).Unix(),
		}
		s.NoError(s.service.ScheduleTask(context.Background(), task))
		expectedTextRes = append(expectedTextRes, getLogString(task))
	}

	time.Sleep(time.Second * 11)

	s.Len(s.tr.text, 10)
	s.ElementsMatch(expectedTextRes, s.tr.text)
}

func (s *ServiceTestSuite) TestSchedule10TasksIn10Queues_Success() {

	expectedTextRes := make([]string, 0, 100)
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			task := &Task{
				Action:  fmt.Sprintf("task #%d", i),
				QueueID: strconv.Itoa(j),
				Time:    time.Now().Add(time.Second * time.Duration(i)).Unix(),
			}
			s.NoError(s.service.ScheduleTask(context.Background(), task))
			expectedTextRes = append(expectedTextRes, getLogString(task))
		}
	}

	time.Sleep(time.Second * 10)

	// За десять секунд должно выполнится 100 задач, т.к. запущенно 10 очередей
	s.Len(s.tr.text, 100)
	s.ElementsMatch(expectedTextRes, s.tr.text)
}

func (s *ServiceTestSuite) TestSchedule101TasksIn10Queues_Success() {

	expectedTextRes := make([]string, 0, 1010)
	for i := 0; i < 101; i++ {
		for j := 0; j < 10; j++ {
			task := &Task{
				Action:  fmt.Sprintf("task #%d", i),
				QueueID: strconv.Itoa(j),
				Time:    time.Now().Add(time.Second * time.Duration(i)).Unix(),
			}
			s.NoError(s.service.ScheduleTask(context.Background(), task))
			expectedTextRes = append(expectedTextRes, getLogString(task))
		}
	}

	time.Sleep(time.Second * 101)

	s.Len(s.tr.text, 1010)
	s.ElementsMatch(expectedTextRes, s.tr.text)
}
