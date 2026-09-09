package cefevent

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

// HeaderFieldLimits maps the CEF header fields that carry a documented
// maximum length to that length, in characters, per the "Header Field
// Definitions" table in "Implementing ArcSight Common Event Format (CEF)"
// (Version 26, ArcSight SmartConnectors 8.3 CEF Implementation Standard).
// Map keys are the corresponding CefEvent struct field names.
//
// CEF Version and Severity are intentionally excluded: Version is a
// numeric format identifier (currently 0 or 1), not a length-limited
// string, and Severity is a constrained enumeration/integer
// (Unknown/Low/Medium/High/Very-High, or 0-10), not a free-form string
// with a documented character limit. Neither has a length in the spec.
//
// An earlier ArcSight CEF whitepaper ("Common Event Format v25") omitted
// header field lengths entirely; this newer Implementation Standard
// document defines them explicitly, which is what this map reflects.
//
// This map exists to support the opt-in enforcement in
// ValidateHeaderLengths; see ExtensionFieldLimits for the equivalent for
// extension fields.
var HeaderFieldLimits = map[string]int{
	"DeviceVendor":       63,
	"DeviceProduct":      63,
	"DeviceVersion":      31,
	"DeviceEventClassId": 1023,
	"Name":               512,
}

// ExtensionFieldLimits maps CEF Extension Dictionary key names, as sent on
// the wire (for example "act", "requestClientApplication"), to their
// maximum length in characters as documented in the ArcSight Common Event
// Format specification's Extension Dictionary chapter.
//
// Only keys whose documented Data Type is String (or a string-derived
// type such as a Label) carry a length limit in the specification;
// numeric, IP/MAC address, timestamp and floating-point typed keys have
// no documented length limit and are therefore absent from this map.
// Custom extension keys — anything not part of the predefined Extension
// Dictionary — are outside the spec's scope and have no defined limit
// either, so they are absent too.
//
// This map exists to support the opt-in enforcement in
// ValidateExtensionLengths. cefevent's core methods (String, Build, Read)
// never consult it — the package remains a loose CEF implementation by
// default, and enforcing spec field-length limits is the caller's
// responsibility unless they explicitly opt in.
var ExtensionFieldLimits = map[string]int{
	"act":                                 63,
	"agentDnsDomain":                      255,
	"agentNtDomain":                       255,
	"agentTranslatedZoneExternalID":       200,
	"agentTranslatedZoneURI":              2048,
	"agentZoneExternalID":                 200,
	"agentZoneURI":                        2048,
	"ahost":                               1023,
	"aid":                                 40,
	"app":                                 31,
	"at":                                  63,
	"atz":                                 255,
	"av":                                  31,
	"c6a1Label":                           1023,
	"c6a3Label":                           1023,
	"c6a4Label":                           1023,
	"cat":                                 1023,
	"cfp1Label":                           1023,
	"cfp2Label":                           1023,
	"cfp3Label":                           1023,
	"cfp4Label":                           1023,
	"cn1Label":                            1023,
	"cn2Label":                            1023,
	"cn3Label":                            1023,
	"cs1":                                 4000,
	"cs1Label":                            1023,
	"cs2":                                 4000,
	"cs2Label":                            1023,
	"cs3":                                 4000,
	"cs3Label":                            1023,
	"cs4":                                 4000,
	"cs4Label":                            1023,
	"cs5":                                 4000,
	"cs5Label":                            1023,
	"cs6":                                 4000,
	"cs6Label":                            1023,
	"customerExternalID":                  200,
	"customerURI":                         2048,
	"destinationDnsDomain":                255,
	"destinationServiceName":              1023,
	"destinationTranslatedZoneExternalID": 200,
	"destinationTranslatedZoneURI":        2048,
	"destinationZoneExternalID":           200,
	"destinationZoneURI":                  2048,
	"deviceCustomDate1Label":              1023,
	"deviceCustomDate2Label":              1023,
	"deviceDnsDomain":                     255,
	"deviceExternalId":                    255,
	"deviceFacility":                      1023,
	"deviceInboundInterface":              128,
	"deviceNtDomain":                      255,
	"deviceOutboundInterface":             128,
	"devicePayloadId":                     128,
	"deviceProcessName":                   1023,
	"deviceTranslatedZoneExternalID":      200,
	"deviceTranslatedZoneURI":             2048,
	"deviceZoneExternalID":                200,
	"deviceZoneURI":                       2048,
	"dhost":                               1023,
	"dntdom":                              255,
	"dpriv":                               1023,
	"dproc":                               1023,
	"dtz":                                 255,
	"duid":                                1023,
	"duser":                               1023,
	"dvchost":                             100,
	"externalId":                          40,
	"fileHash":                            255,
	"fileId":                              1023,
	"filePath":                            1023,
	"filePermission":                      1023,
	"fileType":                            1023,
	"flexDate1Label":                      128,
	"flexString1":                         1023,
	"flexString1Label":                    128,
	"flexString2":                         1023,
	"flexString2Label":                    128,
	"fname":                               1023,
	"msg":                                 1023,
	"oldFileHash":                         255,
	"oldFileId":                           1023,
	"oldFileName":                         1023,
	"oldFilePath":                         1023,
	"oldFilePermission":                   1023,
	"oldFileType":                         1023,
	"outcome":                             63,
	"proto":                               31,
	"rawEvent":                            4000,
	"reason":                              1023,
	"request":                             1023,
	"requestClientApplication":            1023,
	"requestContext":                      2048,
	"requestCookies":                      1023,
	"requestMethod":                       1023,
	"shost":                               1023,
	"sntdom":                              255,
	"sourceDnsDomain":                     255,
	"sourceServiceName":                   1023,
	"sourceTranslatedZoneExternalID":      200,
	"sourceTranslatedZoneURI":             2048,
	"sourceZoneExternalID":                200,
	"sourceZoneURI":                       2048,
	"spriv":                               1023,
	"sproc":                               1023,
	"suid":                                1023,
	"suser":                               1023,
}

