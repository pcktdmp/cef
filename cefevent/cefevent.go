package cefevent

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type CefEventer interface {
	Generate() string
}

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

func cefEscapeField(field string) string {

	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"|", "\\|",
		"\n", "\\n",
	)

	return replacer.Replace(field)
}

func cefEscapeExtension(field string) string {

	replacer := strings.NewReplacer(
		"\\", "\\\\", "\n",
		"\\n", "=", "\\=",
	)

	return replacer.Replace(field)
}

func (event *CefEvent) Generate() (string, error) {

	assertEvent := reflect.ValueOf(event).Elem()

	// define an array with all the mandatory
	// CEF fields.
	mandatoryFields := []string{
		"Version",
		"DeviceVendor",
		"DeviceProduct",
		"DeviceVersion",
		"DeviceEventClassId",
		"Name",
		"Severity",
	}

	// loop over all mandatory fields
	// and verify if they are not empty
	// according to their String type.
	for _, field := range mandatoryFields {

		if assertEvent.FieldByName(field).String() == "" {
			return "", errors.New("Not all mandatory CEF fields are set.")
		}
	}

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

	eventCef := fmt.Sprintf(
		"CEF:%v|%v|%v|%v|%v|%v|%v|%v",
		event.Version, event.DeviceVendor,
		event.DeviceProduct, event.DeviceVersion,
		event.DeviceEventClassId, event.Name,
		event.Severity, extensionString,
	)

	return eventCef, nil
}
