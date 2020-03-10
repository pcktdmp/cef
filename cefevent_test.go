package main

import (
	"cefevent/cefevent"
	"testing"
)

var event = cefevent.CefEvent{
	Version:            "0",
	DeviceVendor:       "Cool Vendor",
	DeviceProduct:      "Cool Product",
	DeviceVersion:      "1.0",
	DeviceEventClassId: "COOL_THING",
	Name:               "Something cool happened.",
	Severity:           "Unknown",
}

func TestCefEventExpected(t *testing.T) {

	extLocal := make(map[string]string)
	extLocal["sourceAddress"] = "127.0.0.1"

	expectedEvent := event
	expectedEvent.Extensions = extLocal

	want := "CEF:0|Cool Vendor|Cool Product|1.0|COOL_THING|Something cool happened.|Unknown|sourceAddress=127.0.0.1"
	got, _ := expectedEvent.Generate()

	if want != got {
		t.Errorf("event.Generate() = %q, want %q", got, want)
	}

}

func TestCefEventEscape(t *testing.T) {

	extLocal := make(map[string]string)
	extLocal["sourceAddress\\"] = "\n127.0.0.1="

	borkyEvent := event
	borkyEvent.DeviceVendor = "\\Cool\nVendor|"
	borkyEvent.Extensions = extLocal

	want := "CEF:0|\\\\Cool\\nVendor\\||Cool Product|1.0|COOL_THING|Something cool happened.|Unknown|sourceAddress\\\\=\\n127.0.0.1\\="
	got, _ := borkyEvent.Generate()

	if want != got {
		t.Errorf("event.Generate() = %q, want %q", got, want)
	}

}

func TestCefEventMandatoryVersionField(t *testing.T) {

	brokenEvent := event
	brokenEvent.Version = ""
	_, err := brokenEvent.Generate()

	if err == nil {
		t.Errorf("%v", err)
	}
}

func TestCefEventMandatoryDeviceVendorField(t *testing.T) {

	brokenEvent := event
	brokenEvent.DeviceVendor = ""
	_, err := brokenEvent.Generate()

	if err == nil {
		t.Errorf("%v", err)
	}
}

func TestCefEventMandatoryDeviceProductField(t *testing.T) {

	brokenEvent := event
	brokenEvent.DeviceProduct = ""
	_, err := brokenEvent.Generate()

	if err == nil {
		t.Errorf("%v", err)
	}
}

func TestCefEventMandatoryDeviceVersionField(t *testing.T) {

	brokenEvent := event
	brokenEvent.DeviceVersion = ""
	_, err := brokenEvent.Generate()

	if err == nil {
		t.Errorf("%v", err)
	}
}

func TestCefEventMandatoryDeviceEventClassIdField(t *testing.T) {

	brokenEvent := event
	brokenEvent.DeviceEventClassId = ""
	_, err := brokenEvent.Generate()

	if err == nil {
		t.Errorf("%v", err)
	}
}

func TestCefEventMandatoryNameField(t *testing.T) {

	brokenEvent := event
	brokenEvent.Name = ""
	_, err := brokenEvent.Generate()

	if err == nil {
		t.Errorf("%v", err)
	}
}

func TestCefEventMandatorySeverityField(t *testing.T) {

	brokenEvent := event
	brokenEvent.Severity = ""
	_, err := brokenEvent.Generate()

	if err == nil {
		t.Errorf("%v", err)
	}
}

func someImplementationOfCefEventer(e cefevent.CefEventer) bool {
	return e.Validate()
}

func TestCefEventerValidate(t *testing.T) {

	if !someImplementationOfCefEventer(&event) {
		t.Errorf("Validation should be succesful here.")
	}

	noVersion := event
	noVersion.Version = ""
	if someImplementationOfCefEventer(&noVersion) {
		t.Errorf("Validation should fail here.")
	}
}
