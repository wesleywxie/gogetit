package model

import (
	"github.com/jinzhu/gorm"
	"time"
)

type Video struct {
	gorm.Model
	UID         string
	Duration    string
	Director    string
	Publisher   string
	Series      string
	Categories  string
	Actors      string
	Torrents    []Torrent
	PublishedAt string
	Source      string
	Status      Status
}

func ExistsVideo(UID string) bool {
	var video Video
	err := db.Where("uid=?", UID).First(&video).Error

	if (err != nil && err.Error() == "record not found") || video.UID == "" {
		return false
	}

	// If video already processed, no need further process
	if video.Status == INIT {
		return false
	}

	return true
}

func UpdateStatus(v *Video, newStatus Status) {
	v.Status = newStatus
	db.Save(v)
}

func FindVideosByStatus(status Status) []Video {
	var videos []Video
	db.Where("status=?", status).Find(&videos)

	return videos
}

func AddVideo(v *Video) (video Video, err error) {
	if err := db.Where("uid=?", v.UID).Find(&video).Error; err != nil {
		if err.Error() == "record not found" {
			video.UID = v.UID
			video.PublishedAt = v.PublishedAt
			video.Duration = v.Duration
			video.Director = v.Director
			video.Publisher = v.Publisher
			video.Series = v.Series
			video.Categories = v.Categories
			video.Actors = v.Actors
			video.Source = v.Source
			video.CreatedAt = time.Now()
			video.UpdatedAt = time.Now()
			video.Status = INIT
			if db.Create(&video).Error == nil {
				return video, nil
			}
		}
	}
	return
}
