//go:build !windows && !plan9

package cefevent

import (
	"log/syslog"
	"testing"
)

func TestCefEventLogToSyslog(t *testing.T) {

	// Not every environment (containers, CI sandboxes) has a reachable local
	// syslog daemon; skip rather than fail when that's the case, since that's
	// an environment limitation, not a code bug.
	probe, err := syslog.New(syslog.LOG_INFO, "cefevent-test")
	if err != nil {
		t.Skipf("no local syslog available in this environment: %v", err)
	}
	probe.Close()

	e := event

	if err := e.LogToSyslog(syslog.LOG_INFO, "cefevent-test"); err != nil {
		t.Errorf("LogToSyslog() = %v, want nil", err)
	}
}

func TestCefEventLogToSyslogMandatoryFieldMissing(t *testing.T) {

	// String() fails before LogToSyslog ever attempts to reach syslog, so
	// this doesn't need a reachable local syslog daemon to run.
	brokenEvent := event
	brokenEvent.DeviceVendor = ""

	if err := brokenEvent.LogToSyslog(syslog.LOG_INFO, "cefevent-test"); err == nil {
		t.Errorf("LogToSyslog() = nil, want an error for a missing mandatory field")
	}
}
