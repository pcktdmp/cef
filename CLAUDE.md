# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`cefevent` is a Go package implementing ArcSight's Common Event Format (CEF) — a
loose implementation by default; `String()`/`Build()`/`Read()` never enforce the CEF
spec's field-length or field-type rules. Opt-in enforcement of those exists
(`ValidateHeaderLengths`/`ValidateExtensionLengths`/`ValidateFieldLengths`/
`TruncateToLimits` in `cefevent/limits.go`, `ValidateExtensionTypes` in
`cefevent/types.go` — see "Field-length limits" and "Other opt-in helpers" below),
but it is always the caller's choice to invoke it — nothing in the core Generate/Parse
path calls any of it automatically. The module root (`examples.go`) is a runnable demo
package, not part of the library API.

## Field-length limits

`cefevent/limits.go` defines two lookup tables — `HeaderFieldLimits` (5 of the 7 CEF
header fields; `Version` and `Severity` have no documented length) and
`ExtensionFieldLimits` (~90 well-known Extension Dictionary keys) — plus three opt-in
validators (`ValidateHeaderLengths`, `ValidateExtensionLengths`, `ValidateFieldLengths`)
that check a `CefEvent`'s fields against them. These are **not** part of the
`CefEventer` interface and are never called by `String`/`Build`/`Read`/`Validate` —
adding them there would make the "loose by default" behavior above a lie.

**`CEF-SPEC.md` (repo root) is the cached source of truth for both tables.** It documents
which spec document and URL the numbers came from (the CEF spec has multiple published
versions with materially different content — see the note below) and is meant to be
checked before changing either table, without needing to re-fetch and re-parse the PDF
each time. If you touch `HeaderFieldLimits` or `ExtensionFieldLimits`, update
`CEF-SPEC.md` in the same change, and vice versa — they must stay in sync.

Known trap: an older ArcSight whitepaper, *Common Event Format v25*, has an Extension
Dictionary identical to the one cached here, but its header field section defines
**no** length limits at all — only the newer *Implementing ArcSight Common Event
Format (CEF)* "Version 26" document (ArcSight SmartConnectors 8.3 docs) does. If you
consult the CEF spec directly rather than `CEF-SPEC.md`, make sure you're looking at
the Version 26 document, not v25, or you'll conclude (incorrectly) that header fields
have no length limits.

## Other opt-in helpers

Same "never called automatically" rule as the length validators above applies to all
of these:

- **`TruncateToLimits`** (`cefevent/limits.go`) — best-effort counterpart to
  `ValidateHeaderLengths`/`ValidateExtensionLengths`: truncates (by rune, never
  splitting a multi-byte character) any field over its documented limit instead of
  rejecting the event, and returns which fields it touched.
- **`ExtensionFieldTypes`/`ValidateExtensionTypes`** (`cefevent/types.go`) — checks
  well-known non-String Extension Dictionary keys (`dst`/`src` as IP addresses,
  `dpt`/`spt` as integers, `dmac`/`smac` as MAC addresses, `rt`/`start`/`end` as
  timestamps, etc.) against their documented CEF data type. This is a curated subset
  individually confirmed against the spec, not an exhaustive transcription — see the
  doc comment on `ExtensionFieldTypes` before assuming a key is or isn't covered.
- **`ReadAll(io.Reader)`** (`cefevent/readall.go`) — parses a multi-line CEF log via
  repeated `Read` calls. Bad lines don't abort the read; they're collected as
  `*LineError` (unwrappable via `errors.Is`/`As`) in a separate return value instead.
  Also strips a leading syslog prefix before the `CEF:` marker, since `Read` itself
  requires the line to start with it.
- **`LogToSyslog`** (`cefevent/syslog.go`) — syslog-native counterpart to `Log()`
  (which only ever writes to stdout/stderr via the standard `log` package). Guarded
  with `//go:build !windows && !plan9`, mirroring `log/syslog`'s own platform support —
  don't remove that constraint without also handling those platforms.

## Commands

```bash
# Run the full test suite (this is what CI runs, via test.yml)
cd cefevent && go test -v

# Run a single test
cd cefevent && go test -run TestCefEventParsed -v

# Build (this is what CI runs, via build.yml)
go build -v cefevent/cefevent.go

# Run the example program
go run examples.go
```

There is no lint config or Makefile in the repo; `golangci-lint`/`gopls` are available
in the devcontainer but not wired into CI.

