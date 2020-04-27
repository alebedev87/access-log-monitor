package alertmanager

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"httplogmonitor/pkg/config"
	"httplogmonitor/pkg/printer"
)

const (
	timeFormat        = "2006-01-02 15:04:05.000"
	timeout           = time.Duration(1000) * time.Millisecond
	alertPattern      = "High traffic generated an alert - hits"
	clearAlertPattern = "High traffic alert cleared"
)

func TestAlertManagerNominal(t *testing.T) {
	expectedBufSize := 5

	cfg := config.NewDefault()
	cfg.PollIntervalSec = 2
	cfg.MonitorWindowSec = 10
	a := New(cfg)

	metCh := make(chan Metric, expectedBufSize)
	printCh := make(chan printer.Formatter)

	t1, _ := time.Parse(timeFormat, "2019-11-30 15:00:01.100")
	t2, _ := time.Parse(timeFormat, "2019-11-30 15:00:02.100")
	t3, _ := time.Parse(timeFormat, "2019-11-30 15:00:03.100")
	t4, _ := time.Parse(timeFormat, "2019-11-30 15:00:04.100")
	metCh <- NewCounterMetric(20, t1)
	metCh <- NewCounterMetric(20, t2)
	metCh <- NewCounterMetric(20, t3)
	metCh <- NewCounterMetric(20, t4)

	go a.Start(metCh, printCh)

	// blocking on avg traffic message which is sent after the metric is added
	<-printCh
	<-printCh
	<-printCh
	<-printCh

	t.Log("Checking monitor window")
	// monitoring window is not full yet
	if a.AlertOn() {
		t.Fatal("Alert must not be on yet")
	}

loop1:
	// checking print output: we must have no messages yet
	for {
		select {
		case m := <-printCh:
			if !m.Verbose() {
				t.Fatalf("Got message while monitoring window is not complete yet: %s", m.Format())
			}
		case <-time.After(timeout):
			break loop1
		}
	}

	t.Log("Triggering alert")
	// sending last metric before the monitor window becomes full
	t5, _ := time.Parse(timeFormat, "2019-11-30 15:00:05.100")
	metCh <- NewCounterMetric(23, t5)

	// get avg traffic first
	<-printCh
	// and interresting messages after
	gotAlertOn := <-printCh
	gotAlert := <-printCh

	// 103 total hits for monitor window of size 5
	// this gives 20.6 avg traffic which should rounded up to 21
	expectedAvgTraffic := 21
	alertOnRegExp := regexp.MustCompile(`\[INFO\].*Alerting is on.*`)
	alertRegExp := regexp.MustCompile(fmt.Sprintf(`\[ALERT\].*%s = %d.*triggered at %s`, alertPattern, expectedAvgTraffic, t5.Format(timeFormat)))

	t.Log("Checking alert message")
	if !alertOnRegExp.MatchString(gotAlertOn.Format()) {
		t.Fatalf("Got wrong alerton message: %s", gotAlertOn.Format())
	}

	if !alertRegExp.MatchString(gotAlert.Format()) {
		t.Fatalf("Got wrong alert message: %s", gotAlert.Format())
	}

	t6, _ := time.Parse(timeFormat, "2019-11-30 15:00:06.100")
	metCh <- NewCounterMetric(20, t6)

	// skip avg traffic
	<-printCh

	t.Log("Checking for no duplicate alert message")
loop2:
	// checking print output: we must have no messages as alert and alerton don't repeat
	for {
		select {
		case m := <-printCh:
			if !m.Verbose() {
				t.Fatalf("Got message while should not be repeated: %s", m.Format())
			}
		case <-time.After(timeout):
			break loop2
		}
	}

	t.Log("Lowering the traffic")
	t7, _ := time.Parse(timeFormat, "2019-11-30 15:00:07.100")
	// 17 h/s
	metCh <- NewCounterMetric(0, t7)
	t8, _ := time.Parse(timeFormat, "2019-11-30 15:00:08.100")
	// 13 h/s
	metCh <- NewCounterMetric(0, t8)
	t9, _ := time.Parse(timeFormat, "2019-11-30 15:00:09.100")
	// 9 h/s
	metCh <- NewCounterMetric(0, t9)

	// skipping all avg traffic ones
	<-printCh
	<-printCh
	<-printCh
	gotClearAlert := <-printCh

	expectedAvgTraffic = 9
	clearAlertRegExp := regexp.MustCompile(fmt.Sprintf(`\[CLEAR\].*%s at %s.*hits = %d`, clearAlertPattern, t9.Format(timeFormat), expectedAvgTraffic))

	t.Log("Checking clear alert message")
	if !clearAlertRegExp.MatchString(gotClearAlert.Format()) {
		t.Fatalf("Got wrong clear alert message: %s", gotClearAlert.Format())
	}
}
