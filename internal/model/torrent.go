package model

import (
	"github.com/jinzhu/gorm"
	"time"
)

type Torrent struct {
	gorm.Model
	VideoID     uint
	UID         string
	MagnetLink  string
	FileSize    int
	FileNum     int
	PublishedAt time.Time
}

func PickTop(v *Video) Torrent {
	var torrent Torrent
	db.Where("video_id=? AND file_num > 0 AND file_num < 5 AND published_at > date('now', 'start of day','-1 months') ", v.ID).
		Order("file_size").
		First(&torrent)
	return torrent
}

func AddTorrent(t *Torrent) (torrent Torrent, err error) {
	if err := db.Where("magnet_link=?", t.MagnetLink).Find(&torrent).Error; err != nil {
		if err.Error() == "record not found" {
			torrent.VideoID = t.VideoID
			torrent.UID = t.UID
			torrent.MagnetLink = t.MagnetLink
			torrent.FileSize = t.FileSize
			torrent.FileNum = t.FileNum
			torrent.PublishedAt = t.PublishedAt
			torrent.CreatedAt = time.Now()
			torrent.UpdatedAt = time.Now()
			if db.Create(&torrent).Error == nil {
				return torrent, nil
			}
		}
	}
	return
}
