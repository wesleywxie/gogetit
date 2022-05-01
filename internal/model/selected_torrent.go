package model

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"time"
)

type SelectedTorrent struct {
	gorm.Model
	VideoID     uint
	TorrentID   uint
	UID         string
	MagnetLink  string
	FileSize    int
	FileNum     int
	Status      Status
	PublishedAt time.Time
}

func AddSelectedTorrent(v *Video, t *Torrent) (selectedTorrent SelectedTorrent, err error) {
	if err := db.Where("UID=?", v.UID).Find(&selectedTorrent).Error; err != nil {
		if err.Error() == "record not found" {
			selectedTorrent.VideoID = t.VideoID
			selectedTorrent.TorrentID = t.ID
			selectedTorrent.UID = v.UID
			selectedTorrent.MagnetLink = fmt.Sprintf("%v&dn=%s_%d_files_%.2fGB", t.MagnetLink, v.UID, t.FileNum, float64(t.FileSize)/1024.0)
			selectedTorrent.FileSize = t.FileSize
			selectedTorrent.FileNum = t.FileNum
			selectedTorrent.PublishedAt = t.PublishedAt
			selectedTorrent.CreatedAt = time.Now()
			selectedTorrent.UpdatedAt = time.Now()
			selectedTorrent.Status = INIT
			if db.Create(&selectedTorrent).Error == nil {
				return selectedTorrent, nil
			}
		}
	}
	return
}

func FindAndUpdateSelectedTorrents(update bool) []SelectedTorrent {
	var torrents []SelectedTorrent
	db.Where("status=?", INIT).Order("created_at desc").Find(&torrents)
	if update {
		db.Table("selected_torrents").Where("status = ?", INIT).Updates(map[string]interface{}{"status": COMPLETED, "updated_at": time.Now()})
	}
	return torrents
}
