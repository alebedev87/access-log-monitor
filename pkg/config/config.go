package config

import (
	"errors"
	"flag"
	"strings"
)

const (
	defaultLogFilePath        = "/tmp/access.log"
	defaultSummaryIntervalSec = 10
	defaultPollIntervalSec    = 1
	defaultMonitorWindowSec   = 120
	defaultAlertThreshold     = 10
	defaultTopSectionNum      = 10
	defaultLogBufferSize      = 10
	defaultMetricBufferSize   = 5
	defaultVerbose            = false
)

// Config stores the configuration to the whole program
type Config struct {
	LogFilePath        string
	SummaryIntervalSec int
	PollIntervalSec    int
	MonitorWindowSec   int
	AlertThreshold     int
	TopSectionNum      int
	LogBufferSize      int
	MetricBufferSize   int
	Verbose            bool
}

// NewDefault returns the configuration with only default values
func NewDefault() *Config {
	return &Config{
		LogFilePath:        defaultLogFilePath,
		SummaryIntervalSec: defaultSummaryIntervalSec,
		PollIntervalSec:    defaultPollIntervalSec,
		MonitorWindowSec:   defaultMonitorWindowSec,
		AlertThreshold:     defaultAlertThreshold,
		TopSectionNum:      defaultTopSectionNum,
		LogBufferSize:      defaultLogBufferSize,
		MetricBufferSize:   defaultMetricBufferSize,
		Verbose:            defaultVerbose,
	}
}

// NewFromArgs returns the configuration filled from flags passed to the program
func NewFromArgs() *Config {
	cfg := &Config{
		LogBufferSize:    defaultLogBufferSize,
		MetricBufferSize: defaultMetricBufferSize,
	}

	flag.StringVar(&cfg.LogFilePath, "f", defaultLogFilePath, "Path to the log file.")
	flag.IntVar(&cfg.SummaryIntervalSec, "i", defaultSummaryIntervalSec, "Interval between summary displays (seconds).")
	flag.IntVar(&cfg.PollIntervalSec, "p", defaultPollIntervalSec, "Polling interval (seconds).")
	flag.IntVar(&cfg.MonitorWindowSec, "w", defaultMonitorWindowSec, "Monitoring window (seconds).")
	flag.IntVar(&cfg.AlertThreshold, "t", defaultAlertThreshold, "Alerting threshold (hits per second).")
	flag.IntVar(&cfg.TopSectionNum, "n", defaultTopSectionNum, "How many most hitted sections need to be displayed.")
	flag.BoolVar(&cfg.Verbose, "v", defaultVerbose, "Be verbose (show regular average traffic stats).")
	flag.Parse()

	err := cfg.Validate()
	if err != nil {
		panic(err)
	}

	return cfg
}

// Validate validates the important fields of the configuration
func (c *Config) Validate() error {
	if len(strings.TrimSpace(c.LogFilePath)) == 0 {
		return errors.New("No log file provided")
	}

	if c.PollIntervalSec <= 0 {
		return errors.New("polling interval cannot be less than 1 second")
	}

	if c.SummaryIntervalSec <= 0 {
		return errors.New("interval between summary displays cannot be less than 1 second")
	}

	if c.PollIntervalSec >= c.SummaryIntervalSec {
		return errors.New("summary interval must be greater than polling interval")
	}

	if c.MonitorWindowSec <= 0 {
		return errors.New("monitoring window cannot less than 1 second")
	}

	// adding this pre-requisite just to simplify the implementation
	if c.MonitorWindowSec%c.PollIntervalSec != 0 {
		return errors.New("polling interval must be a divisor of monitoring window value. Try 1s for the polling interval, it's a good divisor ;)")
	}

	if c.AlertThreshold <= 0 {
		return errors.New("alert threshold cannot be less than 1 hit per second")
	}

	if c.TopSectionNum <= 0 {
		return errors.New("number of most hitted sections cannot be less than 1")
	}

	return nil
}
