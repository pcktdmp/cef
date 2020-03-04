# cef
Go Package for ArcSight's CEF

## Not yet implemented

* Field limits according to format standard for [known](https://community.microfocus.com/t5/ArcSight-Connectors/ArcSight-Common-Event-Format-CEF-Implementation-Standard/ta-p/1645557?attachment-id=68077) CEF fields
* Error handling
* Mandatory header field checking

## TL;DR

`go get github.com/pcktdmp/cef/cefevent`

```
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
		Version:            "1",
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
