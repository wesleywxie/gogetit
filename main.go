package main

import (
	"github.com/wesleywxie/gogetit/internal/model"
	"github.com/wesleywxie/gogetit/internal/task"
	"go.uber.org/zap"
)

func init() {
	zap.S().Debug("Initialization main module...")
}

func main() {
	model.InitDB()
	defer model.Disconnect()

	task.StartTasks()
}
