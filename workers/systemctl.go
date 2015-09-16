package workers

import (
	"fmt"
	"github.com/CiscoCloud/distributive/chkutil"
	"github.com/CiscoCloud/distributive/errutil"
	"github.com/CiscoCloud/distributive/tabular"
	"os"
	"os/exec"
	"strings"
)

// systemctlService checks to see if a service has a givens status
// status: active | loaded
func systemctlService(service string, activeOrLoaded string) (int, string, error) {
	// cmd depends on whether we're checking active or loaded
	cmd := exec.Command("systemctl", "show", "-p", "ActiveState", service)
	if activeOrLoaded == "loaded" {
		cmd = exec.Command("systemctl", "show", "-p", "LoadState", service)
	}
	outString := chkutil.CommandOutput(cmd)
	contained := "ActiveState=active"
	if activeOrLoaded == "loaded" {
		contained = "LoadState=loaded"
	}
	if strings.Contains(outString, contained) {
		return errutil.Success()
	}
	msg := "Service not " + activeOrLoaded
	return errutil.GenericError(msg, service, []string{outString})
}

/*
#### SystemctlLoaded
Description: Is systemd module loaded?
Parameters:
  - Service (string): Name of the service
Example parameters:
  - TODO
*/

type SystemctlLoaded struct{ service string }

func (chk SystemctlLoaded) ID() string { return "SystemctlLoaded" }

func (chk SystemctlLoaded) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	chk.service = params[0]
	return chk, nil
}

func (chk SystemctlLoaded) Status() (int, string, error) {
	return systemctlService(chk.service, "loaded")
}

/*
#### SystemctlActive
Description: Is systemd module active?
Parameters:
  - Service (string): Name of the service
Example parameters:
  - TODO
*/

type SystemctlActive struct{ service string }

func (chk SystemctlActive) ID() string { return "SystemctlActive" }

func (chk SystemctlActive) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	chk.service = params[0]
	return chk, nil
}

func (chk SystemctlActive) Status() (int, string, error) {
	return systemctlService(chk.service, "active")
}

// SystemctlSock is an abstraction of SystemctlSockPath and SystemctlSockUnit,
// it reads from `systemctl list-sockets` and sees if the value is in the
// appropriate column.
func SystemctlSock(value string, column string) (int, string, error) {
	outstr := chkutil.CommandOutput(exec.Command("systemctl", "list-sockets"))
	lines := tabular.Lines(outstr)
	msg := "systemctl list-sockers didn't output enough rows"
	errutil.IndexError(msg, len(lines)-4, lines)
	unlines := tabular.Unlines(lines[:len(lines)-4])
	table := tabular.SeparateOnAlignment(unlines)
	values := tabular.GetColumnByHeader(column, table)
	if tabular.StrIn(value, values) {
		return errutil.Success()
	}
	return errutil.GenericError("Socket not found", value, values)
}

/*
#### SystemctlSockListening
Description: Is systemd socket in the LISTEN state?
Parameters:
  - Path (filepath): Path to socket
Example parameters:
  - /var/lib/docker.sock, /new/striped.sock
*/

type SystemctlSockListening struct{ path string }

func (chk SystemctlSockListening) ID() string { return "SystemctlSock" }

func (chk SystemctlSockListening) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	} else if _, err := os.Stat(params[0]); err != nil {
		return chk, errutil.ParameterTypeError{params[0], "filepath"}
	}
	chk.path = params[0]
	return chk, nil
}

func (chk SystemctlSockListening) Status() (int, string, error) {
	return SystemctlSock(chk.path, "LISTEN")
}

/*
#### SystemctlSockUnit
Description: Is a socket registered with this unit?
Parameters:
  - Unit (string): Name of systemd unit
Example parameters:
  - TODO
*/

type SystemctlSockUnit struct{ path, unit string }

func (chk SystemctlSockUnit) ID() string { return "SystemctlSockUnit" }

func (chk SystemctlSockUnit) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	chk.unit = params[0]
	return chk, nil
}

func (chk SystemctlSockUnit) Status() (int, string, error) {
	return SystemctlSock(chk.unit, "UNIT")
}

