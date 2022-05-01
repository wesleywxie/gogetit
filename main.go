package main

import (
	"github.com/wesleywxie/gogetit/internal/model"
	"github.com/wesleywxie/gogetit/internal/task"
	server "github.com/wesleywxie/gogetit/internal/web"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	zap.S().Debug("Initialization main module...")
}

func main() {
	model.InitDB()
	task.StartTasks()

	go handleSignal()
	server.Start(31065)
}

func handleSignal() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	<-c

	task.StopTasks()
	model.Disconnect()
	os.Exit(0)
}
