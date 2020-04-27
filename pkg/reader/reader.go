package reader

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	alert "httplogmonitor/pkg/alertmanager"
	"httplogmonitor/pkg/config"
	"httplogmonitor/pkg/printer"
)

// Reader implement a tailer (tail -f) on a given file
type Reader struct {
	config *config.Config
}

// New returns an instance of Reader
func New(cfg *config.Config) *Reader {
	return &Reader{
		config: cfg,
	}
}

// Start sends raw log entries to logCh, counter metric to metCh and errors to printCh
// gracefully stops closing the log file passed through the configuration
func (r *Reader) Start(ctx context.Context, logCh chan<- string, metCh chan<- alert.Metric, printCh chan<- printer.Formatter, wg *sync.WaitGroup) {
	defer wg.Done()

	f, err := os.Open(r.config.LogFilePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// put the offset to the end of the file
	_, err = f.Seek(0, 2)
	if err != nil {
		panic(err)
	}

	reader := bufio.NewReader(f)

	tick := time.NewTicker(time.Duration(r.config.PollIntervalSec) * time.Second)
	defer tick.Stop()

loop:
	for {
		select {
		case t := <-tick.C:
			// number of log entries read for each tick
			logCnt := 0
			for {
				msg, err := reader.ReadBytes('\n')
				if err != nil {
					if err != io.EOF {
						printCh <- printer.NewErrorMessage(fmt.Sprintf("Error reading log file: %s", err.Error()))
					}
					break
				}
				logCh <- strings.TrimSpace(string(msg))
				logCnt++
			}
			metCh <- alert.NewCounterMetric(logCnt, t)
		case <-ctx.Done():
			break loop
		}
	}
}
