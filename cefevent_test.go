package main

import (
	"cefevent/cefevent"
	"testing"
)

func TestCefEvent(t *testing.T) {

	ext := make(map[string]string)
	ext["sourceAddress"] = "127.0.0.1"
	ext["requestClientApplication"] = "Go-http-client/1.1"

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

	want := "CEF:0|Cool Vendor|Cool Product|1.0|COOL_THING|Something cool happened.|Unknown|sourceAddress=127.0.0.1 requestClientApplication=Go-http-client/1.1"
	got := event.Generate()

	if want != got {
		t.Errorf("event.Generate() = %q, want %q", got, want)
	}

}