// ValidateExtensionLengths checks every extension key present in
// event.Extensions that is part of the ArcSight CEF Extension Dictionary
// against its documented maximum length (see ExtensionFieldLimits).
// Extension keys not found in ExtensionFieldLimits are either custom keys
// or non-string typed dictionary keys, and are not checked, since the CEF
// specification does not define a length limit for them.
//
// Length is measured in characters (Unicode code points), not bytes,
// matching how the spec describes field lengths.
//
// This is opt-in: String, Build, and Read never call it, so cefevent
// remains a loose CEF implementation by default. Call
// ValidateExtensionLengths explicitly if you want the spec's extension
// length limits enforced, e.g. before generating a message with String or
// Build, or after parsing one with Read.
//
// Returns nil if every checked field is within its documented limit, or
// an error describing every field that exceeds its limit.
func (event *CefEvent) ValidateExtensionLengths() error {

	var violations []string

	keys := make([]string, 0, len(event.Extensions))
	for k := range event.Extensions {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		limit, ok := ExtensionFieldLimits[k]
		if !ok {
			continue
		}
		if length := utf8.RuneCountInString(event.Extensions[k]); length > limit {
			violations = append(violations, fmt.Sprintf("%s: length %d exceeds maximum %d", k, length, limit))
		}
	}

	if len(violations) > 0 {
		return fmt.Errorf("CEF extension field(s) exceed the CEF spec's maximum length: %s", strings.Join(violations, "; "))
	}

	return nil
}

