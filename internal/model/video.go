package model

import "time"

type Video struct {
	UID         string
	Duration    string
	Director    string
	Publisher   string
	Series      string
	Category    []string
	Actress     []string
	Actors      []string
	Torrents    []Torrent
	PublishedAt string
	Source      string
	CrawledAt   time.Time
}
