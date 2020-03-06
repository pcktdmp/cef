# Common Event Format in Go
Go Package for ArcSight's Common Event Format

![Test Workflow](https://github.com/pcktdmp/cef/workflows/Test/badge.svg)
![Build Workflow](https://github.com/pcktdmp/cef/workflows/Build/badge.svg)

# Motivation

Learning Go, help people who need to generate CEF events in Golang.

## TL;DR

### Install the package

```bash
$ go get github.com/pcktdmp/cef/cefevent
```

### cef.go

```go
package main

import (
	"fmt"
	"github.com/pcktdmp/cef/cefevent"
)

func main() {

	f := make(map[string]string)
	f["sourceAddress"] = "127.0.0.1"
	f["requestClientApplication"] = "Go-http-client/1.1"

	event := cefevent.CefEvent{
		Version:            "0",
		DeviceVendor:       "Cool Vendor",
		DeviceProduct:      "Cool Product",
		DeviceVersion:      "1.0",
		DeviceEventClassId: "FLAKY_EVENT",
		Name:               "Something flaky happened.",
		Severity:           "3",
		Extensions:         f}.Generate()

	fmt.Println(event)

}

```
### Example output

```bash
$ ./cef
CEF:1|Cool Vendor|Cool Product|1.0|FLAKY_EVENT|Something flaky happened.|3|sourceAddress=127.0.0.1 requestClientApplication=Go-http-client/1.1
```

## Not yet implemented

* Field limits according to format standard for [known](https://community.microfocus.com/t5/ArcSight-Connectors/ArcSight-Common-Event-Format-CEF-Implementation-Standard/ta-p/1645557?attachment-id=68077) CEF fields
* Error handling
* Mandatory header field checking