// getTimers returns of all the timers under the UNIT column of
// `systemctl list-timers`
func getTimers(all bool) []string {
	cmd := exec.Command("systemctl", "list-timers")
	if all {
		cmd = exec.Command("systemctl", "list-timers", "--all")
	}
	out, err := cmd.CombinedOutput()
	outstr := string(out)
	errutil.ExecError(cmd, outstr, err)
	// last three lines are junk
	lines := tabular.Lines(outstr)
	msg := fmt.Sprint(cmd.Args) + " didn't output enough lines"
	errutil.IndexError(msg, 3, lines)
	table := tabular.SeparateOnAlignment(tabular.Unlines(lines[:len(lines)-3]))
	column := tabular.GetColumnByHeader("UNIT", table)
	return column
}

// timerCheck is pure DRY for SystemctlTimer and SystemctlTimerLoaded
func timerCheck(unit string, all bool) (int, string, error) {
	timers := getTimers(all)
	if tabular.StrIn(unit, timers) {
		return errutil.Success()
	}
	return errutil.GenericError("Timer not found", unit, timers)
}

/*
#### SystemctlTimer
Description: Is a timer by this name running?
Parameters:
  - Unit (string): Name of systemd unit
Example parameters:
  - TODO
*/

type SystemctlTimer struct{ unit string }

func (chk SystemctlTimer) ID() string { return "SystemctlTimer" }

func (chk SystemctlTimer) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	chk.unit = params[0]
	return chk, nil
}

func (chk SystemctlTimer) Status() (int, string, error) {
	return timerCheck(chk.unit, false)
}

/*
#### SystemctlTimerLoaded
Description: Is a timer by this name loaded?
Parameters:
  - Unit (string): Name of systemd unit
Example parameters:
  - TODO
*/

type SystemctlTimerLoaded struct{ unit string }

func (chk SystemctlTimerLoaded) ID() string { return "SystemctlTimerLoaded" }

func (chk SystemctlTimerLoaded) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	chk.unit = params[0]
	return chk, nil
}

func (chk SystemctlTimerLoaded) Status() (int, string, error) {
	return timerCheck(chk.unit, true)
}

/*
#### SystemctlUnitFileStatus
Description: Does this unit file have this status?
Parameters:
  - Unit (string): Name of systemd unit
  - Status (string): "static" | "enabled" | "disabled"
Example parameters:
  - TODO
  - "static", "enabled", "disabled"
*/

type SystemctlUnitFileStatus struct{ unit, status string }

func (chk SystemctlUnitFileStatus) ID() string { return "SystemctlUnitFileStatus" }

func (chk SystemctlUnitFileStatus) New(params []string) (chkutil.Check, error) {
	if len(params) != 2 {
		return chk, errutil.ParameterLengthError{2, params}
	}
	chk.unit = params[0]
	chk.status = params[1]
	return chk, nil
}

func (chk SystemctlUnitFileStatus) Status() (int, string, error) {
	// getUnitFilesWithStatuses returns a pair of string slices that hold
	// the name of unit files with their current statuses.
	getUnitFilesWithStatuses := func() (units []string, statuses []string) {
		cmd := exec.Command("systemctl", "--no-pager", "list-unit-files")
		units = chkutil.CommandColumnNoHeader(0, cmd)
		cmd = exec.Command("systemctl", "--no-pager", "list-unit-files")
		statuses = chkutil.CommandColumnNoHeader(1, cmd)
		// last two are empty line and junk statistics we don't care about
		msg := fmt.Sprint(cmd.Args) + " didn't output enough lines"
		errutil.IndexError(msg, 2, units)
		errutil.IndexError(msg, 2, statuses)
		return units[:len(units)-2], statuses[:len(statuses)-2]
	}
	units, statuses := getUnitFilesWithStatuses()
	var actualStatus string
	// TODO check if unit could be found at all
	for i, un := range units {
		if un == chk.unit {
			actualStatus = statuses[i]
			if actualStatus == chk.status {
				return errutil.Success()
			}
		}
	}
	msg := "Unit didn't have status"
	return errutil.GenericError(msg, chk.status, []string{actualStatus})
}
