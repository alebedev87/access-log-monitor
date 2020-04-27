package collector

import (
	"reflect"
	"testing"
)

func TestLogMessageEqual(t *testing.T) {
	expectedMsg := LogMessage{
		Section: "/api",
		Method:  "GET",
		Code:    200,
	}

	testCases := []struct {
		name     string
		input    LogMessage
		expected bool
	}{
		{
			name: "Equal",
			input: LogMessage{
				Section: "/api",
				Method:  "GET",
				Code:    200,
			},
			expected: true,
		},
		{
			name: "Not equal section",
			input: LogMessage{
				Section: "/api2",
				Method:  "GET",
				Code:    200,
			},
			expected: false,
		},
		{
			name: "Not equal method",
			input: LogMessage{
				Section: "/api",
				Method:  "POST",
				Code:    200,
			},
			expected: false,
		},
		{
			name: "Not equal code",
			input: LogMessage{
				Section: "/api",
				Method:  "GET",
				Code:    202,
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := expectedMsg.Equal(&tc.input)
			if output != tc.expected {
				t.Errorf("Test case %q: output didn't match", tc.name)
			}
		})
	}
}

func TestNewLogMessageFromLogEntry(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expected    LogMessage
		expectedErr bool
	}{
		{
			name:  "Nominal 1",
			input: `127.0.0.1 - james [09/May/2018:16:00:39 +0000] "GET /report HTTP/1.0" 200 123`,
			expected: LogMessage{
				Section: "/report",
				Method:  "GET",
				Code:    200,
			},
		},
		{
			name:  "Nominal 2",
			input: `127.0.0.1 - jill [09/May/2018:16:00:41 +0000] "GET /api/user HTTP/1.0" 200 234`,
			expected: LogMessage{
				Section: "/api",
				Method:  "GET",
				Code:    200,
			},
		},
		{
			name:  "Nominal 3",
			input: `127.0.0.1 - frank [09/May/2018:16:00:42 +0000] "POST / HTTP/1.0" 200 34`,
			expected: LogMessage{
				Section: "/",
				Method:  "POST",
				Code:    200,
			},
		},
		{
			name:  "Nominal 4",
			input: `127.0.0.1 - mary [09/May/2018:16:00:42 +0000] "GET /api/user HTTP/1.0" 503 12`,
			expected: LogMessage{
				Section: "/api",
				Method:  "GET",
				Code:    503,
			},
		},
		{
			name:  "Nominal dns name",
			input: `hostname.xyz.com - mary [09/May/2018:16:00:42 +0000] "GET /api/user HTTP/1.0" 200 12`,
			expected: LogMessage{
				Section: "/api",
				Method:  "GET",
				Code:    200,
			},
		},
		{
			name:  "Nominal absent rfc931 and authuser",
			input: `hostname.xyz.com - - [09/May/2018:16:00:42 +0000] "GET /api/user HTTP/1.0" 200 12`,
			expected: LogMessage{
				Section: "/api",
				Method:  "GET",
				Code:    200,
			},
		},
		{
			name:        "Error empty request",
			input:       `127.0.0.1 - - [09/May/2018:16:00:42 +0000] "" 200 12`,
			expectedErr: true,
		},
		{
			name:        "Error no request",
			input:       `127.0.0.1 - mary [09/May/2018:16:00:42 +0000] 200 12`,
			expectedErr: true,
		},
		{
			name:        "Error no code",
			input:       `127.0.0.1 - mary [09/May/2018:16:00:42 +0000] "GET /api/user HTTP/1.0" 12`,
			expectedErr: true,
		},
		{
			name:        "Error wrong code",
			input:       `127.0.0.1 - mary [09/May/2018:16:00:42 +0000] "GET /api/user HTTP/1.0" 999 12`,
			expectedErr: true,
		},
		{
			name:        "Error no remotehost",
			input:       `- mary [09/May/2018:16:00:42 +0000] "GET /api/user HTTP/1.0" 200 12`,
			expectedErr: true,
		},
		{
			name:        "Error no rfc931",
			input:       `127.0.0.1 mary [09/May/2018:16:00:42 +0000] "GET /api/user HTTP/1.0" 200 12`,
			expectedErr: true,
		},
		{
			name:        "Error no authuser",
			input:       `127.0.0.1 - [09/May/2018:16:00:42 +0000] "GET /api/user HTTP/1.0" 200 12`,
			expectedErr: true,
		},
		{
			name:        "Error empty date",
			input:       `127.0.0.1 - mary [] "GET /api/user HTTP/1.0" 200 12`,
			expectedErr: true,
		},
		{
			name:        "Error no date",
			input:       `127.0.0.1 - - "GET /api/user HTTP/1.0" 200 12`,
			expectedErr: true,
		},
		{
			name:        "Error no bytes",
			input:       `127.0.0.1 - mary [09/May/2018:16:00:42 +0000] "GET /api/user HTTP/1.0" 200`,
			expectedErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := NewLogMessageFromLogEntry(tc.input)
			if err != nil {
				if !tc.expectedErr {
					t.Errorf("Test case %q got not expected error", tc.name)
				}
				return
			}
			if tc.expectedErr {
				t.Errorf("Test case %q got no error while one is expected", tc.name)
				return
			}

			if !output.Equal(&tc.expected) {
				t.Errorf("Test case %q: output didn't match", tc.name)
			}
		})
	}
}

