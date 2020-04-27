package printer

import (
	"fmt"
	"strings"
	"time"
)

const (
	timeFormat = "2006-01-02 15:04:05.000"
)

// Formatter represents a formatted message
type Formatter interface {
	Format() string
	// Verbose returns true if the message is only for the verbose mode
	Verbose() bool
}

// Message represents a simple verbose message
type Message struct {
	Text string
}

// NewMessage gives a new instance of the simple verbose message with given text
func NewMessage(text string) Message {
	return Message{Text: text}
}

// Format doesn't do any formatting, just prints the given text
func (m Message) Format() string {
	return m.Text
}

// Verbose returns true as the simple message is for the verbose mode
func (m Message) Verbose() bool {
	return true
}

// ErrorMessage represents an error message
type ErrorMessage struct {
	Message
}

// NewErrorMessage gives a new instance of the error message with given text
func NewErrorMessage(text string) ErrorMessage {
	return ErrorMessage{NewMessage(text)}
}

// Format wraps the given text into ERROR label
func (m ErrorMessage) Format() string {
	return wrapError(m.Text)
}

// Verbose returns false as the error message is to be always displayed
func (m ErrorMessage) Verbose() bool {
	return false
}

// AlertMessage represents the high traffic alert messsage
type AlertMessage struct {
	Hits int
	Time time.Time
}

// NewAlertMessage gives a new instance of the alert message
// with given number of hits and the time at which it was triggered
func NewAlertMessage(h int, t time.Time) AlertMessage {
	return AlertMessage{
		Hits: h,
		Time: t,
	}
}

// Format returns the predefined alert text for the high traffic
// wrapped into ALERT label
func (m AlertMessage) Format() string {
	return wrapAlert(fmt.Sprintf("High traffic generated an alert - hits = %d, triggered at %s", m.Hits, m.Time.Format(timeFormat)))
}

// Verbose returns false as the alert message is to be always displayed
func (m AlertMessage) Verbose() bool {
	return false
}

// ClearAlertMessage represents the clearance message for a previously generated high traffic alert
type ClearAlertMessage struct {
	AlertMessage
}

// NewClearAlertMessage gives a new instance of the clearance message,
// just like the alert message it expects the same inputs
func NewClearAlertMessage(h int, t time.Time) ClearAlertMessage {
	return ClearAlertMessage{NewAlertMessage(h, t)}
}

// Format returns the predefined clearance text for the previously generated alert
// wrapped into CLEAR label
func (m ClearAlertMessage) Format() string {
	return wrapClearAlert(fmt.Sprintf("High traffic alert cleared at %s. Current hits = %d", m.Time.Format(timeFormat), m.Hits))
}

// Verbose returns false as the clearance message is to be always displayed
func (m ClearAlertMessage) Verbose() bool {
	return false
}

// InfoMessage represents an information message
type InfoMessage struct {
	Message
}

// NewInfoMessage gives a new instance of the information message with the given text
func NewInfoMessage(text string) InfoMessage {
	return InfoMessage{NewMessage(text)}
}

// Format wraps the given text into INFO label
func (m InfoMessage) Format() string {
	return wrapInfo(m.Text)
}

// Verbose returns false as the information message is to be always displayed
func (m InfoMessage) Verbose() bool {
	return false
}

// string key/value pair
type strPair struct {
	Key   string
	Value string
}

// Table2dMessage helper struct to construct primitive 2D tables for printing
type Table2dMessage struct {
	Name   string
	Titles [2]string
	Rows   []strPair // no map used as we need the stable order
	max    [2]int
	colSep string
}

// NewTable2dMessage gives a new instance of 2d table with given name and 2 titles
func NewTable2dMessage(n string, t1, t2 string) *Table2dMessage {
	tbl := &Table2dMessage{Name: n, colSep: "    "}
	tbl.Titles = [2]string{t1, t2}
	tbl.max = [2]int{len(t1), len(t2)}
	return tbl
}

