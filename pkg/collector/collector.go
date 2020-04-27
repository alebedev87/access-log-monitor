package collector

import (
	"fmt"
	"time"

	"httplogmonitor/pkg/config"
	"httplogmonitor/pkg/printer"
)

// Collector stores the summary stats for the given summary interval
type Collector struct {
	config *config.Config
	sum    *Summary
}

// New returns a new instance of Collector
func New(cfg *config.Config) *Collector {
	return &Collector{
		config: cfg,
		sum:    NewSummary(cfg.TopSectionNum),
	}
}

// Start collects the log message statistics (most hitted sections and some interesting info)
// and sends it to the printer every summary interval
func (c *Collector) Start(logCh <-chan string, printCh chan<- printer.Formatter) {
	tick := time.NewTicker(time.Duration(c.config.SummaryIntervalSec) * time.Second)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			// time to print the summary
			c.sum.CalcTraffic(c.config.SummaryIntervalSec)
			printCh <- *c.sum
			c.sum = NewSummary(c.config.TopSectionNum)
		case l := <-logCh:
			// transform raw log entries into log messages
			msg, err := NewLogMessageFromLogEntry(l)
			if err != nil {
				printCh <- printer.NewErrorMessage(fmt.Sprintf("Failed to parse log entry: %q. Error: %s", l, err))
				break
			}
			// add messages to the summary
			c.sum.Add(msg)
		}
	}
}
