package humanize

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/Gastove/humanize/internal/terminal"
)

// Follow Logrus' colorings
const (
	red    = 31
	green  = 32
	yellow = 33
	blue   = 36
	grey   = 37
)

// Log-level configuration from env vars
const (
	HumanizeFormatVar = "HUMANIZE"
	FormatFull        = "FULL"
	FormatCompact     = "COMPACT"
	FormatJSON        = "JSON"
)

var DefaultLevelColors = map[logrus.Level]int{
	logrus.DebugLevel: blue,
	logrus.ErrorLevel: red,
	logrus.FatalLevel: red,
	logrus.WarnLevel:  yellow,
	logrus.TraceLevel: grey,
	logrus.PanicLevel: red,
}

var ElementColors = map[string]int{
	"date":   grey,
	"time":   blue,
	"caller": green,
}

// Formatter renders Logrus log lines in a nicer format for humans.
type Formatter struct {
	// It's possible you don't want ISO-8601. You're wrong, but... OK.
	DateTimeFormat string
	// Take up more whitespace or less.
	Compact bool
	// We assume you wont want to ship logs to logstash or what have you in this
	// format. What do we use when a human isn't looking?
	Fallback logrus.Formatter
	// Info about a potential terminal we're displaying to
	termInfo terminal.TermInfo
	// Sync
	terminalInitOnce sync.Once
}

func (fmttr *Formatter) initTermInfo(entry *logrus.Entry) error {
	if entry.Logger != nil {
		termInfo, err := terminal.GetTermInfo(entry.Logger.Out)

		if err != nil {
			return err
		}

		fmttr.termInfo = termInfo
	}
	return nil
}

// NewHumanizeFormatter creates a new formatter with defaults set.
func NewHumanizeFormatter() *Formatter {
	return &Formatter{
		Compact:        false,
		DateTimeFormat: "2006-01-02T15:04:05",
		Fallback:       &logrus.JSONFormatter{},
	}
}

func parseFormatFromEnv() (string, error) {
	envVal, present := os.LookupEnv(HumanizeFormatVar)
	if present {
		switch strings.ToUpper(envVal) {
		case FormatFull:
			return FormatFull, nil
		case FormatCompact:
			return FormatCompact, nil
		case FormatJSON:
			return FormatJSON, nil
		default:
			err := fmt.Errorf("No such format as %s", envVal)
			return "", err
		}
	}

	err := fmt.Errorf(
		"%s not set in env, no format value could be read",
		HumanizeFormatVar,
	)

	return "", err
}

// NewHumanizeFormatterFromEnv reads a configuration variable from the
// environment, using it to configure and return a logger. Humanize looks for
// its configuration using the value of the `HumanizeFormatVar` constant, and
// can respond to one of three values:
//   1. FULL    -> Format using long formatting (this is the default).
//   2. COMPACT -> Format using compact formatting.
//   3. JSON    -> Return the default Fallback formatter, logrus.JSONFormatter
func NewHumanizeFormatterFromEnv() (logrus.Formatter, error) {
	defaultFormatter := NewHumanizeFormatter()

	providedFormat, err := parseFormatFromEnv()

	if err != nil {
		return defaultFormatter, err
	}

	switch providedFormat {
	case FormatCompact:
		defaultFormatter.Compact = true
		return defaultFormatter, nil
	case FormatFull:
		defaultFormatter.Compact = false
		return defaultFormatter, nil
	case FormatJSON:
		return defaultFormatter.Fallback, nil
	}

	return defaultFormatter, nil
}

// Format implements the Formatter interfact for our Formatter.
func (fmttr *Formatter) Format(entry *logrus.Entry) ([]byte, error) {

	fmttr.terminalInitOnce.Do(func() {
		err := fmttr.initTermInfo(entry)
		if err != nil {
			fmt.Printf("Failed to initialize terminal with err %s", err)
		}
	})

	timeFormat := fmttr.DateTimeFormat

	tpl := "\n%s [%s]: %s"

	// Get the three fields that _aren't_ part of entry.Data
	// Every time you format a date in Golang, a kitten weeps
	ts := entry.Time.Format(timeFormat)
	msg := entry.Message
	level := entry.Level

	line := fmt.Sprintf(tpl, ts, level, msg)
	fields, err := fmttr.renderFields(entry)
	if err != nil {
		return []byte{}, err
	}

	errMsg, err := fmttr.renderError(entry)
	if err != nil {
		return []byte{}, err
	}

	return []byte(line + fields + errMsg), nil
}

