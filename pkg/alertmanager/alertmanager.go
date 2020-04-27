package alertmanager

import (
	"fmt"
	"math"
	"time"

	"httplogmonitor/pkg/config"
	"httplogmonitor/pkg/printer"
)

// AlertManager collects the traffic metrics and prints them every summary interval
type AlertManager struct {
	buf            []int
	threshold      int
	ptr            int
	sum            int
	alertTriggered bool
}

// New returns a new instance of AlertManager
func New(cfg *config.Config) *AlertManager {
	return &AlertManager{
		buf:       make([]int, 0, cfg.MonitorWindowSec/cfg.PollIntervalSec),
		threshold: cfg.AlertThreshold,
		ptr:       0,
		sum:       0,
	}
}

// Start listens on the metric channel and sends the alert/clear alert/alerton and regular avg traffic messages to the printer
func (a *AlertManager) Start(metCh <-chan Metric, printCh chan<- printer.Formatter) {
	alertOnPrinted := false
	for m := range metCh {
		a.add(m)
		// regular avg traffic message, displayed only in verbose mode
		printCh <- printer.NewMessage(fmt.Sprintf("\tAverage traffic: %d/s", a.AvgTraffic()))

		if a.AlertOn() && !alertOnPrinted {
			// we have all the needed data, let's inform the user about that
			printCh <- printer.NewInfoMessage("All needed metrics are collected. Alerting is on")
			alertOnPrinted = true
		}
		switch a.Alert() {
		case 1:
			// fire the alert
			printCh <- printer.NewAlertMessage(a.AvgTraffic(), m.Time())
		case -1:
			// clear the alert message
			printCh <- printer.NewClearAlertMessage(a.AvgTraffic(), m.Time())
		}
	}
}

// AvgTraffic average traffic (hits per second) for the monitoring window
func (a *AlertManager) AvgTraffic() int {
	if a.sum == 0 && len(a.buf) == 0 {
		return 0
	}
	// (x).5 will be rounded up to (x+1).0
	return int(math.Round(float64(a.sum) / float64(len(a.buf))))
}

// Alert returns 1 if the average traffic for the past monitoring window is higher than the threshold
// returns -1 if the average traffic has decreased below the threshold
// return 0 if no alerts/clear alerts need to be sent
// Note: alerting is ON only if we have enough data (monitoring window)
func (a *AlertManager) Alert() int {
	if !a.AlertOn() {
		// no alert until we get enough data
		return 0
	}

	if !a.alertTriggered {
		if a.AvgTraffic() >= a.threshold {
			a.alertTriggered = true
			// alert is to be triggered
			return 1
		}
	} else {
		if a.AvgTraffic() < a.threshold {
			a.alertTriggered = false
			// alert is to be cleared
			return -1
		}
	}
	// alert is not triggered but the traffic is still ok
	// or
	// alert is triggered but the traffic is still not ok
	// ==
	// nothing to do
	return 0
}

// AlertOn returns true if the alerting is ready (enough data is collected)
func (a *AlertManager) AlertOn() bool {
	return a.bufFull()
}

// add adds the given metric to the stats collected by the alertmanager
func (a *AlertManager) add(m Metric) {
	// we know that we deal with the counter here
	// as it's the only metric used in the whole program
	cnt, _ := m.Value().(int)

	if !a.fill(cnt) {
		return
	}
	a.addCircular(cnt)
}

// bufFull returns true if the internal buffer is full
func (a *AlertManager) bufFull() bool {
	return len(a.buf) == cap(a.buf)
}

func (a *AlertManager) getBuf() []int {
	return a.buf
}

// fill adds to the sum of all the hits
// returns true if the window is filled
func (a *AlertManager) fill(cnt int) bool {
	if !a.bufFull() {
		a.buf = append(a.buf, cnt)
		a.sum += cnt
		return false
	}
	return true
}

// addCircular adds to the sum in the circular manner
// that is, it overrides the oldest items with the new ones
func (a *AlertManager) addCircular(cnt int) {
	prevCnt := a.buf[a.ptr]
	a.sum -= prevCnt

	a.buf[a.ptr] = cnt
	a.sum += cnt
	if a.ptr >= len(a.buf)-1 {
		a.ptr = 0
	} else {
		a.ptr++
	}
}

// Metric represents the generic metric type
type Metric interface {
	Time() time.Time
	Value() interface{}
}

// CounterMetric represent a single counter metric
type CounterMetric struct {
	count int
	time  time.Time
}

// NewCounterMetric returns a new instance of CounterMetric
func NewCounterMetric(cnt int, time time.Time) CounterMetric {
	return CounterMetric{
		count: cnt,
		time:  time,
	}
}

// Time returns exact time at which the metric was started to be collected
func (c CounterMetric) Time() time.Time {
	return c.time
}

// Value returns the counter
func (c CounterMetric) Value() interface{} {
	return c.count
}