// ValidateHeaderLengths checks the CEF header fields that carry a
// documented maximum length (see HeaderFieldLimits) against that limit:
// DeviceVendor, DeviceProduct, DeviceVersion, DeviceEventClassId, and
// Name. Version and Severity are not checked; the CEF specification does
// not define a character limit for either.
//
// Length is measured in characters (Unicode code points), not bytes,
// matching how the spec describes field lengths.
//
// This is opt-in, exactly like ValidateExtensionLengths: String, Build,
// and Read never call it, so cefevent remains a loose CEF implementation
// by default. Call ValidateHeaderLengths explicitly if you want the
// spec's header length limits enforced.
//
// Returns nil if every field is within its documented limit, or an error
// describing every field that exceeds its limit.
func (event *CefEvent) ValidateHeaderLengths() error {

	fields := []struct {
		name  string
		value string
	}{
		{"DeviceVendor", event.DeviceVendor},
		{"DeviceProduct", event.DeviceProduct},
		{"DeviceVersion", event.DeviceVersion},
		{"DeviceEventClassId", event.DeviceEventClassId},
		{"Name", event.Name},
	}

	var violations []string

	for _, field := range fields {
		limit := HeaderFieldLimits[field.name]
		if length := utf8.RuneCountInString(field.value); length > limit {
			violations = append(violations, fmt.Sprintf("%s: length %d exceeds maximum %d", field.name, length, limit))
		}
	}

	if len(violations) > 0 {
		return fmt.Errorf("CEF header field(s) exceed the CEF spec's maximum length: %s", strings.Join(violations, "; "))
	}

	return nil
}

// ValidateFieldLengths runs both ValidateHeaderLengths and
// ValidateExtensionLengths, combining their results into a single error.
// It exists for callers who want to opt into every documented CEF
// field-length limit — header and extension — in one call, rather than
// invoking the two checks separately.
//
// Like both underlying checks, this is opt-in: String, Build, and Read
// never call it.
func (event *CefEvent) ValidateFieldLengths() error {

	var errs []string

	if err := event.ValidateHeaderLengths(); err != nil {
		errs = append(errs, err.Error())
	}
	if err := event.ValidateExtensionLengths(); err != nil {
		errs = append(errs, err.Error())
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}

// TruncateToLimits truncates any header or extension field that exceeds its
// documented CEF spec maximum length (see HeaderFieldLimits and
// ExtensionFieldLimits) down to that limit, mutating event in place. It's the
// best-effort counterpart to ValidateHeaderLengths/ValidateExtensionLengths:
// those reject an over-length event outright, this repairs it so a producer
// can still emit *a* valid-length event rather than dropping it.
//
// Like the validators, this operates on raw (not yet escaped) field values —
// call it before String()/Build(), the same point you'd call
// ValidateHeaderLengths/ValidateExtensionLengths. Truncating post-escaping
// would count the extra bytes escaping adds toward the limit, which isn't
// what the spec's length limits describe.
//
// Truncation is by Unicode character (rune), not byte, so it never splits a
// multi-byte character. Header field names in the result use the CefEvent
// struct field name (e.g. "DeviceVendor"); extension field names use the CEF
// key (e.g. "act"). Version and Severity are never touched — like
// ValidateHeaderLengths, this only knows about the 5 header fields with a
// documented length, and unrecognized extension keys are left alone.
//
// Returns the names of every field it actually shortened, in a stable order
// (header fields first in struct order, then extension keys sorted), or nil
// if nothing needed truncating.
func (event *CefEvent) TruncateToLimits() []string {

	var truncated []string

	headerFields := []struct {
		name  string
		value *string
	}{
		{"DeviceVendor", &event.DeviceVendor},
		{"DeviceProduct", &event.DeviceProduct},
		{"DeviceVersion", &event.DeviceVersion},
		{"DeviceEventClassId", &event.DeviceEventClassId},
		{"Name", &event.Name},
	}

	for _, field := range headerFields {
		if truncateToRuneLimit(field.value, HeaderFieldLimits[field.name]) {
			truncated = append(truncated, field.name)
		}
	}

	keys := make([]string, 0, len(event.Extensions))
	for k := range event.Extensions {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		limit, ok := ExtensionFieldLimits[k]
		if !ok {
			continue
		}
		value := event.Extensions[k]
		if truncateToRuneLimit(&value, limit) {
			event.Extensions[k] = value
			truncated = append(truncated, k)
		}
	}

	return truncated
}

// truncateToRuneLimit shortens *s to at most limit runes (Unicode code
// points), in place, and reports whether it changed anything.
func truncateToRuneLimit(s *string, limit int) bool {
	if utf8.RuneCountInString(*s) <= limit {
		return false
	}
	*s = string([]rune(*s)[:limit])
	return true
}
