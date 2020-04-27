package collector

import (
	"errors"
	"math"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"httplogmonitor/pkg/printer"
)

// Example: 127.0.0.1 - james [09/May/2018:16:00:39 +0000] "GET /report HTTP/1.0" 200 123
var (
	w3cLogEntryRegExp = regexp.MustCompile(`^.+? .+? .+? \[.+?\] "(.+?)" (\d+) \d+$`)
	methodPathRegExp  = regexp.MustCompile(`^(\w+) (/.*?) `)
)

// LogMessage represents the parsed log entry
// with only the data we are interested in
type LogMessage struct {
	Section string
	Method  string
	Code    int
}

// NewLogMessageFromLogEntry parses the raw log entry validating it therefore
// and extracts only the interesting fields
func NewLogMessageFromLogEntry(str string) (*LogMessage, error) {
	msg := &LogMessage{}
	m1 := w3cLogEntryRegExp.FindStringSubmatch(str)
	if m1 != nil && len(m1) == 3 {
		// got request and code
		code, err := strconv.Atoi(m1[2])
		if err != nil {
			return nil, err
		}
		if st := http.StatusText(code); len(st) == 0 {
			return nil, errors.New("unknown http code")
		}
		msg.Code = code

		m2 := methodPathRegExp.FindStringSubmatch(m1[1])
		if m2 != nil && len(m2) == 3 {
			// got method and section
			msg.Method = m2[1]

			// skipping first slash as cheched by regexp
			if i := strings.Index(m2[2][1:], "/"); i == -1 {
				// only section in path
				msg.Section = m2[2]
			} else {
				// cut all but section
				msg.Section = m2[2][0 : i+1]
			}
		} else {
			return nil, errors.New("method and path format not matched")
		}
	} else {
		return nil, errors.New("w3c log entry format not matched")
	}
	return msg, nil
}

// Equal compares the log message field by field to the given one
func (m *LogMessage) Equal(other *LogMessage) bool {
	if m.Section != other.Section {
		return false
	}
	if m.Method != other.Method {
		return false
	}
	if m.Code != other.Code {
		return false
	}
	return true
}

const (
	hitsKey     = "hits"
	successKey  = "success"
	redirectKey = "redirect"
	errorsKey   = "errors"
	trafficKey  = "traffic"
)

// Summary represents the whole summary to be displayed every summary interval
// is made of 2 parts: top hitted sections and summary of interesting stats for the past summary interval
type Summary struct {
	Sections map[string]int
	Sum      map[string]int
	topNum   int
}

// NewSummary returns a new instance of Summary with given limit for most hitted sections
func NewSummary(top int) *Summary {
	return &Summary{
		Sections: map[string]int{},
		Sum:      map[string]int{},
		topNum:   top,
	}
}

// Add adds the stats of the given log message to the summary
func (s *Summary) Add(m *LogMessage) {
	s.Sum[hitsKey]++
	s.Sections[m.Section]++

	switch m.Code / 100 {
	case 5:
		fallthrough
	case 4:
		s.Sum[errorsKey]++
	case 3:
		s.Sum[redirectKey]++
	case 2:
		s.Sum[successKey]++
	}
}

// CalcTraffic calculates the traffic for the one summary
func (s *Summary) CalcTraffic(win int) {
	s.Sum[trafficKey] = int(math.Round(float64(s.Sum[hitsKey]) / float64(win)))
}

// Format formats the summary structure as two 2d tables ready to be printed
func (s Summary) Format() string {
	b := strings.Builder{}

	// top sections table
	tblT := printer.NewTable2dMessage("TOP SECTIONS", "Section", "Number of hits")
	if len(s.Sections) == 0 {
		tblT.AddRow("<no section data>", "")
	} else {
		// sorting, filtering and formatting the section data
		om := make([]struct {
			key   string
			value int
		}, 0, len(s.Sections))
		for k, v := range s.Sections {
			p := struct {
				key   string
				value int
			}{k, v}
			om = append(om, p)
		}
		sort.Slice(om, func(i, j int) bool { return om[i].value > om[j].value })
		for i := 0; i < len(om) || i == s.topNum; i++ {
			tblT.AddRow(om[i].key, strconv.Itoa(om[i].value))
		}
	}

	// summary table
	tblS := printer.NewTable2dMessage("SUMMARY", "Detail", "Value")
	tblS.AddRow("Total hits", strconv.Itoa(s.Sum[hitsKey]))
	tblS.AddRow("Traffic (per second)", strconv.Itoa(s.Sum[trafficKey]))
	tblS.AddRow("Total success", strconv.Itoa(s.Sum[successKey]))
	tblS.AddRow("Total redirects", strconv.Itoa(s.Sum[redirectKey]))
	tblS.AddRow("Total errors", strconv.Itoa(s.Sum[errorsKey]))

	// align 2 tables
	if tblS.Length() > tblT.Length() {
		tblT.Enlarge(tblS.Length())
	} else if tblT.Length() > tblS.Length() {
		tblS.Enlarge(tblT.Length())
	}

	// add both tables to the output after the alignment
	b.WriteString(tblT.Format())
	b.WriteString(tblS.Format())

	return b.String()
}

// Verbose returns false as summary is to be always displayed
func (s Summary) Verbose() bool {
	return false
}
