//go:build !windows && !plan9

package cefevent

import (
	"log/syslog"
)

// LogToSyslog generates a CEF message via String() and writes it to the local
// syslog daemon at the given priority and tag, using the standard log/syslog
// package. It's the syslog-native counterpart to Log(), which always writes
// to stdout/stderr via the standard log package instead — matching how CEF is
// actually shipped in most real deployments.
//
// priority combines both severity and facility, exactly as it does for
// syslog.New (e.g. syslog.LOG_ERR|syslog.LOG_USER).
//
// This method doesn't exist on windows or plan9: log/syslog itself has no
// implementation for either, so a file carrying this method is excluded from
// those builds via a build constraint, the same as log/syslog's own source
// does internally.
func (event *CefEvent) LogToSyslog(priority syslog.Priority, tag string) error {

	message, err := event.String()
	if err != nil {
		return err
	}

	writer, err := syslog.New(priority, tag)
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = writer.Write([]byte(message))
	return err
}
