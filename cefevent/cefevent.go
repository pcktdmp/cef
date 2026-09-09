package cefevent

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// CefEventer defines the interface for handling Common Event Format (CEF) events.
// It includes methods to create (String()), Validate(), Read(), and Log() CEF events.
type CefEventer interface {
	Validate() error                    // Validate if the CEF message is according to the specification.
	String() (string, error)            // String constructs and returns a CEF message string.
	Build() (CefEvent, error)           // Build constructs and returns a CEF message according to CefEvent.
	Read(line string) (CefEvent, error) // Read parses a CEF message string and populates the CefEvent struct with the extracted data.
	Log() error                         // Log attempts to generate a CEF message from the current CefEvent and logs it to the standard output.
	escapeEventData() error             // escapeEventData will try to escape all data properly in the struct according the Common Event Format.
}

// CefEvent represents a Common Event Format (CEF) event.
// It includes fields for CEF version, device vendor, device product, device version,
// device event class ID, event name, event severity, and additional extensions.
type CefEvent struct {
	// defaults to 0 which is also the first CEF version.
	Version            int               `json:"Version" yaml:"Version" toml:"Version" xml:"Version" header:"CEF Version" comment:"The version of the CEF specification that the event conforms to."`
	DeviceVendor       string            `json:"DeviceVendor" yaml:"DeviceVendor" toml:"DeviceVendor" xml:"DeviceVendor" header:"Device Vendor" comment:"The name of the device vendor."`
	DeviceProduct      string            `json:"DeviceProduct" yaml:"DeviceProduct" toml:"DeviceProduct" xml:"DeviceProduct" header:"Device Product" comment:"The name of the device product."`
	DeviceVersion      string            `json:"DeviceVersion" yaml:"DeviceVersion" toml:"DeviceVersion" xml:"DeviceVersion" header:"Device Version" comment:"The version of the device product."`
	DeviceEventClassId string            `json:"DeviceEventClassId" yaml:"DeviceEventClassId" toml:"DeviceEventClassId" xml:"DeviceEventClassId" header:"Device Event Class ID" comment:"The ID of the event class that the event conforms to."`
	Name               string            `json:"Name" yaml:"Name" toml:"Name" xml:"Name" header:"Name" comment:"The name of the event."`
	Severity           string            `json:"Severity" yaml:"Severity" toml:"Severity" xml:"Severity" header:"Severity" comment:"The severity of the event."`
	Extensions         map[string]string `json:"Extensions,omitempty" yaml:"Extensions" toml:"Extensions" xml:"Extensions" header:"Extensions" comment:"Additional extensions to the CEF message."`
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

// cefUnescapeField reverses cefEscapeField: "\\" becomes "\", "\|" becomes "|", and
// "\n" (a literal backslash followed by "n") becomes a newline character. An
// unrecognized escape sequence (a "\" followed by anything else) is passed through
// unchanged rather than dropping the backslash.
func cefUnescapeField(field string) string {

	var b strings.Builder
	b.Grow(len(field))

	for i := 0; i < len(field); i++ {
		if field[i] == '\\' && i+1 < len(field) {
			switch field[i+1] {
			case '\\':
				b.WriteByte('\\')
			case '|':
				b.WriteByte('|')
			case 'n':
				b.WriteByte('\n')
			default:
				b.WriteByte(field[i])
				b.WriteByte(field[i+1])
			}
			i++
			continue
		}
		b.WriteByte(field[i])
	}

	return b.String()
}

// cefUnescapeExtension reverses cefEscapeExtension: "\\" becomes "\", "\n" (a literal
// backslash followed by "n") becomes a newline character, and "\=" becomes "=". An
// unrecognized escape sequence is passed through unchanged rather than dropping the
// backslash.
func cefUnescapeExtension(field string) string {

	var b strings.Builder
	b.Grow(len(field))

	for i := 0; i < len(field); i++ {
		if field[i] == '\\' && i+1 < len(field) {
			switch field[i+1] {
			case '\\':
				b.WriteByte('\\')
			case 'n':
				b.WriteByte('\n')
			case '=':
				b.WriteByte('=')
			default:
				b.WriteByte(field[i])
				b.WriteByte(field[i+1])
			}
			i++
			continue
		}
		b.WriteByte(field[i])
	}

	return b.String()
}

// splitUnescapedPipes splits s on "|" characters that are not escaped with a
// preceding "\" (mirroring the escaping cefEscapeField applies to header fields, so
// a header field containing an escaped "\|" is not mistaken for a field separator).
//
// If n > 0, splitting stops after n-1 splits and the final element holds the
// unprocessed remainder of s, exactly like strings.SplitN — this is what lets Read
// treat everything after the 7th unescaped "|" as one opaque extensions blob, even if
// it contains further (unescaped, and therefore legitimate per the CEF spec) "|"
// characters. n <= 0 splits on every unescaped "|".
func splitUnescapedPipes(s string, n int) []string {

	var parts []string
	start := 0

	for i := 0; i < len(s); i++ {
		if n > 0 && len(parts) == n-1 {
			break
		}
		if s[i] == '\\' && i+1 < len(s) {
			i++
			continue
		}
		if s[i] == '|' {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}

	parts = append(parts, s[start:])
	return parts
}

// extensionKeyPattern finds the start of a CEF extension key: a run of letters,
// digits, "_" and "." beginning with a letter, immediately followed by "=", and
// preceded by either the start of the string or a single whitespace character. CEF
// key names are never escaped, so this deliberately doesn't need to worry about
// escape sequences the way header-field splitting does.
//
// The spec's own "Custom Extension Naming Guidelines" call for custom keys to be a
// single alphanumeric word, but real-world producers routinely ignore that and use
// "_"/"." anyway (e.g. underscore-separated names) — matching String()'s existing
// leniency (it never validates key format at all), this stays permissive rather than
// rejecting keys the spec's guidelines merely recommend against.
//
// Using a single "\s" (rather than "\s*") is deliberate: per the CEF spec, when
// multiple spaces precede a key, only the last of those spaces is the delimiter —
// the rest belong to the previous value as trailing whitespace. A regexp search
// scans left to right for the earliest match, so this pattern naturally lands on
// that last space.
var extensionKeyPattern = regexp.MustCompile(`(?:^|\s)([A-Za-z][A-Za-z0-9_.]*)=`)

// splitExtensions parses a raw (still-escaped) CEF extension blob into its key/value
// pairs. Values are returned still-escaped — apply cefUnescapeExtension to each
// before using it — and may legitimately contain literal spaces, per the CEF spec's
// own example ("filePath=/user/username/dir/my file name.txt" is a single key/value
// pair, not two). Trailing spaces are trimmed from the final value only, per the same
// spec note that motivates extensionKeyPattern's single-space matching.
//
// This can misparse a value that happens to contain literal " key=" text that was
// not actually meant as a new key — that ambiguity is inherent to the CEF extension
// format itself (it isn't reliably resolvable without a stricter grammar than the
// spec defines), not something specific to this implementation.
func splitExtensions(blob string) map[string]string {

	result := make(map[string]string)

	matches := extensionKeyPattern.FindAllStringSubmatchIndex(blob, -1)
	if len(matches) == 0 {
		return result
	}

	for i, m := range matches {
		key := blob[m[2]:m[3]]

		valueStart := m[1]
		valueEnd := len(blob)
		if i+1 < len(matches) {
			valueEnd = matches[i+1][0]
		}

		value := blob[valueStart:valueEnd]
		if i == len(matches)-1 {
			value = strings.TrimRight(value, " ")
		}

		result[key] = value
	}

	return result
}

// escapeEventData processes and escapes all necessary fields within the CefEvent struct according
// to the Common Event Format (CEF) specifications. It ensures that fields such as DeviceVendor,
// DeviceProduct, DeviceVersion, DeviceEventClassId, Name, Severity, and Extensions have their
// special characters escaped properly to maintain the integrity of the CEF message.
//
// This function performs the following steps:
//   - Escapes special characters in fields like DeviceVendor, DeviceProduct, DeviceVersion,
//     DeviceEventClassId, Name, and Severity using the cefEscapeField helper function.
//   - Iterates over the Extensions map and escapes both the keys and values using the
//     cefEscapeExtension helper function, ensuring no duplicated keys in the resulting map.
//
// Returns:
// - An error if there is any issue during the escaping process; otherwise, returns nil.
func (event *CefEvent) escapeEventData() error {

	event.DeviceVendor = cefEscapeField(event.DeviceVendor)
	event.DeviceProduct = cefEscapeField(event.DeviceProduct)
	event.DeviceVersion = cefEscapeField(event.DeviceVersion)
	event.DeviceEventClassId = cefEscapeField(event.DeviceEventClassId)
	event.Name = cefEscapeField(event.Name)
	event.Severity = cefEscapeField(event.Severity)

	// TODO: memory usage improvement
	// simple method to make sure escaped strings are not duped in the map keys
	escapedExtensions := make(map[string]string)

	if len(event.Extensions) > 0 {
		for k, v := range event.Extensions {
			escapedExtensions[cefEscapeExtension(k)] = cefEscapeExtension(v)
		}
	}

	event.Extensions = escapedExtensions

	return nil
}

// Validate verifies whether all mandatory fields in the CefEvent struct are set.
// It checks if the fields Version, DeviceVendor, DeviceProduct, DeviceVersion,
// DeviceEventClassId, Name, and Severity are populated and returns nil if they are,
// otherwise, it returns an error.
//
// This method uses reflection to loop over the mandatory fields and check their values.
//
// Returns:
// - An error message indicating whether all mandatory fields are set (err) or not (nil).
func (event *CefEvent) Validate() error {

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
			return errors.New("not all mandatory CEF fields are set")
		}
	}

	return nil
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
// - An error indicating whether the logging operation succeeded (nil) or failed (err).
func (event *CefEvent) Log() error {

	logMessage, err := event.String()

	if err != nil {
		log.SetOutput(os.Stderr)
		errMsg := "unable to create and thereby log the CEF message"
		log.Println(errMsg)
		return errors.New(errMsg)
	}

	log.SetOutput(os.Stdout)
	log.Println(logMessage)
	return nil
}

// Build constructs and returns a CEF (Common Event Format) message just as String() but then as CefEvent type.
//
// Returns:
// - A CefEvent type representing the CEF message.
// - An error if any mandatory field is missing or if there are other issues during generation.
func (event *CefEvent) Build() (CefEvent, error) {

	if event.Validate() != nil {
		return CefEvent{}, errors.New("not all mandatory CEF fields are set")
	}

	if event.escapeEventData() != nil {
		return CefEvent{}, errors.New("unable to escape CEF event data")
	}

	return *event, nil
}

// String constructs and returns a CEF (Common Event Format) message string if all the mandatory
// fields are set in the CefEvent. If any mandatory field is missing, it returns an error.
//
// A CEF message follows the format:
// CEF:Version|Device Vendor|Device Product|Device Version|Device Event Class ID|Name|Severity|Extensions
//
// Each field is escaped to ensure that special characters do not interfere with the CEF format.
//
// Returns:
// - A string representing the CEF message.
// - An error if any mandatory field is missing or if there are other issues during generation.
func (event *CefEvent) String() (string, error) {

	if CefEventer.Validate(event) != nil {
		return "", errors.New("not all mandatory CEF fields are set")
	}

	if event.escapeEventData() != nil {
		return "", errors.New("unable to escape CEF event data")
	}

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
			k,
			event.Extensions[k]),
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
func (event *CefEvent) Read(eventLine string) (CefEvent, error) {
	if strings.HasPrefix(eventLine, "CEF:") {
		// Split on the first 7 unescaped "|" characters only: this both honors an
		// escaped "\|" inside a header field (so it isn't mistaken for the field
		// separator) and treats everything after the 7th as one opaque extensions
		// blob, even if it contains further "|" characters — which is legitimate
		// there, since the spec only requires escaping "|" in header fields.
		eventSlashed := splitUnescapedPipes(strings.TrimPrefix(eventLine, "CEF:"), 8)

		// there must be at least the 7 mandatory header fields
		// (Version, DeviceVendor, DeviceProduct, DeviceVersion,
		// DeviceEventClassId, Name, Severity) before we can safely
		// index into eventSlashed below.
		if len(eventSlashed) < 7 {
			return CefEvent{}, errors.New("not a valid CEF message")
		}

		// convert CEF version to int
		cefVersion, err := strconv.Atoi(eventSlashed[0])
		if err != nil {
			return CefEvent{}, err
		}

		event.Version = cefVersion
		parsedExtensions := make(map[string]string)

		// splitExtensions locates key/value pairs by their " key=" boundaries
		// rather than naively splitting on every space, since extension values
		// may legitimately contain spaces. Its values come back still-escaped,
		// hence cefUnescapeExtension below.
		if len(eventSlashed) >= 8 {
			for k, v := range splitExtensions(eventSlashed[7]) {
				parsedExtensions[k] = cefUnescapeExtension(v)
			}
		}

		// The struct's fields always hold raw, unescaped data — the same
		// invariant a freshly constructed CefEvent{} has — so String()/Build()
		// can escape it symmetrically when generating a message. Unescaping
		// here (rather than the former re-escaping) is what makes Read the
		// correct inverse of String/Build.
		event.DeviceVendor = cefUnescapeField(eventSlashed[1])
		event.DeviceProduct = cefUnescapeField(eventSlashed[2])
		event.DeviceVersion = cefUnescapeField(eventSlashed[3])
		event.DeviceEventClassId = cefUnescapeField(eventSlashed[4])
		event.Name = cefUnescapeField(eventSlashed[5])
		event.Severity = cefUnescapeField(eventSlashed[6])
		event.Extensions = parsedExtensions

		if CefEventer.Validate(event) != nil {
			return CefEvent{}, errors.New("not all mandatory CEF fields are set")
		}

		return *event, nil
	}
	return CefEvent{}, errors.New("not a valid CEF message")
}

// ToJSON converts the CefEvent instance to a JSON string.
//
// This method first validates the CefEvent to ensure all mandatory fields are set,
// escapes the event data the same way String()/Build()/Read() do, and then attempts
// to marshal the event into a JSON formatted string.
//
// Returns:
// - A JSON string representation of the CefEvent if successful.
// - An error if the CefEvent is not valid or if there is an error during the JSON marshaling process.
func (event *CefEvent) ToJSON() (string, error) {

	if CefEventer.Validate(event) != nil {
		return "", errors.New("not all mandatory CEF fields are set")
	}

	if event.escapeEventData() != nil {
		return "", errors.New("unable to escape CEF event data")
	}

	// Attempt to convert the event to JSON
	jsonData, err := json.Marshal(event)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}
