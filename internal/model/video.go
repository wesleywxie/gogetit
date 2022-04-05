package model

import "github.com/jinzhu/gorm"

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
}