func TestSummary(t *testing.T) {
	limit := 5
	window := 5
	logMsg := []*LogMessage{
		&LogMessage{
			Section: "/here",
			Method:  "GET",
			Code:    200,
		},
		&LogMessage{
			Section: "/here",
			Method:  "PUT",
			Code:    200,
		},
		&LogMessage{
			Section: "/here",
			Method:  "GET",
			Code:    202,
		},
		&LogMessage{
			Section: "/here",
			Method:  "POST",
			Code:    202,
		},
		&LogMessage{
			Section: "/",
			Method:  "POST",
			Code:    404,
		},
		&LogMessage{
			Section: "/",
			Method:  "POST",
			Code:    500,
		},
		&LogMessage{
			Section: "/",
			Method:  "GET",
			Code:    200,
		},
		&LogMessage{
			Section: "/there",
			Method:  "GET",
			Code:    200,
		},
		&LogMessage{
			Section: "/there",
			Method:  "PUT",
			Code:    202,
		},
		&LogMessage{
			Section: "/redirect",
			Method:  "POST",
			Code:    302,
		},
	}
	expectedSections := map[string]int{
		"/here":     4,
		"/":         3,
		"/there":    2,
		"/redirect": 1,
	}

	expectedSum := map[string]int{
		hitsKey:     10,
		successKey:  7,
		errorsKey:   2,
		redirectKey: 1,
		trafficKey:  2,
	}

	sum := NewSummary(limit)
	for i := range logMsg {
		sum.Add(logMsg[i])
	}
	sum.CalcTraffic(window)

	t.Log("Checking summary internals")
	if !reflect.DeepEqual(sum.Sections, expectedSections) {
		t.Fatalf("Expected sections %+v, got sections %+v", expectedSections, sum.Sections)
	}
	if !reflect.DeepEqual(sum.Sum, expectedSum) {
		t.Fatalf("Expected summary %+v, got summary %+v", expectedSum, sum.Sum)
	}

	t.Log("Checking summary formatting")
	expectedFormat := `
--------TOP SECTIONS---------
  Section      Number of hits
-----------    --------------
/here                       4
/                           3
/there                      2
/redirect                   1

-----------SUMMARY-----------
       Detail           Value
--------------------    -----
Total hits                 10
Traffic (per second)        2
Total success               7
Total redirects             1
Total errors                2
`
	gotFormat := sum.Format()

	if expectedFormat != gotFormat {
		t.Fatalf("Expected format %s, got format %s", expectedFormat, gotFormat)
	}
}
