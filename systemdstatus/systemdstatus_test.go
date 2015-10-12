package systemdstatus

import (
	"testing"
)

var dummyServices = []string{"foo", "bar", "ipsum lorem", "541"}

func TestServiceLoaded(t *testing.T) {
	t.Parallel()
	for _, name := range dummyServices {
		if _, err := ServiceLoaded(name); err != nil {
			t.Errorf("ServiceLoaded failed on input %s", name)
		}
	}
}

func TestServiceActive(t *testing.T) {
	t.Parallel()
	for _, name := range dummyServices {
		if _, err := ServiceActive(name); err != nil {
			t.Errorf("ServiceActive failed on input %s", name)
		}
	}
}

func TestListeningSockets(t *testing.T) {
	t.Parallel()
	if _, err := ListeningSockets(); err != nil {
		t.Error("ListeningSockets failed")
	}
}

func TestTimers(t *testing.T) {
	t.Parallel()
	for _, all := range []bool{true, false} {
		if _, err := Timers(all); err != nil {
			t.Errorf("Timers failed on input %s", all)
		}
	}
}

func TestUnitFileStatuses(t *testing.T) {
	t.Parallel()
	if _, _, err := UnitFileStatuses(); err != nil {
		t.Error("UnitFileStatuses failed")
	}
}
