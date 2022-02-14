package main

import (
	"log"
	"os"
	"time"

	"kroog-test/config"
	"kroog-test/internal/task_service"
	"kroog-test/internal/transport/http"
)

func init() {
	time.Local = time.UTC
}

func main() {
	cfg := config.GetConfig(".env")

	logger := log.New(os.Stdout, "", log.Ldate|log.Lshortfile)

	svc := task_service.NewService(cfg, logger)

	controller := http.NewController(logger, svc)

	err := controller.Run()
	if err != nil {
		logger.Fatal("Controller has stopped with err: ", err)
	}
}
