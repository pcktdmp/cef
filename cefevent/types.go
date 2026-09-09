package cefevent

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ExtensionFieldType identifies one of the CEF Extension Dictionary's non-String
// data types that ValidateExtensionTypes knows how to check a value against.
type ExtensionFieldType int

const (
	extensionTypeUnknown ExtensionFieldType = iota
	ExtensionTypeInteger
	ExtensionTypeLong
	ExtensionTypeFloat
	ExtensionTypeIPAddress
	ExtensionTypeMACAddress
	ExtensionTypeTimestamp
)

// String returns the CEF spec's own name for the data type, as used in its
// "Data Type" column (e.g. "IP Address", "Time Stamp").
func (t ExtensionFieldType) String() string {
	switch t {
	case ExtensionTypeInteger:
		return "Integer"
	case ExtensionTypeLong:
		return "Long"
	case ExtensionTypeFloat:
		return "Floating Point"
	case ExtensionTypeIPAddress:
		return "IP Address"
	case ExtensionTypeMACAddress:
		return "MAC Address"
	case ExtensionTypeTimestamp:
		return "Time Stamp"
	default:
		return "unknown"
	}
}

// ExtensionFieldTypes maps CEF Extension Dictionary keys with a non-String data
// type to that type, for the well-known keys most commonly used in practice
// (e.g. dst/src as IP addresses, dpt/spt as integers, dmac/smac as MAC
// addresses, rt/start/end as timestamps).
//
// This is a curated subset, not an exhaustive transcription of every
// non-String key in the dictionary — each entry here was individually
// confirmed against the spec's own Data Type column (see CEF-SPEC.md).
// String-typed keys aren't listed here at all; see ExtensionFieldLimits for
// those. A key absent from this map is simply not checked by
// ValidateExtensionTypes, the same way a key absent from ExtensionFieldLimits
// is not checked by ValidateExtensionLengths.
var ExtensionFieldTypes = map[string]ExtensionFieldType{
	// Integer
	"cnt":             ExtensionTypeInteger,
	"deviceDirection": ExtensionTypeInteger,
	"dpid":            ExtensionTypeInteger,
	"dpt":             ExtensionTypeInteger,
	"dvcpid":          ExtensionTypeInteger,
	"fsize":           ExtensionTypeInteger,
	"in":              ExtensionTypeInteger,
	"oldFileSize":     ExtensionTypeInteger,
	"out":             ExtensionTypeInteger,
	"spid":            ExtensionTypeInteger,
	"spt":             ExtensionTypeInteger,
	"type":            ExtensionTypeInteger,

	// Long
	"cn1":     ExtensionTypeLong,
	"cn2":     ExtensionTypeLong,
	"cn3":     ExtensionTypeLong,
	"eventId": ExtensionTypeLong,

	// Floating Point / Double (the spec distinguishes them but both parse the
	// same way as a Go float64, so both map to ExtensionTypeFloat)
	"cfp1":  ExtensionTypeFloat,
	"cfp2":  ExtensionTypeFloat,
	"cfp3":  ExtensionTypeFloat,
	"cfp4":  ExtensionTypeFloat,
	"dlat":  ExtensionTypeFloat,
	"dlong": ExtensionTypeFloat,
	"slat":  ExtensionTypeFloat,
	"slong": ExtensionTypeFloat,

	// IP Address. The spec notes that in CEF 0.1 these fields only ever held
	// IPv4 addresses, but from CEF 1.0 onward they may hold IPv6 too — this
	// accepts either, via net.ParseIP, rather than rejecting valid IPv6 values
	// on the assumption a message is CEF 0.1.
	"agt": ExtensionTypeIPAddress,
	"dst": ExtensionTypeIPAddress,
	"dvc": ExtensionTypeIPAddress,
	"src": ExtensionTypeIPAddress,

	// MAC Address
	"amac":   ExtensionTypeMACAddress,
	"dmac":   ExtensionTypeMACAddress,
	"dvcmac": ExtensionTypeMACAddress,
	"smac":   ExtensionTypeMACAddress,

	// Time Stamp
	"art":   ExtensionTypeTimestamp,
	"end":   ExtensionTypeTimestamp,
	"rt":    ExtensionTypeTimestamp,
	"start": ExtensionTypeTimestamp,
}

// cefTimestampLayouts are the Time reference layouts ValidateExtensionTypes
// accepts for a CEF Time Stamp value, corresponding to the spec's documented
// "MMM dd yyyy HH:mm:ss" format — both zero-padded ("Jan 02") and
// space/non-padded ("Jan 2") day forms are accepted, since the spec's examples
// are inconsistent about which they use. This is not the full list of date
// formats CEF implementations may accept elsewhere in the spec, just the one
// documented for this header-adjacent format; milliseconds-since-epoch
// (all-digit values) is handled separately in isValidCefTimestamp.
var cefTimestampLayouts = []string{
	"Jan 02 2006 15:04:05",
	"Jan 2 2006 15:04:05",
}

// isValidCefTimestamp reports whether value is a valid CEF Time Stamp: either
// an integer (milliseconds since the Unix epoch) or one of cefTimestampLayouts.
func isValidCefTimestamp(value string) bool {
	if _, err := strconv.ParseInt(value, 10, 64); err == nil {
		return true
	}
	for _, layout := range cefTimestampLayouts {
		if _, err := time.Parse(layout, value); err == nil {
			return true
		}
	}
	return false
}

// isValidExtensionTypeValue reports whether value is well-formed for CEF data
// type t.
func isValidExtensionTypeValue(t ExtensionFieldType, value string) bool {
	switch t {
	case ExtensionTypeInteger, ExtensionTypeLong:
		_, err := strconv.ParseInt(value, 10, 64)
		return err == nil
	case ExtensionTypeFloat:
		_, err := strconv.ParseFloat(value, 64)
		return err == nil
	case ExtensionTypeIPAddress:
		return net.ParseIP(value) != nil
	case ExtensionTypeMACAddress:
		_, err := net.ParseMAC(value)
		return err == nil
	case ExtensionTypeTimestamp:
		return isValidCefTimestamp(value)
	default:
		return true
	}
}

// ValidateExtensionTypes checks every extension key present in
// event.Extensions that ValidateExtensionTypes recognizes (see
// ExtensionFieldTypes) against that key's documented CEF data type — for
// example, that dst/src parse as an IP address, dpt/spt as an integer,
// dmac/smac as a MAC address, and rt/start/end as a CEF timestamp. Keys not
// in ExtensionFieldTypes — including every String-typed key — are not
// checked here; see ValidateExtensionLengths for those.
//
// This is opt-in, exactly like the length validators: nothing in the package
// calls it automatically.
//
// Returns nil if every checked field is well-formed for its type, or an
// error describing every field that isn't.
func (event *CefEvent) ValidateExtensionTypes() error {

	var violations []string

	keys := make([]string, 0, len(event.Extensions))
	for k := range event.Extensions {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		t, ok := ExtensionFieldTypes[k]
		if !ok {
			continue
		}
		if !isValidExtensionTypeValue(t, event.Extensions[k]) {
			violations = append(violations, fmt.Sprintf("%s: %q is not a valid %s", k, event.Extensions[k], t))
		}
	}

	if len(violations) > 0 {
		return fmt.Errorf("CEF extension field(s) have invalid values for their documented type: %s", strings.Join(violations, "; "))
	}

	return nil
}
