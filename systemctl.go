package main

import (
	"log"
	"os/exec"
	"regexp"
	"strings"
)

// systemctlExists returns whether or not systemctl is available ona given
// machine
func systemctlExists() bool {
	_, err := exec.Command("systemctl", "--version").CombinedOutput()
	if err != nil && strings.Contains(err.Error(), "not found in $PATH") {
		return false
	}
	return true
}

// systemctlShouldExist logs and quits if it doesn't.
func systemctlShouldExist() {
	if !systemctlExists() {
		log.Fatal("Couldn't execute systemctl")
	}
}

// systemctlServices checks on either the loaded or active field of
// `systemctl list-units`. It is an abstraction of systemctlLoaded and
// systemctlActive.
func systemctlService(service string, loaded bool) (exitCode int, exitMessage string) {
	systemctlShouldExist() // error out if the command doesn't work
	column := 2            // active, not loaded
	state := "active"
	if loaded { // loaded, not active
		column = 1
		state = "loaded"
	}
	// get columns
	cmd := exec.Command("systemctl", "--no-pager", "list-units")
	names := commandColumnNoHeader(1, cmd)
	// can't execute the same command twice
	cmd = exec.Command("systemctl", "--no-pager", "list-units")
	statuses := commandColumnNoHeader(column+1, cmd) // weird offset
	// parse through columns
	var actualState string
	for i, srv := range names {
		if service == srv && len(statuses) > i {
			actualState = statuses[i]
			if actualState == state {
				return 0, ""
			}
			msg := "Service did not have state"
			return genericError(msg, state, []string{actualState})
		}
	}
	return 1, "You shouldn't be seeing this message. File a bug report please."
}

// systemctlLoaded checks to see whether or not a given service is loaded
func systemctlLoaded(parameters []string) (exitCode int, exitMessage string) {
	return systemctlService(parameters[0], true)
}

// systemctlActive checks to see whether or not a given service is active
func systemctlActive(parameters []string) (exitCode int, exitMessage string) {
	return systemctlService(parameters[0], false)
}

// systemctlSock is an abstraction of systemctlSockPath and systemctlSockUnit,
// it reads from `systemctl list-sockets` and sees if the value is in the
// appropriate column.
func systemctlSock(value string, path bool) (exitCode int, exitMessage string) {
	systemctlShouldExist() // log.Fatal if it doesn't
	column := 1
	if path {
		column = 0
	}
	cmd := exec.Command("systemctl", "list-sockets")
	values := commandColumnNoHeader(column, cmd)
	if strIn(value, values) {
		return 0, ""
	}
	return genericError("Socket not found", value, values)
}

// systemctlSock checks to see whether the sock at the given path is registered
// within systemd using the sock's filesystem path.
func systemctlSockPath(parameters []string) (exitCode int, exitMessage string) {
	return systemctlSock(parameters[0], true)
}

// systemctlSock checks to see whether the sock at the given path is registered
// within systemd using the sock's unit name.
func systemctlSockUnit(parameters []string) (exitCode int, exitMessage string) {
	return systemctlSock(parameters[0], false)
}

func getTimers(all bool) []string {
	cmd := exec.Command("systemctl", "list-timers")
	if all {
		cmd = exec.Command("systemctl", "list-timers", "--all")
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := "Couldn't execute `systemctl list-timers`:\n\t" + err.Error()
		log.Fatal(msg)
	}
	// matches anything with hyphens or letters, then a ".timer"
	re := regexp.MustCompile("(\\w|\\-)+\\.timer")
	return re.FindAllString(string(out), -1)
}

// timers(exitCode int, exitMessage string) is pure DRY for systemctlTimer and systemctlTimerLoaded
func timersWorker(unit string, all bool) (exitCode int, exitMessage string) {
	timers := getTimers(all)
	if strIn(unit, timers) {
		return 0, ""
	}
	return genericError("Timer not found", unit, timers)
}

// systemctlTimer reports whether a given timer is running (by unit).
func systemctlTimer(parameters []string) (exitCode int, exitMessage string) {
	return timersWorker(parameters[0], false)
}

// systemctlTimerLoaded checks to see if a timer is loaded, even if it might
// not be active
func systemctlTimerLoaded(parameters []string) (exitCode int, exitMessage string) {
	return timersWorker(parameters[0], true)
}

// systemctlUnitFileStatus checks whether or not the given unit file has the
// given status: static | enabled | disabled
func systemctlUnitFileStatus(parameters []string) (exitCode int, exitMessage string) {
	// getUnitFilesWithStatuses returns a pair of string slices that hold
	// the name of unit files with their current statuses.
	getUnitFilesWithStatuses := func() (units []string, statuses []string) {
		cmd := exec.Command("systemctl", "--no-pager", "list-unit-files")
		units = commandColumnNoHeader(0, cmd)
		cmd = exec.Command("systemctl", "--no-pager", "list-unit-files")
		statuses = commandColumnNoHeader(1, cmd)
		// last two are empty line and junk statistics we don't care about
		return units[:len(units)-2], statuses[:len(statuses)-2]
	}
	unit := parameters[0]
	status := parameters[1]
	units, statuses := getUnitFilesWithStatuses()
	var actualStatus string
	for i, un := range units {
		if un == unit {
			actualStatus = statuses[i]
			if actualStatus == status {
				return 0, ""
			}
		}
	}
	msg := "Unit didn't have status"
	return genericError(msg, status, []string{actualStatus})
}
