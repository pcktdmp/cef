package cefevent

import (
	"strings"
	"testing"
)

func TestExtensionFieldTypeString(t *testing.T) {

	want := map[ExtensionFieldType]string{
		ExtensionTypeInteger:    "Integer",
		ExtensionTypeLong:       "Long",
		ExtensionTypeFloat:      "Floating Point",
		ExtensionTypeIPAddress:  "IP Address",
		ExtensionTypeMACAddress: "MAC Address",
		ExtensionTypeTimestamp:  "Time Stamp",
	}

	for typ, name := range want {
		if got := typ.String(); got != name {
			t.Errorf("%v.String() = %q, want %q", int(typ), got, name)
		}
	}
}

func TestValidateExtensionTypesValid(t *testing.T) {

	e := event
	e.Extensions = map[string]string{
		"dst":  "192.168.1.1",
		"dpt":  "443",
		"dmac": "00:0D:60:AF:1B:61",
		"rt":   "1700000000000",
		"cn1":  "42",
		"cfp1": "3.14",
	}

	if err := e.ValidateExtensionTypes(); err != nil {
		t.Errorf("ValidateExtensionTypes() = %v, want nil", err)
	}
}

func TestValidateExtensionTypesValidTimestampLayout(t *testing.T) {

	e := event
	e.Extensions = map[string]string{"start": "Jan 02 2024 15:04:05"}

	if err := e.ValidateExtensionTypes(); err != nil {
		t.Errorf("ValidateExtensionTypes() = %v, want nil for a spec-format timestamp", err)
	}

	e.Extensions = map[string]string{"start": "Jan 2 2024 15:04:05"}

	if err := e.ValidateExtensionTypes(); err != nil {
		t.Errorf("ValidateExtensionTypes() = %v, want nil for a non-zero-padded-day timestamp", err)
	}
}

func TestValidateExtensionTypesInvalidIP(t *testing.T) {

	e := event
	e.Extensions = map[string]string{"dst": "not-an-ip"}

	err := e.ValidateExtensionTypes()
	if err == nil {
		t.Fatalf("ValidateExtensionTypes() = nil, want an error for an invalid IP address")
	}
	if !strings.Contains(err.Error(), "dst") {
		t.Errorf("ValidateExtensionTypes() error = %q, want it to mention %q", err.Error(), "dst")
	}
}

func TestValidateExtensionTypesInvalidInteger(t *testing.T) {

	e := event
	e.Extensions = map[string]string{"dpt": "not-a-port"}

	if err := e.ValidateExtensionTypes(); err == nil {
		t.Errorf("ValidateExtensionTypes() = nil, want an error for a non-numeric dpt")
	}
}

func TestValidateExtensionTypesInvalidMAC(t *testing.T) {

	e := event
	e.Extensions = map[string]string{"dmac": "not-a-mac"}

	if err := e.ValidateExtensionTypes(); err == nil {
		t.Errorf("ValidateExtensionTypes() = nil, want an error for an invalid MAC address")
	}
}

func TestValidateExtensionTypesInvalidTimestamp(t *testing.T) {

	e := event
	e.Extensions = map[string]string{"rt": "not-a-timestamp"}

	if err := e.ValidateExtensionTypes(); err == nil {
		t.Errorf("ValidateExtensionTypes() = nil, want an error for an invalid timestamp")
	}
}

func TestValidateExtensionTypesIPv6Accepted(t *testing.T) {

	e := event
	// CEF 1.0+ allows IPv6 in fields that were IPv4-only in CEF 0.1.
	e.Extensions = map[string]string{"src": "2001:db8::1"}

	if err := e.ValidateExtensionTypes(); err != nil {
		t.Errorf("ValidateExtensionTypes() = %v, want nil: IPv6 is valid for src from CEF 1.0 onward", err)
	}
}

func TestValidateExtensionTypesIgnoresUnknownAndStringKeys(t *testing.T) {

	e := event
	e.Extensions = map[string]string{
		"act":         "block",     // String-typed, not in ExtensionFieldTypes
		"myCustomKey": "whatever!", // custom key, not in the dictionary at all
	}

	if err := e.ValidateExtensionTypes(); err != nil {
		t.Errorf("ValidateExtensionTypes() = %v, want nil for keys outside ExtensionFieldTypes", err)
	}
}

func TestValidateExtensionTypesReportsAllViolations(t *testing.T) {

	e := event
	e.Extensions = map[string]string{
		"dst": "not-an-ip",
		"dpt": "not-a-port",
	}

	err := e.ValidateExtensionTypes()
	if err == nil {
		t.Fatalf("ValidateExtensionTypes() = nil, want an error")
	}
	if !strings.Contains(err.Error(), "dst") || !strings.Contains(err.Error(), "dpt") {
		t.Errorf("ValidateExtensionTypes() error = %q, want it to mention both %q and %q", err.Error(), "dst", "dpt")
	}
}
