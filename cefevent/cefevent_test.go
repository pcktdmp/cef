package cefevent

import (
	"reflect"
	"testing"
)

var event = CefEvent{
	Version:            0,
	DeviceVendor:       "Cool Vendor",
	DeviceProduct:      "Cool Product",
	DeviceVersion:      "1.0",
	DeviceEventClassId: "COOL_THING",
	Name:               "Something cool happened.",
	Severity:           "Unknown",
	Extensions:         map[string]string{"src": "127.0.0.1"},
}

var eventLine = "CEF:0|Cool Vendor|Cool Product|1.0|COOL_THING|Something cool happened.|Unknown|src=127.0.0.1"

func TestCefEventExpected(t *testing.T) {

	expectedEvent := event

	want := "CEF:0|Cool Vendor|Cool Product|1.0|COOL_THING|Something cool happened.|Unknown|src=127.0.0.1"
	got, _ := expectedEvent.String()

	if want != got {
		t.Errorf("event.String() = %q, want %q", got, want)
	}

}

func TestCefEventParsed(t *testing.T) {

	newEvent := CefEvent{}
	want := event
	got, _ := newEvent.Read(eventLine)

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Parse() = %v, want %v", got, want)
	}
}

func TestCefEventParsedAndGenerated(t *testing.T) {

	newEvent := CefEvent{}
	want := eventLine
	parsedEvent, _ := newEvent.Read(eventLine)
	got, _ := parsedEvent.String()

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Parse() = %v, want %v", got, want)
	}
}

func TestCefEventParsedFail(t *testing.T) {

	newEvent := CefEvent{}

	got, err := newEvent.Read("This should definitely fail.")

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
	got, _ := borkyEvent.String()

	if want != got {
		t.Errorf("event.String() = %q, want %q", got, want)
	}

}

func TestCefEventMandatoryVersionField(t *testing.T) {

	brokenEvent := event
	brokenEvent.DeviceVendor = ""
	_, err := brokenEvent.String()

	if err == nil {
		t.Errorf("%v", err)
	}
}

func TestCefEventMandatoryDeviceVendorField(t *testing.T) {

	brokenEvent := event
	brokenEvent.DeviceVendor = ""
	_, err := brokenEvent.String()

	if err == nil {
		t.Errorf("%v", err)
	}
}

func TestCefEventMandatoryDeviceProductField(t *testing.T) {

	brokenEvent := event
	brokenEvent.DeviceProduct = ""
	_, err := brokenEvent.String()

	if err == nil {
		t.Errorf("%v", err)
	}
}

func TestCefEventMandatoryDeviceVersionField(t *testing.T) {

	brokenEvent := event
	brokenEvent.DeviceVersion = ""
	_, err := brokenEvent.String()

	if err == nil {
		t.Errorf("%v", err)
	}
}

func TestCefEventMandatoryDeviceEventClassIdField(t *testing.T) {

	brokenEvent := event
	brokenEvent.DeviceEventClassId = ""
	_, err := brokenEvent.String()

	if err == nil {
		t.Errorf("%v", err)
	}
}

func TestCefEventMandatoryNameField(t *testing.T) {

	brokenEvent := event
	brokenEvent.Name = ""
	_, err := brokenEvent.String()

	if err == nil {
		t.Errorf("%v", err)
	}
}

func TestCefEventMandatorySeverityField(t *testing.T) {

	brokenEvent := event
	brokenEvent.Severity = ""
	_, err := brokenEvent.String()

	if err == nil {
		t.Errorf("%v", err)
	}
}

func someImplementationOfCefEventer(e CefEventer) error {
	return e.Validate()
}

func TestCefEventerValidate(t *testing.T) {

	if someImplementationOfCefEventer(&event) != nil {
		t.Errorf("Validation should be succesful here.")
	}

	noDeviceVendor := event
	noDeviceVendor.DeviceVendor = ""
	if someImplementationOfCefEventer(&noDeviceVendor) == nil {
		t.Errorf("Validation should fail here.")
	}
}

func TestCefEventerLoggingSuccess(t *testing.T) {

	err := event.Log()

	if err != nil {
		t.Errorf("%v", err)
	}
}

func TestCefEventerLoggingFail(t *testing.T) {

	brokenEvent := event
	brokenEvent.DeviceVendor = ""
	err := brokenEvent.Log()

	if err == nil {
		t.Errorf("%v", err)
	}
}

func TestCefEvent_ToJSON(t *testing.T) {
	var tests = []struct {
		cev      CefEvent
		want     string
		hasError bool
	}{
		{
			cev: CefEvent{
				Version:            1,
				DeviceVendor:       "Test Vendor",
				DeviceProduct:      "Test Product",
				DeviceVersion:      "1.0.0",
				DeviceEventClassId: "Test Class ID",
				Name:               "Test Name",
				Severity:           "Test Severity",
				Extensions:         map[string]string{"Extension1": "Value1", "Extension2": "Value2"},
			},
			want:     `{"Version":1,"DeviceVendor":"Test Vendor","DeviceProduct":"Test Product","DeviceVersion":"1.0.0","DeviceEventClassId":"Test Class ID","Name":"Test Name","Severity":"Test Severity","Extensions":{"Extension1":"Value1","Extension2":"Value2"}}`,
			hasError: false,
		},
		{
			cev: CefEvent{
				Version:            1,
				DeviceVendor:       "",
				DeviceProduct:      "Test Product",
				DeviceVersion:      "1.0.0",
				DeviceEventClassId: "Test Class ID",
				Name:               "Test Name",
				Severity:           "Test Severity",
				Extensions:         map[string]string{"Extension1": "Value1", "Extension2": "Value2"},
			},
			want:     "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		got, err := tt.cev.ToJSON()
		if (err != nil) != tt.hasError {
			t.Errorf("Expected error status: %v, got: %v", tt.hasError, err)
		}
		if got != tt.want {
			t.Errorf("Expected json `%v`, but got `%v`", tt.want, got)
		}
	}
}
