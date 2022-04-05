package model

import "time"

type Item struct {
	UID       string
	URL       string
	CrawledAt time.Time
}
