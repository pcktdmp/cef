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

func TestCefEventBuild(t *testing.T) {

	buildEvent := event

	got, err := buildEvent.Build()
	if err != nil {
		t.Fatalf("Build() returned an unexpected error: %v", err)
	}

	want := event
	if !reflect.DeepEqual(want, got) {
		t.Errorf("Build() = %v, want %v", got, want)
	}
}

func TestCefEventBuildMandatoryFieldMissing(t *testing.T) {

	brokenEvent := event
	brokenEvent.DeviceVendor = ""

	got, err := brokenEvent.Build()
	if err == nil {
		t.Errorf("Build() = %v, want an error for a missing mandatory field", got)
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

func TestCefEventParsedNoExtensions(t *testing.T) {

	newEvent := CefEvent{}

	got, err := newEvent.Read("CEF:0|Cool Vendor|Cool Product|1.0|COOL_THING|Something cool happened.|Unknown")

	if err != nil {
		t.Errorf("Read() returned an unexpected error: %v", err)
	}

	if len(got.Extensions) != 0 {
		t.Errorf("Read() Extensions = %v, want empty map", got.Extensions)
	}
}

func TestCefEventParsedTooShort(t *testing.T) {

	newEvent := CefEvent{}

	got, err := newEvent.Read("CEF:0|Cool Vendor|Cool Product")

	if err == nil {
		t.Errorf("Read() = %v, want an error for a message missing mandatory fields", got)
	}
}

func TestCefEventParsedInvalidVersion(t *testing.T) {

	newEvent := CefEvent{}

	got, err := newEvent.Read("CEF:notanumber|Cool Vendor|Cool Product|1.0|COOL_THING|Something cool happened.|Unknown")

	if err == nil {
		t.Errorf("Read() = %v, want an error for a non-numeric Version field", got)
	}
}

func TestCefEventParsedEmptyMandatoryField(t *testing.T) {

	newEvent := CefEvent{}

	got, err := newEvent.Read("CEF:0||Cool Product|1.0|COOL_THING|Something cool happened.|Unknown")

	if err == nil {
		t.Errorf("Read() = %v, want an error for a message with an empty mandatory field (DeviceVendor)", got)
	}
}

func TestCefEventReadUnescapesLikeStringEscapes(t *testing.T) {

	// The value, not the key, carries the special characters here: CEF keys
	// (standard or custom) are always plain identifiers, so it's the value's
	// escape round trip that matters for a realistic message.
	original := event
	original.DeviceVendor = "\\Cool\nVendor|"
	original.Extensions = map[string]string{"cs1": "a\\b\nc=d"}

	// String() no longer mutates its receiver (see
	// TestStringBuildToJSONDoNotMutateReceiver), so original is still the raw,
	// pre-escape value after this call and can be compared against directly.
	want := original

	wire, err := original.String()
	if err != nil {
		t.Fatalf("String() returned an unexpected error: %v", err)
	}

	roundTripped := CefEvent{}
	got, err := roundTripped.Read(wire)
	if err != nil {
		t.Fatalf("Read() returned an unexpected error: %v", err)
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Read(String()) = %+v, want %+v (round trip should be lossless)", got, want)
	}
}

func TestCefEventReadHandlesEscapedPipeInHeaderField(t *testing.T) {

	newEvent := CefEvent{}

	got, err := newEvent.Read("CEF:0|A\\|B|Cool Product|1.0|COOL_THING|Something cool happened.|Unknown")
	if err != nil {
		t.Fatalf("Read() returned an unexpected error: %v", err)
	}

	if got.DeviceVendor != "A|B" {
		t.Errorf("DeviceVendor = %q, want %q", got.DeviceVendor, "A|B")
	}
}

func TestCefEventReadHandlesSpaceInExtensionValue(t *testing.T) {

	newEvent := CefEvent{}

	// straight from the CEF spec's own example of a legitimate space-containing value.
	got, err := newEvent.Read("CEF:0|Cool Vendor|Cool Product|1.0|COOL_THING|Something cool happened.|Unknown|filePath=/user/username/dir/my file name.txt act=block")
	if err != nil {
		t.Fatalf("Read() returned an unexpected error: %v", err)
	}

	want := map[string]string{
		"filePath": "/user/username/dir/my file name.txt",
		"act":      "block",
	}
	if !reflect.DeepEqual(want, got.Extensions) {
		t.Errorf("Extensions = %v, want %v", got.Extensions, want)
	}
}

func TestCefEventReadHandlesUnescapedPipeInExtensionValue(t *testing.T) {

	newEvent := CefEvent{}

	// "|" is only required to be escaped in header fields, not extension values.
	got, err := newEvent.Read("CEF:0|Cool Vendor|Cool Product|1.0|COOL_THING|Something cool happened.|Unknown|msg=a|b")
	if err != nil {
		t.Fatalf("Read() returned an unexpected error: %v", err)
	}

	if got.Extensions["msg"] != "a|b" {
		t.Errorf("Extensions[\"msg\"] = %q, want %q", got.Extensions["msg"], "a|b")
	}
}

func TestCefEventReadTrimsTrailingSpaceOnFinalValueOnly(t *testing.T) {

	newEvent := CefEvent{}

	// Multiple spaces before a key: all but the last belong to the prior value.
	// Trailing spaces on the very last value, though, are dropped.
	got, err := newEvent.Read("CEF:0|Cool Vendor|Cool Product|1.0|COOL_THING|Something cool happened.|Unknown|a=one   act=two  ")
	if err != nil {
		t.Fatalf("Read() returned an unexpected error: %v", err)
	}

	want := map[string]string{
		"a":   "one  ",
		"act": "two",
	}
	if !reflect.DeepEqual(want, got.Extensions) {
		t.Errorf("Extensions = %v, want %v", got.Extensions, want)
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

func TestStringBuildToJSONDoNotMutateReceiver(t *testing.T) {

	original := event
	original.DeviceVendor = "\\Cool\nVendor|"
	original.Extensions = map[string]string{"cs1": "a\\b\nc=d"}

	snapshot := original
	snapshot.Extensions = map[string]string{"cs1": "a\\b\nc=d"}

	e := original
	if _, err := e.String(); err != nil {
		t.Fatalf("String() returned an unexpected error: %v", err)
	}
	if !reflect.DeepEqual(e, snapshot) {
		t.Errorf("String() mutated its receiver: got %+v, want unchanged %+v", e, snapshot)
	}

	e = original
	if _, err := e.Build(); err != nil {
		t.Fatalf("Build() returned an unexpected error: %v", err)
	}
	if !reflect.DeepEqual(e, snapshot) {
		t.Errorf("Build() mutated its receiver: got %+v, want unchanged %+v", e, snapshot)
	}

	e = original
	if _, err := e.ToJSON(); err != nil {
		t.Fatalf("ToJSON() returned an unexpected error: %v", err)
	}
	if !reflect.DeepEqual(e, snapshot) {
		t.Errorf("ToJSON() mutated its receiver: got %+v, want unchanged %+v", e, snapshot)
	}
}

func TestStringCalledTwiceDoesNotDoubleEscape(t *testing.T) {

	e := event
	e.DeviceVendor = "\\Cool\nVendor|"
	e.Extensions = map[string]string{"cs1": "a\\b\nc=d"}

	first, err := e.String()
	if err != nil {
		t.Fatalf("String() returned an unexpected error: %v", err)
	}

	second, err := e.String()
	if err != nil {
		t.Fatalf("String() (second call) returned an unexpected error: %v", err)
	}

	if first != second {
		t.Errorf("String() called twice produced different output: first=%q second=%q (should be identical - escaping should not accumulate)", first, second)
	}
}

func TestBuildCalledTwiceDoesNotDoubleEscape(t *testing.T) {

	e := event
	e.DeviceVendor = "\\Cool\nVendor|"
	e.Extensions = map[string]string{"cs1": "a\\b\nc=d"}

	first, err := e.Build()
	if err != nil {
		t.Fatalf("Build() returned an unexpected error: %v", err)
	}

	second, err := e.Build()
	if err != nil {
		t.Fatalf("Build() (second call) returned an unexpected error: %v", err)
	}

	if !reflect.DeepEqual(first, second) {
		t.Errorf("Build() called twice produced different output: first=%+v second=%+v (should be identical - escaping should not accumulate)", first, second)
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

func TestCefEvent_ToJSONEscapesLikeString(t *testing.T) {

	borkyEvent := event
	borkyEvent.DeviceVendor = "\\Cool\nVendor|"
	borkyEvent.Extensions = map[string]string{"broken_src\\": "\n127.0.0.2="}

	got, err := borkyEvent.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() returned an unexpected error: %v", err)
	}

	want := `{"Version":0,"DeviceVendor":"\\\\Cool\\nVendor\\|","DeviceProduct":"Cool Product","DeviceVersion":"1.0","DeviceEventClassId":"COOL_THING","Name":"Something cool happened.","Severity":"Unknown","Extensions":{"broken_src\\\\":"\\n127.0.0.2\\="}}`

	if got != want {
		t.Errorf("ToJSON() = %q, want %q", got, want)
	}
}
