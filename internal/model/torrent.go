package model

import (
	"github.com/jinzhu/gorm"
	"time"
)

type Torrent struct {
	gorm.Model
	VideoID     uint
	Magnet      string
	Size        string
	Num         string
	PublishedAt string
}

func AddTorrent(t *Torrent) (torrent Torrent, err error) {
	if err := db.Where("magnet=?", t.Magnet).Find(&torrent).Error; err != nil {
		if err.Error() == "record not found" {
			torrent.VideoID = t.VideoID
			torrent.Magnet = t.Magnet
			torrent.Size = t.Size
			torrent.Num = t.Num
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
