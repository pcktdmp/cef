package cefevent

import (
	"errors"
	"strings"
	"testing"
)

func TestReadAllParsesMultipleLines(t *testing.T) {

	input := strings.Join([]string{
		"CEF:0|Cool Vendor|Cool Product|1.0|COOL_THING|Something cool happened.|Unknown|src=127.0.0.1",
		"CEF:0|Cool Vendor|Cool Product|1.0|COOL_THING|Something else happened.|Low|src=10.0.0.1",
	}, "\n")

	events, lineErrs, err := ReadAll(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ReadAll() returned an unexpected error: %v", err)
	}
	if len(lineErrs) != 0 {
		t.Fatalf("ReadAll() lineErrs = %v, want none", lineErrs)
	}
	if len(events) != 2 {
		t.Fatalf("ReadAll() returned %d events, want 2", len(events))
	}
	if events[0].Name != "Something cool happened." || events[1].Name != "Something else happened." {
		t.Errorf("ReadAll() events out of order or wrong: %+v", events)
	}
}

func TestReadAllSkipsBlankLines(t *testing.T) {

	input := "\nCEF:0|Cool Vendor|Cool Product|1.0|COOL_THING|Something cool happened.|Unknown|src=127.0.0.1\n\n   \n"

	events, lineErrs, err := ReadAll(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ReadAll() returned an unexpected error: %v", err)
	}
	if len(lineErrs) != 0 {
		t.Errorf("ReadAll() lineErrs = %v, want none (blank lines aren't errors)", lineErrs)
	}
	if len(events) != 1 {
		t.Fatalf("ReadAll() returned %d events, want 1", len(events))
	}
}

func TestReadAllStripsSyslogPrefix(t *testing.T) {

	input := "Sep 19 08:26:10 host CEF:0|Security|threatmanager|1.0|100|worm successfully stopped|10|src=10.0.0.1 dst=2.1.2.2 spt=1232"

	events, lineErrs, err := ReadAll(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ReadAll() returned an unexpected error: %v", err)
	}
	if len(lineErrs) != 0 {
		t.Fatalf("ReadAll() lineErrs = %v, want none", lineErrs)
	}
	if len(events) != 1 {
		t.Fatalf("ReadAll() returned %d events, want 1", len(events))
	}
	if events[0].DeviceVendor != "Security" || events[0].Extensions["dst"] != "2.1.2.2" {
		t.Errorf("ReadAll() with syslog prefix = %+v, want DeviceVendor=Security, Extensions[dst]=2.1.2.2", events[0])
	}
}

func TestReadAllCollectsPerLineErrorsWithoutAborting(t *testing.T) {

	input := strings.Join([]string{
		"CEF:0|Cool Vendor|Cool Product|1.0|COOL_THING|Something cool happened.|Unknown|src=127.0.0.1",
		"this is not a CEF line at all",
		"CEF:0|Cool Vendor|Cool Product|1.0|COOL_THING|Something else happened.|Low|src=10.0.0.1",
	}, "\n")

	events, lineErrs, err := ReadAll(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ReadAll() returned an unexpected error: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("ReadAll() returned %d events, want 2 (the bad line should be skipped, not fatal)", len(events))
	}
	if len(lineErrs) != 1 {
		t.Fatalf("ReadAll() lineErrs = %v, want exactly 1", lineErrs)
	}
	if lineErrs[0].Line != 2 {
		t.Errorf("lineErrs[0].Line = %d, want 2", lineErrs[0].Line)
	}
	if lineErrs[0].Text != "this is not a CEF line at all" {
		t.Errorf("lineErrs[0].Text = %q, want %q", lineErrs[0].Text, "this is not a CEF line at all")
	}
}

func TestLineErrorUnwrap(t *testing.T) {

	underlying := errors.New("boom")
	le := &LineError{Line: 3, Text: "whatever", Err: underlying}

	if !errors.Is(le, underlying) {
		t.Errorf("errors.Is(le, underlying) = false, want true (LineError should unwrap)")
	}
	if got := le.Error(); got != "line 3: boom" {
		t.Errorf("LineError.Error() = %q, want %q", got, "line 3: boom")
	}
}

func TestReadAllEmptyInput(t *testing.T) {

	events, lineErrs, err := ReadAll(strings.NewReader(""))
	if err != nil {
		t.Fatalf("ReadAll() returned an unexpected error: %v", err)
	}
	if len(events) != 0 || len(lineErrs) != 0 {
		t.Errorf("ReadAll(\"\") = events:%v lineErrs:%v, want both empty", events, lineErrs)
	}
}
