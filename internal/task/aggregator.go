package task

import (
	"github.com/wesleywxie/gogetit/internal/model"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"strings"
	"time"
)

func init() {
	task := NewAggregator()
	registerTask(task)
}

// Aggregator 种子汇总
type Aggregator struct {
	isStop atomic.Bool
}

// NewAggregator 构造 Aggregator
func NewAggregator() *Aggregator {
	task := &Aggregator{}
	task.isStop.Store(false)
	return task
}

// Name 任务名称
func (t *Aggregator) Name() string {
	return "Aggregator"
}

func (t *Aggregator) IsStopped() bool {
	return t.isStop.Load()
}

// Stop 停止
func (t *Aggregator) Stop() {
	t.isStop.Store(true)
}

// Start 启动
func (t *Aggregator) Start() {
	t.isStop.Store(false)
	go func() {
		taskCompletedCount := 0
		for taskCompletedCount < (len(taskList) - 1) {
			// sleep for 1 second to allow all other tasks to start running
			time.Sleep(1 * time.Second)
			taskCompletedCount = 0
			for _, task := range taskList {
				if task.Name() != t.Name() && task.IsStopped() {
					taskCompletedCount++
				}
			}
		}

		// If all other tasks are completed
		processVideoAndTorrent()
		t.Stop()
	}()
}

func processVideoAndTorrent() {
	zap.S().Info("Starting video and torrent process... ")

	videos := model.FindVideosByStatus(model.INIT)
	for _, video := range videos {
		if strings.Contains(video.Categories, "VR,") {
			model.UpdateStatus(&video, model.SKIPPED)
			continue
		}
		torrent := model.PickTop(&video)
		if torrent.ID > 0 {
			selectedTorrent, _ := model.AddSelectedTorrent(&video, &torrent)
			zap.S().Infof("Found top pick torrent for %v with magnet link %v",
				selectedTorrent.UID, selectedTorrent.MagnetLink)
		}
		model.UpdateStatus(&video, model.COMPLETED)
	}
}
