// Package cefevent implements ArcSight's Common Event Format (CEF): building a CEF
// message, generating one as a string or as JSON, parsing one (or many, from a
// multi-line log via ReadAll), and logging one to stdout/stderr or to syslog.
//
// # Loose by design
//
// The CEF spec defines maximum lengths for header and extension fields, and data
// types (integer, IP address, MAC address, timestamp, ...) for many extension keys.
// None of that is enforced by [CefEvent.String], [CefEvent.Build], or
// [CefEvent.Read] — this package stays a loose implementation by default, and
// nothing in that core generate/parse path calls any of the validators below on its
// own. Enforcing spec compliance is always something a caller opts into explicitly:
//
//   - [CefEvent.ValidateHeaderLengths], [CefEvent.ValidateExtensionLengths], and
//     [CefEvent.ValidateFieldLengths] (both combined) reject an event whose header
//     or extension fields exceed the spec's documented maximum length.
//   - [CefEvent.TruncateToLimits] is the best-effort counterpart: instead of
//     rejecting an over-length event, it shortens the offending fields in place.
//   - [CefEvent.ValidateExtensionTypes] checks well-known non-string extension
//     keys (dst/src as an IP address, dpt/spt as an integer, dmac/smac as a MAC
//     address, and so on) against their documented CEF data type.
//
// See [HeaderFieldLimits], [ExtensionFieldLimits], and [ExtensionFieldTypes] for
// exactly which fields and keys each of those covers — in particular,
// ExtensionFieldTypes is a curated subset of the spec's non-string-typed extension
// keys, individually verified against the spec rather than exhaustively
// transcribed.
//
// # Basic usage
//
//	event := cefevent.CefEvent{
//		Version:            0,
//		DeviceVendor:       "Cool Vendor",
//		DeviceProduct:      "Cool Product",
//		DeviceVersion:      "1.0",
//		DeviceEventClassId: "COOL_THING",
//		Name:               "Something cool happened.",
//		Severity:           "Unknown",
//		Extensions:         map[string]string{"src": "127.0.0.1"},
//	}
//
//	message, err := event.String()
//	// message == "CEF:0|Cool Vendor|Cool Product|1.0|COOL_THING|Something cool happened.|Unknown|src=127.0.0.1"
//
// See the module's examples/main.go for a complete runnable program.
package cefevent
