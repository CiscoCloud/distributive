package checks

import (
	"testing"
)

var activeServices = [][]string{
	{"dev-mqueue.mount"}, {"tmp.mount"}, {"dbus.service"},
}

func TestSystemctlLoaded(t *testing.T) {
	t.Parallel()
	testParameters(names, notLengthOne, SystemctlLoaded{}, t)
	testCheck(activeServices, names, SystemctlLoaded{}, t)
}

func TestSystemctlActive(t *testing.T) {
	t.Parallel()
	testParameters(names, notLengthOne, SystemctlActive{}, t)
	testCheck(activeServices, names, SystemctlLoaded{}, t)
}

func TestSystemctlSockPath(t *testing.T) {
	t.Parallel()
	invalidInputs := append(notLengthOne, names...)
	testParameters(fileParameters, invalidInputs, SystemctlSockListening{}, t)
	testCheck([][]string{}, fileParameters, SystemctlSockListening{}, t)
}

func TestSystemctlTimer(t *testing.T) {
	t.Parallel()
	testParameters(names, notLengthOne, SystemctlTimer{}, t)
	testCheck([][]string{}, names, SystemctlTimer{}, t)
}

func TestSystemctlTimerLoaded(t *testing.T) {
	t.Parallel()
	testParameters(names, notLengthOne, SystemctlTimerLoaded{}, t)
	testCheck([][]string{}, names, SystemctlTimerLoaded{}, t)
}

func TestSystemctlUnitFileStatus(t *testing.T) {
	t.Parallel()
	goodEggs := [][]string{
		{"dbus.service", "static"},
		{"polkit.service", "static"},
		{"systemd-initctl.service", "static"},
	}
	validInputs := appendParameter(names, "static")
	testParameters(validInputs, notLengthTwo, SystemctlUnitFileStatus{}, t)
	testCheck(goodEggs, validInputs, SystemctlUnitFileStatus{}, t)
}
