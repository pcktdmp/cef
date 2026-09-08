# CEF field-length reference

This file is a cached, condensed reference for the CEF field-length limits
implemented in `limits.go` (`HeaderFieldLimits`, `ExtensionFieldLimits`).
It exists so that changes to those tables can be checked against the spec
without re-fetching and re-parsing the source PDF every time.

**Source:** *Implementing ArcSight Common Event Format (CEF)* — Version 26
of the standard, published as part of the ArcSight SmartConnectors 8.3
documentation set.
<https://www.microfocus.com/documentation/arcsight/arcsight-smartconnectors-8.3/pdfdoc/cef-implementation-standard/cef-implementation-standard.pdf>
(retrieved 2026-09-08).

An older ArcSight whitepaper, *Common Event Format v25*
(<https://www.microfocus.com/documentation/arcsight/arcsight-smartconnectors/pdfdoc/common-event-format-v25/common-event-format-v25.pdf>),
covers the same Extension Dictionary with identical lengths, but its
"Header Field Definitions" section has **no length limits at all** for the
header fields. The Version 26 document above is newer and is the one that
actually defines header field lengths — it is the authoritative source for
`HeaderFieldLimits`. If these two sources ever disagree on an extension
length, prefer the newer document.

If Micro Focus/OpenText republish this document at a different URL or
version, re-fetch it, diff the two tables below against the new content,
and update `limits.go` (plus this file) together in the same change.

## Header fields

From the "Header Field Definitions" table. Field names below are the
`CefEvent` struct field names; the table's own column uses lowerCamelCase
(e.g. `deviceEventClassId`).

| CefEvent field       | Data Type              | Max length | Notes |
|-----------------------|-------------------------|-----------:|-------|
| `Version`              | Numeric                 | *(none)*   | Format identifier, currently `0` or `1` — not a length-limited string. |
| `DeviceVendor`          | String                  | 63         | |
| `DeviceProduct`         | String                  | 63         | |
| `DeviceVersion`         | String                  | 31         | |
| `DeviceEventClassId`    | String                  | 1023       | Also known as Signature ID. |
| `Name`                  | String                  | 512        | |
| `Severity`              | AgentSeverityEnumeration| *(none)*   | Constrained values: `Unknown`/`Low`/`Medium`/`High`/`Very-High`, or integer `0`-`10` — not a length-limited string. |

## Extension Dictionary (string-typed keys only)

Only extension keys with a `String` (or string-derived, e.g. Label) data
type carry a documented length. Numeric, IP/MAC address, timestamp and
floating-point typed keys (e.g. `dst`, `spt`, `cnt`, `rt`, `cfp1`, `cn1`)
have no documented length limit and are omitted here — as are custom
extension keys, which aren't part of the predefined dictionary at all.

| CEF key | Max length | | CEF key | Max length |
|---|---:|---|---|---:|
| `act` | 63 | | `duid` | 1023 |
| `agentDnsDomain` | 255 | | `duser` | 1023 |
| `agentNtDomain` | 255 | | `dvchost` | 100 |
| `agentTranslatedZoneExternalID` | 200 | | `externalId` | 40 |
| `agentTranslatedZoneURI` | 2048 | | `fileHash` | 255 |
| `agentZoneExternalID` | 200 | | `fileId` | 1023 |
| `agentZoneURI` | 2048 | | `filePath` | 1023 |
| `ahost` | 1023 | | `filePermission` | 1023 |
| `aid` | 40 | | `fileType` | 1023 |
| `app` | 31 | | `flexDate1Label` | 128 |
| `at` | 63 | | `flexString1` | 1023 |
| `atz` | 255 | | `flexString1Label` | 128 |
| `av` | 31 | | `flexString2` | 1023 |
| `c6a1Label` | 1023 | | `flexString2Label` | 128 |
| `c6a3Label` | 1023 | | `fname` | 1023 |
| `c6a4Label` | 1023 | | `msg` | 1023 |
| `cat` | 1023 | | `oldFileHash` | 255 |
| `cfp1Label` | 1023 | | `oldFileId` | 1023 |
| `cfp2Label` | 1023 | | `oldFileName` | 1023 |
| `cfp3Label` | 1023 | | `oldFilePath` | 1023 |
| `cfp4Label` | 1023 | | `oldFilePermission` | 1023 |
| `cn1Label` | 1023 | | `oldFileType` | 1023 |
| `cn2Label` | 1023 | | `outcome` | 63 |
| `cn3Label` | 1023 | | `proto` | 31 |
| `cs1` | 4000 | | `rawEvent` | 4000 |
| `cs1Label` | 1023 | | `reason` | 1023 |
| `cs2` | 4000 | | `request` | 1023 |
| `cs2Label` | 1023 | | `requestClientApplication` | 1023 |
| `cs3` | 4000 | | `requestContext` | 2048 |
| `cs3Label` | 1023 | | `requestCookies` | 1023 |
| `cs4` | 4000 | | `requestMethod` | 1023 |
| `cs4Label` | 1023 | | `shost` | 1023 |
| `cs5` | 4000 | | `sntdom` | 255 |
| `cs5Label` | 1023 | | `sourceDnsDomain` | 255 |
| `cs6` | 4000 | | `sourceServiceName` | 1023 |
| `cs6Label` | 1023 | | `sourceTranslatedZoneExternalID` | 200 |
| `customerExternalID` | 200 | | `sourceTranslatedZoneURI` | 2048 |
| `customerURI` | 2048 | | `sourceZoneExternalID` | 200 |
| `destinationDnsDomain` | 255 | | `sourceZoneURI` | 2048 |
| `destinationServiceName` | 1023 | | `spriv` | 1023 |
| `destinationTranslatedZoneExternalID` | 200 | | `sproc` | 1023 |
| `destinationTranslatedZoneURI` | 2048 | | `suid` | 1023 |
| `destinationZoneExternalID` | 200 | | `suser` | 1023 |
| `destinationZoneURI` | 2048 | | | |
| `deviceCustomDate1Label` | 1023 | | | |
| `deviceCustomDate2Label` | 1023 | | | |
| `deviceDnsDomain` | 255 | | | |
| `deviceExternalId` | 255 | | | |
| `deviceFacility` | 1023 | | | |
| `deviceInboundInterface` | 128 | | | |
| `deviceNtDomain` | 255 | | | |
| `deviceOutboundInterface` | 128 | | | |
| `devicePayloadId` | 128 | | | |
| `deviceProcessName` | 1023 | | | |
| `deviceTranslatedZoneExternalID` | 200 | | | |
| `deviceTranslatedZoneURI` | 2048 | | | |
| `deviceZoneExternalID` | 200 | | | |
| `deviceZoneURI` | 2048 | | | |
| `dhost` | 1023 | | | |
| `dntdom` | 255 | | | |
| `dpriv` | 1023 | | | |
| `dproc` | 1023 | | | |
| `dtz` | 255 | | | |

This table is kept in sync with `ExtensionFieldLimits` in `limits.go` — the
map there is the source of truth for code; this file exists for humans
(and Claude) to check against the spec without re-fetching the PDF.
