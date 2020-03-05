package cefevent

import (
	"fmt"
	"strings"
)

type CefEvent struct {
	Version            string
	DeviceVendor       string
	DeviceProduct      string
	DeviceVersion      string
	DeviceEventClassId string
	Name               string
	Severity           string
	Extensions         map[string]string
}

// todo: don't dupe the function but handle
// with methods part of an Extension struct
func cefEscapeField(field string) string {

	replacer := strings.NewReplacer(
		"\\", "\\\\", "|", "\\|",
		"\n", "\\n", "=", "\\n",
	)

	return replacer.Replace(field)
}

// todo: don't dupe the function but handle
// with methods part of an Extension struct
func cefEscapeExtension(field string) string {

	replacer := strings.NewReplacer(
		"\\", "\\\\", "\n",
		"\\n", "=", "\\n",
	)

	return replacer.Replace(field)
}

func (event CefEvent) Generate() string {

	// todo: do this nicely with methods
	event.Version = cefEscapeField(event.Version)
	event.DeviceVendor = cefEscapeField(event.DeviceVendor)
	event.DeviceProduct = cefEscapeField(event.DeviceProduct)
	event.DeviceVersion = cefEscapeField(event.DeviceVersion)
	event.DeviceEventClassId = cefEscapeField(event.DeviceEventClassId)
	event.Name = cefEscapeField(event.Name)
	event.Severity = cefEscapeField(event.Severity)

	var p strings.Builder

	// construct the extension string according to the CEF format
	for k, v := range event.Extensions {
		p.WriteString(fmt.Sprintf(
			"%s=%s ",
			cefEscapeExtension(k),
			cefEscapeExtension(v)),
		)
	}

	// make sure there is not a trailing space for the extension
	// fields according to the CEF standard.
	extensionString := strings.TrimSpace(p.String())

	return fmt.Sprintf(
		"CEF:%v|%v|%v|%v|%v|%v|%v|%v",
		event.Version, event.DeviceVendor,
		event.DeviceProduct, event.DeviceVersion,
		event.DeviceEventClassId, event.Name,
		event.Severity, extensionString)
}
