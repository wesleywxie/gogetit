package model

import "github.com/jinzhu/gorm"

type Torrent struct {
	gorm.Model
	VideoID     uint
	Magnets     string
	Size        string
	Num         string
	PublishedAt string
}
