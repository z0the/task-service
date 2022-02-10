package main

import (
	"log"
	"os"
	"time"

	"kroog-test/internal/task_service"
	"kroog-test/internal/transport/http"
)

func init() {
	time.Local = time.UTC
}

func main() {
	logger := log.New(os.Stdout, "", log.Ldate|log.Lshortfile)

	logger.Println("test")

	svc := task_service.NewService(logger)

	controller := http.NewController(logger, svc)

	err := controller.Run()
	if err != nil {
		logger.Fatal("Controller has stopped with err: ", err)
	}
}
