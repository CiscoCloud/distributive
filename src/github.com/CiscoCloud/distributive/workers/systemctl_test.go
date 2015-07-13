package workers

import (
	"testing"
)

var activeServices = []parameters{
	[]string{"dev-mqueue.mount"},
	[]string{"tmp.mount"},
	[]string{"dbus.service"},
}

func TestSystemctlLoaded(t *testing.T) {
	t.Parallel()
	testInputs(t, systemctlLoaded, activeServices, names)
}

func TestSystemctlActive(t *testing.T) {
	t.Parallel()
	testInputs(t, systemctlActive, activeServices, names)
}

func TestSystemctlSockPath(t *testing.T) {
	t.Parallel()
	winners := []parameters{
		[]string{"/run/dbus/system_bus_socket"},
		[]string{"/run/systemd/journal/socket"},
		[]string{"/run/dmeventd-client"},
	}
	testInputs(t, systemctlSockPath, winners, names)
}

func TestSystemctlSockUnit(t *testing.T) {
	t.Parallel()
	winners := []parameters{
		[]string{"dbus.socket"},
		[]string{"systemd-journald.socket"},
		[]string{"dm-event.socket"},
	}
	testInputs(t, systemctlSockUnit, winners, names)
}

func TestSystemctlTimer(t *testing.T) {
	t.Parallel()
	testInputs(t, systemctlTimer, []parameters{}, names)
}

func TestSystemctlTimerLoaded(t *testing.T) {
	t.Parallel()
	testInputs(t, systemctlTimerLoaded, []parameters{}, names)
}

func TestSystemctlUnitFileStatus(t *testing.T) {
	t.Parallel()
	winners := []parameters{
		[]string{"dbus.service", "static"},
		[]string{"polkit.service", "static"},
		[]string{"systemd-initctl.service", "static"},
	}
	losers := appendParameter(names, "fakestatus")
	testInputs(t, systemctlUnitFileStatus, winners, losers)
}
