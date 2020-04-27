package config

import (
	"testing"
)

func TestValidateConfig(t *testing.T) {
	testCases := []struct {
		name          string
		input         *Config
		expectedError bool
	}{
		{
			name:          "Defaults",
			input:         NewDefault(),
			expectedError: false,
		},
		{
			name:          "No log file",
			input:         newDefaultLog("\t   \n"),
			expectedError: true,
		},
		{
			name:          "Poll interval too small",
			input:         newDefaultPoll(0),
			expectedError: true,
		},
		{
			name:          "Summary interval too small",
			input:         newDefaultSum(0),
			expectedError: true,
		},
		{
			name:          "Poll greater than summary",
			input:         newDefaultPollSum(5, 5),
			expectedError: true,
		},
		{
			name:          "Monitoring window too small",
			input:         newDefaultWin(0),
			expectedError: true,
		},
		{
			name:          "Poll interval to monitoring window",
			input:         newDefaultWinPoll(10, 3),
			expectedError: true,
		},
		{
			name:          "Threshold is too small",
			input:         newDefaultAlert(0),
			expectedError: true,
		},
		{
			name:          "Top section is too small",
			input:         newDefaultTop(0),
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.input.Validate()
			if err != nil {
				if !tc.expectedError {
					t.Errorf("Test case %q got not expected error", tc.name)
				}
			} else {
				if tc.expectedError {
					t.Errorf("Test case %q got no error while one is expected", tc.name)
				}
			}
		})
	}
}

func newDefaultLog(log string) *Config {
	cfg := NewDefault()
	cfg.LogFilePath = log
	return cfg
}

func newDefaultSum(sum int) *Config {
	cfg := NewDefault()
	cfg.SummaryIntervalSec = sum
	return cfg
}

func newDefaultPoll(poll int) *Config {
	cfg := NewDefault()
	cfg.PollIntervalSec = poll
	return cfg
}

func newDefaultPollSum(poll, sum int) *Config {
	cfg := NewDefault()
	cfg.PollIntervalSec = poll
	cfg.SummaryIntervalSec = sum
	return cfg
}

func newDefaultWin(win int) *Config {
	cfg := NewDefault()
	cfg.MonitorWindowSec = win
	return cfg
}

func newDefaultWinPoll(win, poll int) *Config {
	cfg := NewDefault()
	cfg.PollIntervalSec = poll
	cfg.MonitorWindowSec = win
	return cfg
}

func newDefaultAlert(t int) *Config {
	cfg := NewDefault()
	cfg.AlertThreshold = t
	return cfg
}

func newDefaultTop(t int) *Config {
	cfg := NewDefault()
	cfg.TopSectionNum = t
	return cfg
}
