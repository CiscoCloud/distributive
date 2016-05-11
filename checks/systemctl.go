package checks

import (
	"github.com/CiscoCloud/distributive/chkutil"
	"github.com/CiscoCloud/distributive/errutil"
	"github.com/CiscoCloud/distributive/systemdstatus"
	"github.com/CiscoCloud/distributive/tabular"
	"os"
	"strings"
)

/*
#### SystemctlLoaded
Description: Is systemd module loaded?
Parameters:
  - Service (string): Name of the service
Example parameters:
  - TODO
*/

type SystemctlLoaded struct{ service string }

func init() {
    chkutil.Register("SystemctlLoaded", func() chkutil.Check {
        return &SystemctlLoaded{}
    })
    chkutil.Register("SystemctlActive", func() chkutil.Check {
        return &SystemctlActive{}
    })
    chkutil.Register("SystemctlSock", func() chkutil.Check {
        return &SystemctlSockListening{}
    })
    chkutil.Register("SystemctlTimerLoaded", func() chkutil.Check {
        return &SystemctlTimerLoaded{}
    })
    chkutil.Register("SystemctlUnitFileStatus", func() chkutil.Check {
        return &SystemctlUnitFileStatus{}
    })
}

func (chk SystemctlLoaded) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	chk.service = params[0]
	return chk, nil
}

func (chk SystemctlLoaded) Status() (int, string, error) {
	boo, err := systemdstatus.ServiceLoaded(chk.service)
	if err != nil {
		return 1, "", err
	} else if boo {
		return errutil.Success()
	}
	return 1, "Service wasn't loaded: " + chk.service, nil
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

func (chk SystemctlActive) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	chk.service = params[0]
	return chk, nil
}

func (chk SystemctlActive) Status() (int, string, error) {
	boo, err := systemdstatus.ServiceActive(chk.service)
	if err != nil {
		return 1, "", err
	} else if boo {
		return errutil.Success()
	}
	return 1, "Service wasn't active: " + chk.service, nil
}

/*
#### SystemctlSockListening
Description: Is the systemd socket at this path in the LISTEN state?
Parameters:
  - Path (filepath): Path to socket
Example parameters:
  - /var/lib/docker.sock, /new/striped.sock
*/

type SystemctlSockListening struct{ path string }

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
	listening, err := systemdstatus.ListeningSockets()
	if err != nil {
		return 1, "", err
	}
	if tabular.StrIn(chk.path, listening) {
		return errutil.Success()
	}
	return errutil.GenericError("Socket wasn't listening", chk.path, listening)
}

// timerCheck is pure DRY for SystemctlTimer and SystemctlTimerLoaded
func timerCheck(unit string, all bool) (int, string, error) {
	timers, err := systemdstatus.Timers(all)
	if err != nil {
		return 1, "", err
	} else if tabular.StrIn(unit, timers) {
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

func (chk SystemctlUnitFileStatus) New(params []string) (chkutil.Check, error) {
	if len(params) != 2 {
		return chk, errutil.ParameterLengthError{2, params}
	}
	validStatuses := []string{"static", "enabled", "disabled"}
	if !tabular.StrIn(strings.ToLower(params[1]), validStatuses) {
		validStatusesStr := "static | enabled | disabled"
		return chk, errutil.ParameterTypeError{params[1], validStatusesStr}
	}
	chk.unit = params[0]
	chk.status = params[1]
	return chk, nil
}

func (chk SystemctlUnitFileStatus) Status() (int, string, error) {
	units, statuses, err := systemdstatus.UnitFileStatuses()
	if err != nil {
		return 1, "", err
	}
	var actualStatus string
	found := false
	for _, unit := range units {
		if unit == chk.unit {
			found = true
		}
	}
	if !found {
		return 1, "Unit file could not be found: " + chk.unit, nil
	}
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
