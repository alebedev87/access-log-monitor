package collector

import (
	"reflect"
	"testing"

	"httplogmonitor/pkg/config"
	"httplogmonitor/pkg/printer"
)

func TestCollectorNominal(t *testing.T) {
	cfg := config.NewDefault()
	cfg.TopSectionNum = 2
	cfg.SummaryIntervalSec = 2
	c := New(cfg)

	logCh := make(chan string, 5)
	printCh := make(chan printer.Formatter)

	logCh <- `127.0.0.1 - james [09/May/2018:16:00:39 +0000] "GET /report HTTP/1.0" 200 123`
	logCh <- `127.0.0.1 - james [09/May/2018:16:01:39 +0000] "POST /report HTTP/1.0" 202 123`
	logCh <- `127.0.0.1 - james [09/May/2018:16:02:39 +0000] "PUT /unknown HTTP/1.0" 404 123`
	logCh <- `127.0.0.1 - james [09/May/2018:16:03:39 +0000] "GET /report HTTP/1.0" 200 123`
	logCh <- `127.0.0.1 - james [09/May/2018:16:04:39 +0000] "PUT /unknown HTTP/1.0" 500 123`

	go c.Start(logCh, printCh)

	gotSummary := <-printCh

	expectedSummary := Summary{
		Sections: map[string]int{
			"/report":  3,
			"/unknown": 2,
		},
		Sum: map[string]int{
			hitsKey:    5,
			successKey: 3,
			errorsKey:  2,
			trafficKey: 3,
		},
		topNum: 2,
	}

	if !reflect.DeepEqual(expectedSummary, gotSummary) {
		t.Fatalf("Excepted summary %#v, got summary %#v", expectedSummary, gotSummary)
	}
}
