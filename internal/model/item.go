package model

import "time"

type Item struct {
	UID       string
	CrawledAt time.Time
}
