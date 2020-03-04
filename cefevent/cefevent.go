package cefevent

import (
	"fmt"
	"strings"
)

type CefEventInterface interface {
}

type CefEvent struct {
	version            string
	deviceVendor       string
	deviceProduct      string
	deviceVersion      string
	deviceEventClassId string
	name               string
	severity           string
	extensions         map[string]string
}

// todo: don't dupe the function but handle
// with methods part of an Extension struct
func cefEscapeField(field string) string {

	field = strings.ReplaceAll(field, "\\", "\\\\")
	field = strings.ReplaceAll(field, "|", "\\|")
	field = strings.ReplaceAll(field, "\n", "\\n")
	field = strings.ReplaceAll(field, "=", "\\=")

	return field
}

// todo: don't dupe the function but handle
// with methods part of an Extension struct
func cefEscapeExtension(field string) string {

	field = strings.ReplaceAll(field, "\\", "\\\\")
	field = strings.ReplaceAll(field, "\n", "\\n")
	field = strings.ReplaceAll(field, "=", "\\=")

	return field
}

func (event CefEvent) Generate() string {

	// todo: do this nicely with methods
	event.version = cefEscapeField(event.version)
	event.deviceVendor = cefEscapeField(event.deviceVendor)
	event.deviceProduct = cefEscapeField(event.deviceProduct)
	event.deviceVersion = cefEscapeField(event.deviceVersion)
	event.deviceEventClassId = cefEscapeField(event.deviceEventClassId)
	event.name = cefEscapeField(event.name)
	event.severity = cefEscapeField(event.severity)

	var extension_string string

	// construct the extension string according to the CEF format
	for k, v := range event.extensions {
		extension_string += fmt.Sprintf("%s=%s ", cefEscapeExtension(k), cefEscapeExtension(v))
	}

	// make sure there is not a trailing space for the extension
	// fields according to the CEF standard.
	p := &extension_string
	*p = strings.TrimRight(extension_string, " ")

	return fmt.Sprintf("CEF:%v|%v|%v|%v|%v|%v|%v|%v", event.version,
		event.deviceVendor, event.deviceProduct,
		event.deviceVersion, event.deviceEventClassId,
		event.name, event.severity, extension_string)
}
