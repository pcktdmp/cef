package cefevent

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// LineError describes one line from a ReadAll input that failed to parse as a
// CEF message.
type LineError struct {
	Line int    // 1-indexed line number within the input
	Text string // the raw line content, after any syslog-prefix stripping
	Err  error  // the underlying error Read returned for this line
}

func (e *LineError) Error() string {
	return fmt.Sprintf("line %d: %v", e.Line, e.Err)
}

func (e *LineError) Unwrap() error {
	return e.Err
}

// ReadAll reads CEF messages from r, one per line, using Read to parse each
// non-blank line. It returns every successfully parsed CefEvent, in file
// order.
//
// A line doesn't need to start with "CEF:" itself: if the marker appears
// later in the line (e.g. behind a syslog prefix like
// "Sep 19 08:26:10 host CEF:0|..."), the text before it is stripped before
// parsing, matching how CEF is actually shipped over syslog in practice.
//
// Lines that fail to parse are skipped rather than aborting the whole read;
// each produces a *LineError (identifying its 1-indexed line number) in
// lineErrs, so a caller can decide whether any of them are fatal. Blank lines
// are skipped silently, without an entry in lineErrs.
//
// The returned err is non-nil only for an I/O failure reading from r itself
// — never for a per-line parse failure, which is reported via lineErrs
// instead.
func ReadAll(r io.Reader) (events []CefEvent, lineErrs []*LineError, err error) {

	scanner := bufio.NewScanner(r)
	lineNum := 0

	for scanner.Scan() {
		lineNum++

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if idx := strings.Index(line, "CEF:"); idx > 0 {
			line = line[idx:]
		}

		var e CefEvent
		parsed, readErr := e.Read(line)
		if readErr != nil {
			lineErrs = append(lineErrs, &LineError{Line: lineNum, Text: line, Err: readErr})
			continue
		}

		events = append(events, parsed)
	}

	if scanErr := scanner.Err(); scanErr != nil {
		err = scanErr
	}

	return events, lineErrs, err
}
