# CEF field reference

This file is a cached, condensed reference for the CEF field limits and data
types implemented in `cefevent/limits.go` (`HeaderFieldLimits`,
`ExtensionFieldLimits`) and `cefevent/types.go` (`ExtensionFieldTypes`). It
exists so that changes to those tables can be checked against the spec
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
and update `cefevent/limits.go` (plus this file) together in the same change.

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

This table is kept in sync with `ExtensionFieldLimits` in `cefevent/limits.go` — the
map there is the source of truth for code; this file exists for humans
(and Claude) to check against the spec without re-fetching the PDF.

## Extension Dictionary (non-string data types)

Every Extension Dictionary key also has a "Data Type" column in the spec beyond
`String` — Integer, Long, Floating Point, Double, IP Address, MAC Address, or Time
Stamp. `ExtensionFieldTypes` in `cefevent/types.go` encodes a **curated subset** of
these: only the keys below, each individually confirmed against the spec's own Data
Type column (not bulk-transcribed the way the length tables above were, since the
source PDF's line-wrapped rows made bulk extraction unreliable for this — see the
`git log` for `types.go` if you want the extraction notes). A key not listed here
simply isn't checked by `ValidateExtensionTypes`; that's not a claim that it has no
non-string type, just that it hasn't been individually verified yet.

| CEF key | Data Type | | CEF key | Data Type |
|---|---|---|---|---|
| `agt` | IP Address | | `dvcpid` | Integer |
| `amac` | MAC Address | | `end` | Time Stamp |
| `art` | Time Stamp | | `eventId` | Long |
| `cfp1`-`cfp4` | Floating Point | | `fsize` | Integer |
| `cn1`-`cn3` | Long | | `in` | Integer |
| `cnt` | Integer | | `oldFileSize` | Integer |
| `deviceDirection` | Integer | | `out` | Integer |
| `dlat`, `dlong` | Double | | `rt` | Time Stamp |
| `dmac` | MAC Address | | `slat`, `slong` | Double |
| `dpid` | Integer | | `smac` | MAC Address |
| `dpt` | Integer | | `spid` | Integer |
| `dst` | IP Address | | `spt` | Integer |
| `dvc` | IP Address | | `src` | IP Address |
| `dvcmac` | MAC Address | | `start` | Time Stamp |
| | | | `type` | Integer |

`dvcmac` is a known extraction quirk: the Version 26 PDF's table renders this row's
key column as `dmac` a second time (right where `dvcmac` belongs alphabetically,
between `dvc` and `dvcpid`), almost certainly a PDF kerning/font artifact rather than
the field actually being absent — the older v25 document and Microsoft's
[CEF-to-CommonSecurityLog mapping reference](https://learn.microsoft.com/en-us/azure/sentinel/cef-name-mapping)
both independently confirm `dvcmac` → `deviceMacAddress` → MAC Address.

`IP Address` fields accept both IPv4 and IPv6 in `ValidateExtensionTypes`
(`net.ParseIP`): the spec notes CEF 0.1 held IPv4 only in these fields, but CEF 1.0
onward allows IPv6 too, and rejecting valid IPv6 on the assumption a message is 0.1
would be over-strict.

`Time Stamp` fields accept the spec's documented `MMM dd yyyy HH:mm:ss` format (both
zero-padded and non-zero-padded day) or milliseconds-since-epoch — not the full list
of date formats the spec's later chapters describe elsewhere, just this one.
