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

func TestCefEventMandatoryVersionField(t *testing.T) {

	event := cefevent.CefEvent{
		Version:            "0",
		DeviceVendor:       "Cool Vendor",
		DeviceProduct:      "Cool Product",
		DeviceVersion:      "1.0",
		DeviceEventClassId: "COOL_THING",
		Name:               "Something cool happened.",
		Severity:           "Unknown",
	}

	brokenEvent := event
	brokenEvent.Version = ""
	_, err := brokenEvent.Generate()

	if err == nil {
		t.Errorf("%v", err)
	}
}

func TestCefEventMandatoryDeviceVendorField(t *testing.T) {

	event := cefevent.CefEvent{
		Version:            "0",
		DeviceVendor:       "Cool Vendor",
		DeviceProduct:      "Cool Product",
		DeviceVersion:      "1.0",
		DeviceEventClassId: "COOL_THING",
		Name:               "Something cool happened.",
		Severity:           "Unknown",
	}

	brokenEvent := event
	brokenEvent.DeviceVendor = ""
	_, err := brokenEvent.Generate()

	if err == nil {
		t.Errorf("%v", err)
	}
}

func TestCefEventMandatoryDeviceProductField(t *testing.T) {

	event := cefevent.CefEvent{
		Version:            "0",
		DeviceVendor:       "Cool Vendor",
		DeviceProduct:      "Cool Product",
		DeviceVersion:      "1.0",
		DeviceEventClassId: "COOL_THING",
		Name:               "Something cool happened.",
		Severity:           "Unknown",
	}

	brokenEvent := event
	brokenEvent.DeviceProduct = ""
	_, err := brokenEvent.Generate()

	if err == nil {
		t.Errorf("%v", err)
	}
}

func TestCefEventMandatoryDeviceVersionField(t *testing.T) {

	event := cefevent.CefEvent{
		Version:            "0",
		DeviceVendor:       "Cool Vendor",
		DeviceProduct:      "Cool Product",
		DeviceVersion:      "1.0",
		DeviceEventClassId: "COOL_THING",
		Name:               "Something cool happened.",
		Severity:           "Unknown",
	}

	brokenEvent := event
	brokenEvent.DeviceVersion = ""
	_, err := brokenEvent.Generate()

	if err == nil {
		t.Errorf("%v", err)
	}
}

func TestCefEventMandatoryDeviceEventClassIdField(t *testing.T) {

	event := cefevent.CefEvent{
		Version:            "0",
		DeviceVendor:       "Cool Vendor",
		DeviceProduct:      "Cool Product",
		DeviceVersion:      "1.0",
		DeviceEventClassId: "COOL_THING",
		Name:               "Something cool happened.",
		Severity:           "Unknown",
	}

	brokenEvent := event
	brokenEvent.DeviceEventClassId = ""
	_, err := brokenEvent.Generate()

	if err == nil {
		t.Errorf("%v", err)
	}
}

func TestCefEventMandatoryNameField(t *testing.T) {

	event := cefevent.CefEvent{
		Version:            "0",
		DeviceVendor:       "Cool Vendor",
		DeviceProduct:      "Cool Product",
		DeviceVersion:      "1.0",
		DeviceEventClassId: "COOL_THING",
		Name:               "Something cool happened.",
		Severity:           "Unknown",
	}

	brokenEvent := event
	brokenEvent.Name = ""
	_, err := brokenEvent.Generate()

	if err == nil {
		t.Errorf("%v", err)
	}
}

func TestCefEventMandatorySeverityField(t *testing.T) {

	event := cefevent.CefEvent{
		Version:            "0",
		DeviceVendor:       "Cool Vendor",
		DeviceProduct:      "Cool Product",
		DeviceVersion:      "1.0",
		DeviceEventClassId: "COOL_THING",
		Name:               "Something cool happened.",
		Severity:           "Unknown",
	}

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

	event := cefevent.CefEvent{
		Version:            "0",
		DeviceVendor:       "Cool Vendor",
		DeviceProduct:      "Cool Product",
		DeviceVersion:      "1.0",
		DeviceEventClassId: "COOL_THING",
		Name:               "Something cool happened.",
		Severity:           "Unknown",
	}

	if !someImplementationOfCefEventer(&event) {
		t.Errorf("Validation should be succesful here.")
	}

	noVersion := event
	noVersion.Version = ""
	if someImplementationOfCefEventer(&noVersion) {
		t.Errorf("Validation should fail here.")
	}
}
