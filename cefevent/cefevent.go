package cefevent

import (
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type CefEventer interface {
	Generate() (string, error)
	Validate() bool
	// TODO: implement read feature for just Parsed() events.
	Read() (CefEvent, error)
	Log() (bool, error)
}

type CefEvent struct {
	// defaults to 0 which is also the first CEF version.
	Version            int
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

func (event *CefEvent) Validate() bool {

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
			return false
		}
	}

	return true

}

// Log should be used as a stub in most cases, it either
// succeeds generating the CEF event and send it to stdout
// or doesn't and logs that to stderr. This function
// plays well inside containers.
func (event *CefEvent) Log() (bool, error) {

	logMessage, err := event.Generate()

	if err != nil {
		log.SetOutput(os.Stderr)
		errMsg := "unable to generate and thereby log the CEF message"
		log.Println(errMsg)
		return false, errors.New(errMsg)
	}

	log.SetOutput(os.Stdout)
	log.Println(logMessage)
	return true, nil
}

func (event CefEvent) Generate() (string, error) {

	if !CefEventer.Validate(&event) {
		return "", errors.New("not all mandatory CEF fields are set")
	}

	event.DeviceVendor = cefEscapeField(event.DeviceVendor)
	event.DeviceProduct = cefEscapeField(event.DeviceProduct)
	event.DeviceVersion = cefEscapeField(event.DeviceVersion)
	event.DeviceEventClassId = cefEscapeField(event.DeviceEventClassId)
	event.Name = cefEscapeField(event.Name)
	event.Severity = cefEscapeField(event.Severity)

	var p strings.Builder

	var sortedExtensions []string
	for k := range event.Extensions {
		sortedExtensions = append(sortedExtensions, k)
	}
	sort.Strings(sortedExtensions)

	// construct the extension string according to the CEF format
	for _, k := range sortedExtensions {
		p.WriteString(fmt.Sprintf(
			"%s=%s ",
			cefEscapeExtension(k),
			cefEscapeExtension(event.Extensions[k])),
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

func Parse(eventLine string) (CefEvent, error) {
	if strings.HasPrefix(string(eventLine), "CEF:") {
		eventSlashed := strings.Split(strings.TrimPrefix(string(eventLine), "CEF:"), "|")

		// convert CEF version to int
		cefVersion, err := strconv.Atoi(eventSlashed[0])
		if err != nil {
			return CefEvent{}, err
		}

		parsedExtensions := make(map[string]string)

		// each extension k,v is separated by a " ".
		// in the substring, "=" separator defines the kv pair of the extension
		if len(eventSlashed) >= 7 {
			extensions := strings.Split(eventSlashed[7], " ")
			for _, ext := range extensions {
				kv := strings.SplitN(ext, "=", 2)
				if len(kv) == 2 {
					parsedExtensions[kv[0]] = kv[1]
				}
			}
		}

		eventParsed := CefEvent{
			Version:            cefVersion,
			DeviceVendor:       eventSlashed[1],
			DeviceProduct:      eventSlashed[2],
			DeviceVersion:      eventSlashed[3],
			DeviceEventClassId: eventSlashed[4],
			Name:               eventSlashed[5],
			Severity:           eventSlashed[6],
			Extensions:         parsedExtensions,
		}
		return eventParsed, nil
	}
	return CefEvent{}, errors.New("not a valid CEF message")
}
