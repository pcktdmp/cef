package main

import (
	"reflect"
	"testing"

	"github.com/pcktdmp/cef/cefevent"
)

var event = cefevent.CefEvent{
	Version:            0,
	DeviceVendor:       "Cool Vendor",
	DeviceProduct:      "Cool Product",
	DeviceVersion:      "1.0",
	DeviceEventClassId: "COOL_THING",
	Name:               "Something cool happened.",
	Severity:           "Unknown",
	Extensions:         map[string]string{"src": "127.0.0.1"},
}

var eventLine = ("CEF:0|Cool Vendor|Cool Product|1.0|COOL_THING|Something cool happened.|Unknown|src=127.0.0.1")

func TestCefEventExpected(t *testing.T) {

	expectedEvent := event

	want := "CEF:0|Cool Vendor|Cool Product|1.0|COOL_THING|Something cool happened.|Unknown|src=127.0.0.1"
	got, _ := expectedEvent.Generate()

	if want != got {
		t.Errorf("event.Generate() = %q, want %q", got, want)
	}

}

func TestCefEventParsed(t *testing.T) {

	want := event
	got, _ := cefevent.Parse(eventLine)

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Parse() = %v, want %v", got, want)
	}
}

func TestCefEventParsedFail(t *testing.T) {

	got, err := cefevent.Parse("This should definitely fail.")

	if err == nil {
		t.Errorf("Parse() = %v, want %v", err, got)
	}
}

func TestCefEventEscape(t *testing.T) {

	extLocal := make(map[string]string)
	extLocal["broken_src\\"] = "\n127.0.0.2="

	borkyEvent := event
	borkyEvent.DeviceVendor = "\\Cool\nVendor|"
	borkyEvent.Extensions = extLocal

	want := "CEF:0|\\\\Cool\\nVendor\\||Cool Product|1.0|COOL_THING|Something cool happened.|Unknown|broken_src\\\\=\\n127.0.0.2\\="
	got, _ := borkyEvent.Generate()

	if want != got {
		t.Errorf("event.Generate() = %q, want %q", got, want)
	}

}

func TestCefEventMandatoryVersionField(t *testing.T) {

	brokenEvent := event
	brokenEvent.DeviceVendor = ""
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

	noDeviceVendor := event
	noDeviceVendor.DeviceVendor = ""
	if someImplementationOfCefEventer(&noDeviceVendor) {
		t.Errorf("Validation should fail here.")
	}
}

func TestCefEventerLoggingSuccess(t *testing.T) {

	_, err := event.Log()

	if err != nil {
		t.Errorf("%v", err)
	}
}

func TestCefEventerLoggingFail(t *testing.T) {

	brokenEvent := event
	brokenEvent.DeviceVendor = ""
	_, err := brokenEvent.Log()

	if err == nil {
		t.Errorf("%v", err)
	}
}
