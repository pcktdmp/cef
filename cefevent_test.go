package main

import (
	"cefevent/cefevent"
	"testing"
)

func TestCefEventExpected(t *testing.T) {

	ext := make(map[string]string)
	ext["sourceAddress"] = "127.0.0.1"

	event := cefevent.CefEvent{
		Version:            "0",
		DeviceVendor:       "Cool Vendor",
		DeviceProduct:      "Cool Product",
		DeviceVersion:      "1.0",
		DeviceEventClassId: "COOL_THING",
		Name:               "Something cool happened.",
		Severity:           "Unknown",
		Extensions:         ext,
	}

	want := "CEF:0|Cool Vendor|Cool Product|1.0|COOL_THING|Something cool happened.|Unknown|sourceAddress=127.0.0.1"
	got, _ := event.Generate()

	if want != got {
		t.Errorf("event.Generate() = %q, want %q", got, want)
	}

}

func TestCefEventEscape(t *testing.T) {

	ext := make(map[string]string)
	ext["sourceAddress\\"] = "\n127.0.0.1="

	event := cefevent.CefEvent{
		Version:            "0",
		DeviceVendor:       "\\Cool\nVendor|",
		DeviceProduct:      "Cool Product",
		DeviceVersion:      "1.0",
		DeviceEventClassId: "COOL_THING",
		Name:               "Something cool happened.",
		Severity:           "Unknown",
		Extensions:         ext,
	}

	want := "CEF:0|\\\\Cool\\nVendor\\||Cool Product|1.0|COOL_THING|Something cool happened.|Unknown|sourceAddress\\\\=\\n127.0.0.1\\="
	got, _ := event.Generate()

	if want != got {
		t.Errorf("event.Generate() = %q, want %q", got, want)
	}

}

func TestCefEventMandatoryFields(t *testing.T) {

	event := cefevent.CefEvent{
		Version:            "0",
		DeviceVendor:       "Cool Vendor",
		DeviceProduct:      "Cool Product",
		DeviceVersion:      "1.0",
		DeviceEventClassId: "COOL_THING",
		Name:               "Something cool happened.",
		Severity:           "Unknown",
	}

	noVersion := event
	noVersion.Version = ""
	_, err := noVersion.Generate()

	if err == nil {
		t.Errorf("%v", err)
	}
}