## Architecture

Everything lives in a single package, `cefevent` (`cefevent/cefevent.go`), built around
one struct and one interface:

- **`CefEvent`** struct — the data model for a CEF event: `Version`, `DeviceVendor`,
  `DeviceProduct`, `DeviceVersion`, `DeviceEventClassId`, `Name`, `Severity`, and an
  `Extensions map[string]string`. Struct tags carry `json`/`yaml`/`toml`/`xml` names
  plus `header`/`comment` metadata for other tooling.
- **`CefEventer`** interface — `Validate`, `String`, `Build`, `Read`, `Log`, and the
  unexported `escapeEventData`. `*CefEvent` implements it; methods on `CefEvent` call
  through the interface (e.g. `CefEventer.Validate(event)`) rather than calling
  `event.Validate()` directly in a couple of spots — that's intentional/existing style,
  not a mistake to "fix".

Data flow, both directions go through the same escaping step:

- **Generate**: `String()` validates mandatory fields → `escapeEventData()` escapes all
  fields in place → extensions are sorted by key (`sort.Strings`) and joined as
  `key=value ` pairs → the 7 header fields plus extensions are joined with `|` into
  `CEF:Version|DeviceVendor|...|Extensions`. `Build()` does the same but returns the
  (mutated, escaped) `CefEvent` instead of a string.
- **Parse**: `Read(line string)` requires a `CEF:` prefix, splits the remainder on
  unescaped `|` only (`splitUnescapedPipes`, capped at 8 parts so index 7 — the
  extensions blob — captures everything after the 7th unescaped pipe as one opaque
  remainder, even if it contains further literal `|`, which is legal there), locates
  extension key/value boundaries by their `" key="` pattern rather than blindly on
  every space (`splitExtensions`, honoring the spec's multiple-spaces/trailing-space
  rules), then unescapes every parsed field (`cefUnescapeField`/`cefUnescapeExtension`)
  before validating and returning.
- Escaping is asymmetric by design in the CEF spec: header fields escape `\`, `|`, `\n`
  (`cefEscapeField`); extension keys/values escape `\`, `\n`, `=` (`cefEscapeExtension`).
  `Read` populates `CefEvent` fields with raw, unescaped data — the same invariant a
  freshly constructed `CefEvent{}` has — so a subsequent `String()`/`Build()` escapes it
  symmetrically; it used to re-escape already-escaped wire text instead (silently
  double-escaping on any round trip), which is what made `Read` not a correct inverse
  of `String`/`Build` before. `splitExtensions`'s key-boundary heuristic can still
  misparse a value that happens to contain literal `" key="` text not meant as a new
  key — that ambiguity is inherent to the CEF extension format itself (no stricter
  grammar than that is defined by the spec), not something specific to this
  implementation. `Read` used to panic with index-out-of-range on short or
  extension-less input (`len(eventSlashed) >= 7` let it index `[7]`); that's fixed — it
  now requires `len(eventSlashed) >= 7` before touching indices `1`-`6` and `>= 8`
  before touching `[7]`, returning an error instead of panicking on malformed input.

`String()`/`Build()`/`ToJSON()` used to mutate the receiver in place via
`escapeEventData()` — calling `event.String()` didn't just read `event`, it overwrote
`event.DeviceVendor`/etc. with their escaped form as a side effect, so calling
`String()` (or `Build`/`ToJSON`) a second time on the same event re-escaped
already-escaped data. That's fixed: all three now call `escapeEventData()` on a local
`escaped := *event` copy and build their output from that, leaving the receiver
untouched — see `TestStringBuildToJSONDoNotMutateReceiver` and
`Test{String,Build}CalledTwiceDoesNotDoubleEscape` in `cefevent_test.go`.

`Validate()` uses reflection (`reflect.ValueOf(event).Elem().FieldByName(...)`) over a
hardcoded list of mandatory field names rather than checking struct fields directly —
keep the field-name list in sync with the struct if you rename a mandatory field.

`Log()` always calls `log.SetOutput` (stdout on success, stderr on failure) before
`log.Println` — this mutates the shared standard logger's output target as a
side effect, which matters if other code in the same process also uses `log`.

`ToJSON()` validates via `CefEventer.Validate(event)` and escapes via `escapeEventData()`
before marshaling, the same as `String()`/`Build()`/`Read()` — it used to validate via
`Validate()` directly and skip escaping, but that inconsistency is fixed now.
