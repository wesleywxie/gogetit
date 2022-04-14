package util

import (
	"go.uber.org/zap"
	"time"
)

func ParseTime(str string) (datetime time.Time, err error) {
	layout := "2006-01-02"
	datetime, err = time.Parse(layout, str)

	if err != nil {
		zap.S().Errorw("Error while parsing time",
			"timeStr", str,
			"error", err,
		)
		datetime = time.Now()
	}
	return
}
