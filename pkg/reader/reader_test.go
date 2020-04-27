package reader

import (
	"context"
	"io/ioutil"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	alert "httplogmonitor/pkg/alertmanager"
	"httplogmonitor/pkg/config"
	"httplogmonitor/pkg/printer"
)

func TestReaderNominal(t *testing.T) {
	cfg := config.NewDefault()
	r := New(cfg)

	t.Log("Creating the test file")
	f, err := ioutil.TempFile("", "test")
	if err != nil {
		t.Skip("Failed to create the test file: ", err)
	}
	defer os.Remove(f.Name())
	cfg.LogFilePath = f.Name()

	logCh := make(chan string, 3)
	metCh := make(chan alert.Metric)
	printCh := make(chan printer.Formatter)
	ctx, cancelCtx := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go r.Start(ctx, logCh, metCh, printCh, wg)

	// give a chance to reader to set up the offset to the end
	runtime.Gosched()

	t.Log("Adding log entries")
	inputs := []string{
		" here it comes\n",
		"  here it goes\n",
		"\there we are\n",
	}
	for _, s := range inputs {
		f.WriteString(s)
	}
	f.Sync()

	outputs := []string{}
	tick := time.NewTicker(time.Duration(cfg.PollIntervalSec) * time.Second)
	defer tick.Stop()

	t.Log("Waiting for reader to ingest the entries")
loop:
	for {
		select {
		case l := <-logCh:
			outputs = append(outputs, l)
			if len(outputs) == len(inputs) {
				break loop
			}
		case <-tick.C:
			t.Fatal("Timed out waiting for outputs")
		}
	}

	t.Log("Matching the read entries")
	expected := []string{
		"here it comes",
		"here it goes",
		"here we are",
	}
	for i := range expected {
		if expected[i] != outputs[i] {
			t.Fatalf("Output %q doesn't match expected %q\n", outputs[i], expected[i])
		}
	}

	t.Log("Checking the metric")
	gotMetric := <-metCh
	gotCntMetric, ok := gotMetric.(alert.CounterMetric)
	if !ok {
		t.Fatal("Counter metric expected")
	}
	// we know we have int value here
	gotCnt, _ := gotCntMetric.Value().(int)

	expectedCnt := len(inputs)
	if expectedCnt != gotCnt {
		t.Fatalf("Expected counter %d, got %d", expectedCnt, gotCnt)
	}

	t.Log("Waiting for reader to stop")
	cancelCtx()
	wg.Wait()

	// no way to check the graceful stop of the file
	// checking for the closure of the file may be a race
	// so, no testing of the graceful close of the file
}
