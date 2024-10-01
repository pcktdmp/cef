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
	err := event.Log()

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
