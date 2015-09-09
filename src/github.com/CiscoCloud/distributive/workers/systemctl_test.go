package workers

import (
	"testing"
)

var activeServices = [][]string{
	[]string{"dev-mqueue.mount"},
	[]string{"tmp.mount"},
	[]string{"dbus.service"},
}

func TestSystemctlLoaded(t *testing.T) {
	//t.Parallel()
	testParameters(names, notLengthOne, SystemctlLoaded{}, t)
	testCheck(activeServices, names, SystemctlLoaded{}, t)
}

func TestSystemctlActive(t *testing.T) {
	//t.Parallel()
	testParameters(names, notLengthOne, SystemctlActive{}, t)
	testCheck(activeServices, names, SystemctlLoaded{}, t)
}

func TestSystemctlSockPath(t *testing.T) {
	//t.Parallel()
	goodEggs := [][]string{
		[]string{"/run/dbus/system_bus_socket"},
		[]string{"/run/systemd/journal/socket"},
		[]string{"/run/dmeventd-client"},
	}
	invalidInputs := append(notLengthOne, names...)
	testParameters(fileParameters, invalidInputs, SystemctlSockListening{}, t)
	testCheck(goodEggs, fileParameters, SystemctlSockListening{}, t)
}

func TestSystemctlSockUnit(t *testing.T) {
	//t.Parallel()
	goodEggs := [][]string{
		[]string{"dbus.socket"},
		[]string{"systemd-journald.socket"},
		[]string{"dm-event.socket"},
	}
	testParameters(names, notLengthOne, SystemctlSockUnit{}, t)
	testCheck(goodEggs, names, SystemctlSockUnit{}, t)
}

func TestSystemctlTimer(t *testing.T) {
	//t.Parallel()
	testParameters(names, notLengthOne, SystemctlTimer{}, t)
	testCheck([][]string{}, names, SystemctlTimer{}, t)
}

func TestSystemctlTimerLoaded(t *testing.T) {
	//t.Parallel()
	testParameters(names, notLengthOne, SystemctlTimerLoaded{}, t)
	testCheck([][]string{}, names, SystemctlTimerLoaded{}, t)
}

func TestSystemctlUnitFileStatus(t *testing.T) {
	//t.Parallel()
	goodEggs := [][]string{
		[]string{"dbus.service", "static"},
		[]string{"polkit.service", "static"},
		[]string{"systemd-initctl.service", "static"},
	}
	validInputs := appendParameter(names, "static")
	testParameters(validInputs, notLengthTwo, SystemctlUnitFileStatus{}, t)
	testCheck(goodEggs, validInputs, SystemctlUnitFileStatus{}, t)
}