// AddRow adds a non title row into 2d table
func (t *Table2dMessage) AddRow(k, v string) {
	t.Rows = append(t.Rows, strPair{Key: k, Value: v})
	// calculate maxes
	if len(k) > t.max[0] {
		t.max[0] = len(k)
	}
	if len(v) > t.max[1] {
		t.max[1] = len(v)
	}
}

// Length gives the width of 2d table
// where the width is the sum of the max lengths of the columns
func (t *Table2dMessage) Length() int {
	return t.max[0] + t.max[1] + len(t.colSep)
}

// Enlarge resize the table to the given new width
// padding will be added to all the rows, titles and the name
func (t *Table2dMessage) Enlarge(newLen int) {
	// do not shrink
	if newLen <= t.Length() {
		return
	}

	// size of both columns
	allSize := newLen - len(t.colSep)
	// fair size of one column
	colSize := allSize / 2
	if t.max[0] > colSize {
		// keep the max of left and adjust the right
		t.max[1] = allSize - t.max[0]
	} else if t.max[1] > colSize {
		// keep the max of right and adjust the left
		t.max[0] = allSize - t.max[1]
	} else {
		t.max[0] = colSize
		t.max[1] = colSize
	}
}

// Format returns 2d table ready to be displayed
func (t *Table2dMessage) Format() string {
	b := strings.Builder{}

	// name
	b.WriteByte('\n')
	b.WriteString(center(t.Name, "-", t.Length()))
	b.WriteByte('\n')

	// titles
	for i, tl := range t.Titles {
		b.WriteString(center(tl, " ", t.max[i]))
		if i != len(t.Titles)-1 {
			b.WriteString(t.colSep)
		}
	}
	b.WriteByte('\n')

	// title outlines
	for i, m := range t.max {
		b.WriteString(strings.Repeat("-", m))
		if i != len(t.max)-1 {
			b.WriteString(t.colSep)
		}
	}
	b.WriteByte('\n')

	// rows
	for _, r := range t.Rows {
		b.WriteString(padRight(r.Key, t.max[0]))
		b.WriteString(t.colSep)
		b.WriteString(padLeft(r.Value, t.max[1]))
		b.WriteByte('\n')
	}

	return b.String()
}

// pad from left and right putting the given string to the center
func center(str, filler string, max int) string {
	padLeft := (max - len(str)) / 2
	padRight := padLeft
	if (max-len(str))%2 != 0 {
		padRight++
	}
	b := strings.Builder{}
	b.WriteString(strings.Repeat(filler, padLeft))
	b.WriteString(str)
	b.WriteString(strings.Repeat(filler, padRight))
	return b.String()
}

// pad right up to the given max
func padRight(str string, max int) string {
	pad := max - len(str)
	b := strings.Builder{}
	b.WriteString(str)
	b.WriteString(strings.Repeat(" ", pad))
	return b.String()
}

// pad left up to the given max
func padLeft(str string, max int) string {
	pad := max - len(str)
	b := strings.Builder{}
	b.WriteString(strings.Repeat(" ", pad))
	b.WriteString(str)
	return b.String()
}

func wrapAlert(text string) string {
	b := strings.Builder{}
	b.WriteString("\n[ALERT] ")
	b.WriteString(text)
	b.WriteByte('\n')
	return b.String()
}

func wrapClearAlert(text string) string {
	b := strings.Builder{}
	b.WriteString("\n[CLEAR] ")
	b.WriteString(text)
	b.WriteByte('\n')
	return b.String()
}

func wrapInfo(text string) string {
	b := strings.Builder{}
	b.WriteString("\n[INFO] ")
	b.WriteString(text)
	b.WriteByte('\n')
	return b.String()
}

func wrapError(text string) string {
	b := strings.Builder{}
	b.WriteString("\n[ERR] ")
	b.WriteString(text)
	b.WriteByte('\n')
	return b.String()
}