func (fmttr *Formatter) renderFields(entry *logrus.Entry) (string, error) {
	// No sense in rendering if there are no fields
	if len(entry.Data) == 0 {
		return "", nil
	}

	compact := fmttr.Compact || false

	if compact {
		return fmttr.renderFieldsCompact(entry)
	}

	return fmttr.renderFieldsLong(entry)
}

func newLineWithOffset(offset int) string {
	return "\n" + strings.Repeat(" ", offset)
}

func fieldOrder(entry *logrus.Entry) []string {
	fields := []string{}

	for field, _ := range entry.Data {
		fields = append(fields, field)
	}

	sort.Strings(fields)

	return fields
}

//------------------------------- Render Fields -------------------------------//
// renderFieldsCompact:
// [iso8061 time] [level] [caller?]: [msg]
//     key1: value1    key2: value2    key3: value3 | Wraps to width of TTY
//     error?: wraps to width of TTY
func (fmttr *Formatter) renderFieldsCompact(entry *logrus.Entry) (string, error) {
	offset := 4
	// fieldPadding := 2 // each key is followed by a colon and a single space
	// maxFieldWidth := longestKeyLen(entry.Data) + fieldPadding

	wrapWidth := fmttr.termInfo.WidthCols

	lines := newLineWithOffset(offset)
	currentLine := strings.Repeat(" ", offset)

	orderedFields := fieldOrder(entry)

	for _, field := range orderedFields {
		if field == "error" {
			continue
		}

		value := entry.Data[field]

		line := fmt.Sprintf("%s: %v", field, value)

		if len(line) > wrapWidth {
			// TODO: Handle wrapping a very long field
		}

		if len(line)+len(currentLine) > wrapWidth {
			lines = lines + currentLine
			currentLine = newLineWithOffset(offset) + line
		} else {
			currentLine = currentLine + "\t" + line
		}
	}

	if !(currentLine == "") {
		lines = lines + currentLine
	}

	return lines, nil
}

// renderFieldsLong:
// [iso8061 time] [level] [caller?]: [msg]
//         Fields:
//                key1: value1
//                key2: value2
//                key3: long values should wrap at the width of the TTY, and
//                      should be nicely indented
//         error?: with the exception
//                 of error
//                 which should be nicely indented
//                 but not wrapped
func (fmttr *Formatter) renderFieldsLong(entry *logrus.Entry) (string, error) {
	offset := 1 + len(fmttr.DateTimeFormat)
	spacer := strings.Repeat(" ", offset)

	orderedFields := fieldOrder(entry)

	fieldPadding := 2 // each key is followed by a colon and a single space
	maxFieldWidth := longestKeyLen(entry.Data) + fieldPadding

	// We'll build everything up into a single variable, `lines`
	// First, we newline off the previous line, do half spacing, and give a section title:
	lines := "\n" + strings.Repeat(" ", int(offset/2)) + "Fields:"

	// Now the fields -- all except error, which we'll render separately.
	for _, field := range orderedFields {
		if field == "error" {
			continue
		}

		value := entry.Data[field]

		fieldPadding := strings.Repeat(" ", maxFieldWidth-len(field))
		// TODO: wrap long values
		line := fmt.Sprintf("\n%s%s: %s%v", spacer, field, fieldPadding, value)
		lines = lines + line
	}
	return lines, nil
}

func (fmttr *Formatter) renderError(entry *logrus.Entry) (string, error) {
	error, hasError := entry.Data["error"]
	if hasError {
		return fmt.Sprintf("\nERROR: %v", error), nil
	}

	return "", nil
}

func longestKeyLen(m map[string]interface{}) int {
	longest := 0
	for k := range m {
		if len(k) > longest {
			longest = len(k)
		}
	}
	return longest
}

func colorizeField(color int, field string) string {
	return fmt.Sprintf(" \x1b[%dm%s\x1b[0m=", color, field)
}
