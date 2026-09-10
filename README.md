# Common Event Format in Go
Go Package for ArcSight's Common Event Format

![Build Workflow](https://github.com/pcktdmp/cef/workflows/Build/badge.svg)
![Test Workflow](https://github.com/pcktdmp/cef/workflows/Test/badge.svg)

# Motivation

Learning Go, help people who need to process CEF events in Golang.

## TL;DR

`cefevent` is a [loose implementation](#known-limitations) of the Common Event Format: by
default it doesn't enforce the [documented](https://www.microfocus.com/documentation/arcsight/arcsight-smartconnectors-8.3/pdfdoc/cef-implementation-standard/cef-implementation-standard.pdf)
field-length or field-type rules, and it's up to whoever generates or consumes events to
decide whether that matters for them. If it does, opt-in helpers exist for that — see
[Field limits and type-checking](#field-limits-and-type-checking) below.

### Install the package

```bash
$ go get github.com/pcktdmp/cef/cefevent
```

### examples/main.go

```go
package main

import (
	"fmt"
	"github.com/pcktdmp/cef/cefevent"
)

func main() {

	// create CEF event
	f := make(map[string]string)
	f["src"] = "127.0.0.1"
	f["requestClientApplication"] = "Go-http-client/1.1"

	event := cefevent.CefEvent{
		Version:            0,
		DeviceVendor:       "Cool Vendor",
		DeviceProduct:      "Cool Product",
		DeviceVersion:      "1.0",
		DeviceEventClassId: "FLAKY_EVENT",
		Name:               "Something flaky happened.",
		Severity:           "3",
		Extensions:         f,
	}

	eventString, err := event.String()
	if err != nil {
		fmt.Println("Need to handle this.")
	}
	fmt.Println(eventString)

	// send a CEF event as log message to stdout
	event.Log()

	// or if you want to do error handling when
	// sending the log
	err = event.Log()

	if err != nil {
		fmt.Println("Need to handle this.")
	}

	// if you want read a CEF event from a line
	eventLine := "CEF:0|Cool Vendor|Cool Product|1.0|COOL_THING|Something cool happened.|Unknown|src=127.0.0.1"
	newEvent := cefevent.CefEvent{}
	_, err = newEvent.Read(eventLine)
	if err != nil {
		fmt.Println("Need to handle this.")
	}
	eventString, err = newEvent.String()
	if err != nil {
		fmt.Println("Need to handle this.")
	}
	fmt.Println(eventString)

}

```
### Example output

```bash
$ go run ./examples
CEF:0|Cool Vendor|Cool Product|1.0|FLAKY_EVENT|Something flaky happened.|3|requestClientApplication=Go-http-client/1.1 src=127.0.0.1
2020/03/12 21:28:19 CEF:0|Cool Vendor|Cool Product|1.0|FLAKY_EVENT|Something flaky happened.|3|requestClientApplication=Go-http-client/1.1 src=127.0.0.1
2020/03/12 21:28:19 CEF:0|Cool Vendor|Cool Product|1.0|FLAKY_EVENT|Something flaky happened.|3|requestClientApplication=Go-http-client/1.1 src=127.0.0.1
CEF:0|Cool Vendor|Cool Product|1.0|COOL_THING|Something cool happened.|Unknown|src=127.0.0.1
```

## Field limits and type-checking

CEF's spec defines maximum lengths for header and extension fields, and data types
(integer, IP address, MAC address, timestamp, ...) for many extension keys. None of that
is enforced by `String()`/`Build()`/`Read()` — this package stays a loose implementation
by default, on purpose. If you want it enforced, call one of these explicitly:

```go
event := cefevent.CefEvent{ /* ... */ }

// Reject an event whose header or extension fields exceed the spec's documented
// maximum length.
if err := event.ValidateFieldLengths(); err != nil {
	// event.DeviceVendor is too long, or an extension value like event.Extensions["act"] is
}

// ...or don't reject it - shorten offending fields in place instead, and see which ones
// changed.
truncated := event.TruncateToLimits()

// Check well-known non-string extension keys against their documented CEF data type
// (dst/src as an IP address, dpt/spt as an integer, dmac/smac as a MAC address, ...).
if err := event.ValidateExtensionTypes(); err != nil {
	// e.g. event.Extensions["dst"] doesn't parse as an IP address
}
```

`ValidateHeaderLengths`/`ValidateExtensionLengths` are the header-only/extension-only
halves of `ValidateFieldLengths`, if you only need one. See the doc comments on
`HeaderFieldLimits`, `ExtensionFieldLimits`, and `ExtensionFieldTypes` in the `cefevent`
package for exactly which fields and keys are covered — the type map in particular is a
curated subset (see [`CEF-SPEC.md`](CEF-SPEC.md)), not every non-string key in the spec.

## Converting to JSON

`ToJSON()` validates and escapes the event the same way `String()`/`Build()` do, then
marshals it to a JSON string:

```go
jsonStr, err := event.ToJSON()
```

## Parsing more than one line

`Read` parses a single CEF message. To parse a log file or a multi-line string, use
`ReadAll`, which strips a leading syslog prefix if present and reports per-line parse
failures without aborting the rest of the read:

```go
events, lineErrs, err := cefevent.ReadAll(reader)
// err is only ever an I/O failure reading from `reader`; a malformed line shows up as an
// entry in lineErrs (with its line number) instead of stopping the read.
```

## Logging to syslog

`Log()` writes to stdout/stderr via the standard `log` package. `LogToSyslog` writes to
the local syslog daemon instead, via the standard `log/syslog` package — which CEF is
overwhelmingly shipped over in practice:

```go
err := event.LogToSyslog(syslog.LOG_INFO|syslog.LOG_USER, "myapp")
```

Not available on Windows or Plan 9: `log/syslog` itself has no implementation there.

## Known limitations

* None of the above is applied automatically — enforcing field limits or types is
  always something you opt into by calling the relevant function yourself.
* `ExtensionFieldTypes` covers a curated subset of the spec's non-string-typed
  extension keys, individually verified against the spec rather than exhaustively
  transcribed — see its doc comment for exactly which keys.
* Parsing an extension value that happens to contain literal `" key="` text not meant
  as a new key can be misread as starting a new key/value pair. This is an ambiguity in
  the CEF extension format itself (no stricter grammar is defined by the spec), not
  something specific to this implementation.
