package log

import (
	"github.com/gocolly/colly/v2/debug"
	"go.uber.org/zap"
)

type Debugger struct {
}

// Init initializes the LogDebugger
func (l *Debugger) Init() error {
	return nil
}

// Event receives Collector events and prints them to STDERR
func (l *Debugger) Event(e *debug.Event) {
	zap.S().Debugf("%d [%6d - %s] %q", e.CollectorID, e.RequestID, e.Type, e.Values)
}
