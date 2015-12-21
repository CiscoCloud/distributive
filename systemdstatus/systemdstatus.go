// fsstatus provides utility functions for querying several aspects of systemd's
// status, especially as pertains to monitoring.
package systemdstatus

import (
	"errors"
	"fmt"
	"github.com/CiscoCloud/distributive/tabular"
	"os/exec"
	"strings"
)

// ServiceLoaded returns whether or not the given systemd service has
// LoadState=loaded
func ServiceLoaded(name string) (bool, error) {
	cmd := exec.Command("systemctl", "show", "-p", "LoadState", name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, errors.New(err.Error() + ": output: " + string(out))
	}
	return strings.Contains(string(out), "LoadState=loaded"), nil
}

// ServiceLoaded returns whether or not the given systemd service has
// ActiveState=active
func ServiceActive(name string) (bool, error) {
	cmd := exec.Command("systemctl", "show", "-p", "ActiveState", name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, errors.New(err.Error() + ": output: " + string(out))
	}
	return strings.Contains(string(out), "ActiveState=active"), nil
}

// ListeningSockets returns a list of all sockets in the "LISTENING" state
func ListeningSockets() (socks []string, err error) {
	out, err := exec.Command("systemctl", "list-sockets").CombinedOutput()
	if err != nil {
		return socks, errors.New(err.Error() + ": output: " + string(out))
	}
	table := tabular.ProbabalisticSplit(string(out))
	return tabular.GetColumnByHeader("LISTENING", table), nil
}

// Timers returns a list of the active systemd timers, as found under the
// UNIT column of `systemctl list-timers`. It can optionally list all timers.
func Timers(all bool) (timers []string, err error) {
	cmd := exec.Command("systemctl", "list-timers")
	if all {
		cmd = exec.Command("systemctl", "list-timers", "--all")
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return timers, errors.New(err.Error() + ": output: " + string(out))
	}
	// last three lines are junk
	lines := tabular.Lines(string(out))
	if len(lines) <= 3 {
		msg := fmt.Sprint(cmd.Args) + " didn't output enough lines"
		return timers, errors.New(msg)
	}
	table := tabular.SeparateOnAlignment(tabular.Unlines(lines[:len(lines)-3]))
	column := tabular.GetColumnByHeader("UNIT", table)
	return column, nil
}

// UnitFileStatuses returns a list of all unit files with their current status,
// as shown by `systemctl list-unit-files`.
func UnitFileStatuses() (units, statuses []string, err error) {
	cmd := exec.Command("systemctl", "--no-pager", "list-unit-files")
	out, err := cmd.CombinedOutput()
	if err != nil {
		err2 := errors.New(err.Error() + ": output: " + string(out))
		return units, statuses, err2
	}
	table := tabular.ProbabalisticSplit(string(out))
	units = tabular.GetColumnNoHeader(0, table)
	// last two are empty line and junk statistics we don't care about
	if len(units) <= 3 {
		msg := fmt.Sprint(cmd.Args) + " didn't output enough lines"
		return units, statuses, errors.New(msg)
	}
	cmd = exec.Command("systemctl", "--no-pager", "list-unit-files")
	table = tabular.ProbabalisticSplit(string(out))
	statuses = tabular.GetColumnNoHeader(1, table)
	// last two are empty line and junk statistics we don't care about
	if len(statuses) <= 3 {
		msg := fmt.Sprint(cmd.Args) + " didn't output enough lines"
		return units, statuses, errors.New(msg)
	}
	return units[:len(units)-2], statuses[:len(statuses)-2], nil

}
