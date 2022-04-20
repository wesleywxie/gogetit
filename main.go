package main

import (
	"github.com/wesleywxie/gogetit/internal/model"
	"github.com/wesleywxie/gogetit/internal/task"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func init() {
	zap.S().Debug("Initialization main module...")
}

func main() {
	model.InitDB()
	defer model.Disconnect()
	task.StartTasks()

	go handleSignal()
	keepRunning()
}

func handleSignal() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	<-c

	task.StopTasks()
	model.Disconnect()
	os.Exit(0)
}

func keepRunning() {
	for {
		time.Sleep(5 * time.Minute)
	}
}
