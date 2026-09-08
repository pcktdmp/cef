package cefevent

import "strings"

import "testing"

func TestExtensionFieldLimitsKnownValues(t *testing.T) {

	want := map[string]int{
		"act":              63,
		"cs1":              4000,
		"msg":              1023,
		"requestContext":   2048,
		"dvchost":          100,
		"externalId":       40,
		"flexString1Label": 128,
	}

	for key, limit := range want {
		got, ok := ExtensionFieldLimits[key]
		if !ok {
			t.Errorf("ExtensionFieldLimits[%q] missing, want %d", key, limit)
			continue
		}
		if got != limit {
			t.Errorf("ExtensionFieldLimits[%q] = %d, want %d", key, got, limit)
		}
	}
}

func TestValidateExtensionLengthsWithinLimits(t *testing.T) {

	e := event
	e.Extensions = map[string]string{
		"act": strings.Repeat("a", 63),
		"msg": strings.Repeat("m", 1023),
		"src": "127.0.0.1",
	}

	if err := e.ValidateExtensionLengths(); err != nil {
		t.Errorf("ValidateExtensionLengths() = %v, want nil", err)
	}
}

func TestValidateExtensionLengthsExceeded(t *testing.T) {

	e := event
	e.Extensions = map[string]string{
		"act": strings.Repeat("a", 64),
	}

	err := e.ValidateExtensionLengths()
	if err == nil {
		t.Fatalf("ValidateExtensionLengths() = nil, want an error")
	}
	if !strings.Contains(err.Error(), "act") {
		t.Errorf("ValidateExtensionLengths() error = %q, want it to mention %q", err.Error(), "act")
	}
}

func TestValidateExtensionLengthsReportsAllViolations(t *testing.T) {

	e := event
	e.Extensions = map[string]string{
		"act": strings.Repeat("a", 64),
		"app": strings.Repeat("p", 32),
	}

	err := e.ValidateExtensionLengths()
	if err == nil {
		t.Fatalf("ValidateExtensionLengths() = nil, want an error")
	}
	if !strings.Contains(err.Error(), "act") || !strings.Contains(err.Error(), "app") {
		t.Errorf("ValidateExtensionLengths() error = %q, want it to mention both %q and %q", err.Error(), "act", "app")
	}
}

func TestValidateExtensionLengthsIgnoresCustomAndNonStringKeys(t *testing.T) {

	e := event
	e.Extensions = map[string]string{
		"myCustomKey": strings.Repeat("x", 10000),
		"cnt":         "not length-limited, cnt is an Integer type",
	}

	if err := e.ValidateExtensionLengths(); err != nil {
		t.Errorf("ValidateExtensionLengths() = %v, want nil for keys outside ExtensionFieldLimits", err)
	}
}

func TestHeaderFieldLimitsKnownValues(t *testing.T) {

	want := map[string]int{
		"DeviceVendor":       63,
		"DeviceProduct":      63,
		"DeviceVersion":      31,
		"DeviceEventClassId": 1023,
		"Name":               512,
	}

	for key, limit := range want {
		got, ok := HeaderFieldLimits[key]
		if !ok {
			t.Errorf("HeaderFieldLimits[%q] missing, want %d", key, limit)
			continue
		}
		if got != limit {
			t.Errorf("HeaderFieldLimits[%q] = %d, want %d", key, got, limit)
		}
	}

	if _, ok := HeaderFieldLimits["Version"]; ok {
		t.Errorf("HeaderFieldLimits[%q] should not be present: Version has no documented length limit", "Version")
	}
	if _, ok := HeaderFieldLimits["Severity"]; ok {
		t.Errorf("HeaderFieldLimits[%q] should not be present: Severity has no documented length limit", "Severity")
	}
}

func TestValidateHeaderLengthsWithinLimits(t *testing.T) {

	e := event
	e.DeviceVendor = strings.Repeat("v", 63)
	e.DeviceProduct = strings.Repeat("p", 63)
	e.DeviceVersion = strings.Repeat("d", 31)
	e.DeviceEventClassId = strings.Repeat("c", 1023)
	e.Name = strings.Repeat("n", 512)

	if err := e.ValidateHeaderLengths(); err != nil {
		t.Errorf("ValidateHeaderLengths() = %v, want nil", err)
	}
}

func TestValidateHeaderLengthsExceeded(t *testing.T) {

	e := event
	e.DeviceVendor = strings.Repeat("v", 64)

	err := e.ValidateHeaderLengths()
	if err == nil {
		t.Fatalf("ValidateHeaderLengths() = nil, want an error")
	}
	if !strings.Contains(err.Error(), "DeviceVendor") {
		t.Errorf("ValidateHeaderLengths() error = %q, want it to mention %q", err.Error(), "DeviceVendor")
	}
}

func TestValidateHeaderLengthsIgnoresVersionAndSeverity(t *testing.T) {

	e := event
	e.Severity = strings.Repeat("s", 10000)

	if err := e.ValidateHeaderLengths(); err != nil {
		t.Errorf("ValidateHeaderLengths() = %v, want nil: Severity has no documented length limit", err)
	}
}

func TestValidateFieldLengthsCombinesHeaderAndExtensionViolations(t *testing.T) {

	e := event
	e.DeviceVendor = strings.Repeat("v", 64)
	e.Extensions = map[string]string{
		"act": strings.Repeat("a", 64),
	}

	err := e.ValidateFieldLengths()
	if err == nil {
		t.Fatalf("ValidateFieldLengths() = nil, want an error")
	}
	if !strings.Contains(err.Error(), "DeviceVendor") || !strings.Contains(err.Error(), "act") {
		t.Errorf("ValidateFieldLengths() error = %q, want it to mention both %q and %q", err.Error(), "DeviceVendor", "act")
	}
}

func TestValidateFieldLengthsWithinLimits(t *testing.T) {

	e := event
	e.Extensions = map[string]string{"src": "127.0.0.1"}

	if err := e.ValidateFieldLengths(); err != nil {
		t.Errorf("ValidateFieldLengths() = %v, want nil", err)
	}
}

func TestValidateExtensionLengthsCountsRunesNotBytes(t *testing.T) {

	e := event
	// "app" has a documented limit of 31 characters. Use a multi-byte
	// UTF-8 rune repeated 31 times: 31 characters, but more than 31 bytes.
	e.Extensions = map[string]string{
		"app": strings.Repeat("é", 31),
	}

	if err := e.ValidateExtensionLengths(); err != nil {
		t.Errorf("ValidateExtensionLengths() = %v, want nil (31 runes should be within the 31-character limit)", err)
	}

	e.Extensions = map[string]string{
		"app": strings.Repeat("é", 32),
	}

	if err := e.ValidateExtensionLengths(); err == nil {
		t.Errorf("ValidateExtensionLengths() = nil, want an error for 32 runes exceeding the 31-character limit")
	}
}
