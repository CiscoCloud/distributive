package systemdstatus

import (
	"os/exec"
	"strings"
	"testing"
)

var dummyServices = []string{"foo", "bar", "ipsum lorem", "541"}

func TestServiceLoaded(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("systemctl"); err != nil {
		t.Logf("Couldn't find systemctl binary, skipping test %v", err)
	} else {
		for _, name := range dummyServices {
			_, err := ServiceLoaded(name)
			if err != nil && strings.Contains(err.Error(), "D-Bus connection") {
				t.Logf("Systemd failed (probably in Docker Ubuntu)")
			} else if err != nil {
				t.Errorf("ServiceLoaded failed on input %v with error %v", name, err)
			}
		}
	}
}

func TestServiceActive(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("systemctl"); err != nil {
		t.Logf("Couldn't find systemctl binary, skipping test ")
	} else {
		for _, name := range dummyServices {
			_, err := ServiceActive(name)
			if err != nil && strings.Contains(err.Error(), "D-Bus connection") {
				t.Logf("Systemd failed (probably in Docker Ubuntu)")
			} else if err != nil {
				t.Errorf("ServiceActive failed on input %v with error %v", name, err)
			}
		}
	}
}

func TestListeningSockets(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("systemctl"); err != nil {
		t.Logf("Couldn't find systemctl binary, skipping test ")
	} else {
		_, err := ListeningSockets()
		if err != nil && strings.Contains(err.Error(), "D-Bus connection") {
			t.Logf("Systemd failed (probably in Docker Ubuntu)")
		} else if err != nil {
			t.Errorf("ListeningSockets failed with error %v", err)
		}
	}
}

func TestTimers(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("systemctl"); err != nil {
		t.Logf("Couldn't find systemctl binary, skipping test ")
	} else {
		for _, all := range []bool{true, false} {
			_, err := Timers(all)
			if err != nil && strings.Contains(err.Error(), "D-Bus connection") {
				t.Logf("Systemd failed (probably in Docker Ubuntu)")
			} else if err != nil {
				t.Errorf("Timers failed on input %v with error %v", all, err)
			}
		}
	}
}

func TestUnitFileStatuses(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("systemctl"); err != nil {
		t.Logf("Couldn't find systemctl binary, skipping test ")
	} else {
		_, _, err := UnitFileStatuses()
		if err != nil && strings.Contains(err.Error(), "D-Bus connection") {
			t.Logf("Systemd failed (probably in Docker Ubuntu)")
		} else if err != nil {
			t.Errorf("UnitFileStatuses failed with error %v", err)
		}
	}
}
