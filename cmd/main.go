package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	alert "httplogmonitor/pkg/alertmanager"
	"httplogmonitor/pkg/collector"
	"httplogmonitor/pkg/config"
	"httplogmonitor/pkg/printer"
	"httplogmonitor/pkg/reader"
)

func main() {
	// read the program args
	cfg := config.NewFromArgs()

	// workers
	r := reader.New(cfg)
	c := collector.New(cfg)
	a := alert.New(cfg)
	p := printer.New(cfg)

	// using context and waitgroup to do a graceful stop of Reader.
	// strictly speaking, it's not really necessary as we only read
	// and all the resources (file descriptor) will be freed once the process is terminated.
	// however, as the log file is the only "precious" artifact here - let's close it correctly
	wg := &sync.WaitGroup{}
	wg.Add(1)
	ctx, cancelCtx := context.WithCancel(context.Background())
	// making log and metric channels buffered
	// to not break the reader's tick in case of the slow consumers
	// however, it's unlikely that the consumers will be slower than Reader
	logCh := make(chan string, cfg.LogBufferSize)
	metCh := make(chan alert.Metric, cfg.MetricBufferSize)
	printCh := make(chan printer.Formatter)

	go r.Start(ctx, logCh, metCh, printCh, wg)
	go c.Start(logCh, printCh)
	go a.Start(metCh, printCh)
	go p.Start(printCh)

	// signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-sigCh
	cancelCtx()

	wg.Wait()
	fmt.Println("\nStopped")
}
