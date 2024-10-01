# Common Event Format in Go
Go Package for ArcSight's Common Event Format

![Build Workflow](https://github.com/pcktdmp/cef/workflows/Build/badge.svg)
![Test Workflow](https://github.com/pcktdmp/cef/workflows/Test/badge.svg)

# Motivation

Learning Go, help people who need to process CEF events in Golang.

## TL;DR

`cefevent` is a [loose implementation](#Not-implemented) of the Common Event Format, the one who processes events
needs to handle the [known](https://www.microfocus.com/documentation/arcsight/arcsight-smartconnectors-8.4/pdfdoc/cef-implementation-standard/cef-implementation-standard.pdf) field limits.

### Install the package

```bash
$ go get github.com/pcktdmp/cef/cefevent
```

### examples.go

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

	fmt.Println(event.String())

	// send a CEF event as log message to stdout
	event.Log()

	// or if you want to do error handling when
	// sending the log
	_, err := event.Log()

	if err != nil {
		fmt.Println("Need to handle this.")
	}

	// if you want read a CEF event from a line
	eventLine := "CEF:0|Cool Vendor|Cool Product|1.0|COOL_THING|Something cool happened.|Unknown|src=127.0.0.1"
	newEvent := cefevent.CefEvent{}
	newEvent.Read(eventLine)
	eventString, err := newEvent.String()
	if err != nil {
		fmt.Println("Need to handle this.")
	}
	fmt.Println(eventString)

}

```
### Example output

```bash
$ go run examples.go
CEF:0|Cool Vendor|Cool Product|1.0|FLAKY_EVENT|Something flaky happened.|3|requestClientApplication=Go-http-client/1.1 src=127.0.0.1
2020/03/12 21:28:19 CEF:0|Cool Vendor|Cool Product|1.0|FLAKY_EVENT|Something flaky happened.|3|requestClientApplication=Go-http-client/1.1 src=127.0.0.1
2020/03/12 21:28:19 CEF:0|Cool Vendor|Cool Product|1.0|FLAKY_EVENT|Something flaky happened.|3|requestClientApplication=Go-http-client/1.1 src=127.0.0.1
```

## Not implemented

* Field limits according to format standard for CEF fields
