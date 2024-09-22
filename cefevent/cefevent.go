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

// CefEventer defines the interface for handling Common Event Format (CEF) events.
// It includes methods to generate, validate, read, and log CEF events.
type CefEventer interface {
	Generate() (string, error)          // Generate constructs and returns a CEF message string if all the mandatory fields are set.
	Validate() bool                     // Validate checks whether all mandatory fields in the CefEvent struct are set.
	Read(line string) (CefEvent, error) // Read parses a CEF message string and populates the CefEvent struct with the extracted data.
	Log() (bool, error)                 // Log attempts to generate a CEF message from the current CefEvent and logs it to the standard output.
}

// CefEvent represents a Common Event Format (CEF) event.
// It includes fields for CEF version, device vendor, device product, device version,
// device event class ID, event name, event severity, and additional extensions.
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

// cefEscapeField escapes special characters in a given string that are used in CEF (Common Event Format) fields.
// It replaces backslashes, pipes, and newlines with their escaped counterparts.
//
// The following replacements are performed:
// - "\" becomes "\\"
// - "|" becomes "\\|"
// - "\n" becomes "\\n"
//
// Parameters:
// - field: A string that needs to be escaped.
//
// Returns:
// - A string with the special characters escaped to ensure proper formatting in CEF fields.
func cefEscapeField(field string) string {

	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"|", "\\|",
		"\n", "\\n",
	)

	return replacer.Replace(field)
}

// cefEscapeExtension escapes special characters in a given string that are used in CEF (Common Event Format) extensions.
// It replaces backslashes, newlines, and equals signs with their escaped counterparts.
//
// The following replacements are performed:
// - "\" becomes "\\"
// - "\n" becomes "\\n"
// - "=" becomes "\\="
//
// Parameters:
// - field: A string that needs to be escaped.
//
// Returns:
// - A string with the special characters escaped to ensure proper formatting in CEF extensions.
func cefEscapeExtension(field string) string {

	replacer := strings.NewReplacer(
		"\\", "\\\\", "\n",
		"\\n", "=", "\\=",
	)

	return replacer.Replace(field)
}

// Validate verifies whether all mandatory fields in the CefEvent struct are set.
// It checks if the fields Version, DeviceVendor, DeviceProduct, DeviceVersion,
// DeviceEventClassId, Name, and Severity are populated and returns true if they are,
// otherwise, it returns false.
//
// This method uses reflection to loop over the mandatory fields and check their values.
//
// Returns:
// - A boolean indicating whether all mandatory fields are set (true) or not (false).
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

// Log attempts to generate a CEF message from the current CefEvent
// and logs it to the standard output. If generation fails, it logs
// an error message to the standard error.
//
// The method works well within containerized environments by appropriately
// selecting output targets based on success or failure. If the generation
// of the event is successful, it logs the message to stdout, otherwise,
// it logs an error message to stderr.
//
// Returns:
// - A boolean indicating whether the logging operation succeeded (true) or failed (false).
// - An error if the logging operation could not be completed due to a failure in generating the CEF message.
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

// Generate constructs and returns a CEF (Common Event Format) message string if all the mandatory
// fields are set in the CefEvent. If any mandatory field is missing, it returns an error.
//
// A CEF message follows the format:
// CEF:Version|Device Vendor|Device Product|Device Version|Device Event Class ID|Name|Severity|Extensions
//
// Each field is escaped to ensure that special characters do not interfere with the CEF format.
//
// Returns:
// - A string representing the generated CEF message.
// - An error if any mandatory field is missing or if there are other issues during generation.
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

// Read parses a CEF (Common Event Format) message string and populates the CefEvent struct
// with the extracted data.
//
// The method checks if the provided string starts with the "CEF:" prefix and then splits
// the string into its constituent fields. It also extracts any key-value pairs present in the
// Extensions part of the CEF message.
//
// The format of a CEF message is:
// CEF:Version|Device Vendor|Device Product|Device Version|Device Event Class ID|Name|Severity|Extensions
//
// The method ensures that if any mandatory field is missing or improperly formatted, it returns an error.
//
// Returns:
// - A CefEvent struct populated with the parsed CEF message data.
// - An error if the CEF message is improperly formatted or if any mandatory field is missing.
func (event CefEvent) Read(eventLine string) (CefEvent, error) {
	if strings.HasPrefix(eventLine, "CEF:") {
		eventSlashed := strings.Split(strings.TrimPrefix(eventLine, "CEF:"), "|")

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

		if !CefEventer.Validate(&eventParsed) {
			return CefEvent{}, errors.New("not all mandatory CEF fields are set")
		}

		return eventParsed, nil
	}
	return CefEvent{}, errors.New("not a valid CEF message")
}
