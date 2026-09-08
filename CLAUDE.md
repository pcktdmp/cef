# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`cefevent` is a Go package implementing ArcSight's Common Event Format (CEF) — a
loose implementation by default; `String()`/`Build()`/`Read()` never enforce the CEF
spec's field-length limits. Opt-in enforcement of those limits exists
(`ValidateHeaderLengths`/`ValidateExtensionLengths`/`ValidateFieldLengths` in
`cefevent/limits.go`, see "Field-length limits" below), but it is always the caller's
choice to invoke it — nothing in the core Generate/Parse path calls it automatically.
The module root (`examples.go`) is a runnable demo package, not part of the library API.

## Field-length limits

`cefevent/limits.go` defines two lookup tables — `HeaderFieldLimits` (5 of the 7 CEF
header fields; `Version` and `Severity` have no documented length) and
`ExtensionFieldLimits` (~90 well-known Extension Dictionary keys) — plus three opt-in
validators (`ValidateHeaderLengths`, `ValidateExtensionLengths`, `ValidateFieldLengths`)
that check a `CefEvent`'s fields against them. These are **not** part of the
`CefEventer` interface and are never called by `String`/`Build`/`Read`/`Validate` —
adding them there would make the "loose by default" behavior above a lie.

**`cefevent/CEF-SPEC.md` is the cached source of truth for both tables.** It documents
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
- **Parse**: `Read(line string)` requires a `CEF:` prefix, splits the remainder on `|`
  positionally (index 0 = Version ... index 6 = Severity, index 7 = extensions blob),
  splits the extensions blob on spaces then each piece on the first `=`, then also runs
  the result through `escapeEventData()` before validating and returning.
- Escaping is asymmetric by design in the CEF spec: header fields escape `\`, `|`, `\n`
  (`cefEscapeField`); extension keys/values escape `\`, `\n`, `=` (`cefEscapeExtension`).
  `Read` does not *unescape* — the round-trip tests only exercise strings without
  embedded delimiters. Known correctness gaps (escaped `|` inside a field, extension
  values containing spaces) are real but pre-existing; don't assume they're accidental
  unless asked to fix them. `Read` used to panic with index-out-of-range on short or
  extension-less input (`len(eventSlashed) >= 7` let it index `[7]`); that's fixed — it
  now requires `len(eventSlashed) >= 7` before touching indices `1`-`6` and `>= 8`
  before touching `[7]`, returning an error instead of panicking on malformed input.

`Validate()` uses reflection (`reflect.ValueOf(event).Elem().FieldByName(...)`) over a
hardcoded list of mandatory field names rather than checking struct fields directly —
keep the field-name list in sync with the struct if you rename a mandatory field.

`Log()` always calls `log.SetOutput` (stdout on success, stderr on failure) before
`log.Println` — this mutates the shared standard logger's output target as a
side effect, which matters if other code in the same process also uses `log`.

`ToJSON()` validates via `CefEventer.Validate(event)` and escapes via `escapeEventData()`
before marshaling, the same as `String()`/`Build()`/`Read()` — it used to validate via
`Validate()` directly and skip escaping, but that inconsistency is fixed now.
