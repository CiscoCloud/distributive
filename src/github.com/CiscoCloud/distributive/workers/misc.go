package workers

import (
	"fmt"
	"github.com/CiscoCloud/distributive/tabular"
	"github.com/CiscoCloud/distributive/wrkutils"
	log "github.com/Sirupsen/logrus"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

// command runs a shell command, and collapses its error code to 0 or 1.
// It outputs stderr and stdout if the command has error code != 0.
func command(parameters []string) (exitCode int, exitMessage string) {
	toExec := parameters[0]
	cmd := exec.Command("bash", "-c", toExec)
	err := cmd.Start()
	if err != nil && strings.Contains(err.Error(), "not found in $PATH") {
		return 1, "Executable not found: " + toExec
	} else if err != nil {
		wrkutils.ExecError(cmd, "", err)
	}
	if err = cmd.Wait(); err != nil {
		// this is convoluted, but should work on Windows & Unix
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			}
		}
		// dummy, in case the above failed. We know it's not zero!
		if exitCode == 0 {
			exitCode = 1
		}
		out, _ := cmd.CombinedOutput() // don't care if this fails
		exitMessage += "Command exited with non-zero exit code:"
		exitMessage += "\n\tCommand: " + toExec
		exitMessage += "\n\tExit code: " + fmt.Sprint(exitCode)
		exitMessage += "\n\tOutput: " + string(out)
		return 1, exitMessage
	}
	return 0, ""
}

// commandOutputMatches checks to see if a command's combined output matches a
// given regexp
func commandOutputMatches(parameters []string) (exitCode int, exitMessage string) {
	toExec := parameters[0]
	re := wrkutils.ParseUserRegex(parameters[1])
	cmd := exec.Command("bash", "-c", toExec)
	out, err := cmd.CombinedOutput()
	if err != nil {
		wrkutils.ExecError(cmd, string(out), err)
	}
	if re.Match(out) {
		return 0, ""
	}
	msg := "Command output did not match regexp"
	return wrkutils.GenericError(msg, re.String(), []string{string(out)})
}

// running checks if a process is running using `ps aux`, and searching for the
// process name, excluding this process (in case the process name is in the JSON
// file name)
func running(parameters []string) (exitCode int, exitMessage string) {
	// getRunningCommands returns the entries in the "COMMAND" column of `ps aux`
	getRunningCommands := func() (commands []string) {
		cmd := exec.Command("ps", "aux")
		return wrkutils.CommandColumnNoHeader(10, cmd)
	}
	proc := parameters[0]
	// remove this process from consideration
	commands := getRunningCommands()
	var filtered []string
	for _, cmd := range commands {
		if !strings.Contains(cmd, "distributive") {
			filtered = append(filtered, cmd)
		}
	}
	if tabular.StrIn(proc, filtered) {
		return 0, ""
	}
	return wrkutils.GenericError("Process not running", proc, filtered)
}

// temp parses the output of lm_sensors and determines if Core 0 (all cores) are
// over a certain threshold as specified in the JSON.
func temp(parameters []string) (exitCode int, exitMessage string) {
	// TODO: check for negative, outrageously high temperatures
	// allCoreTemps returns the temperature of each core
	allCoreTemps := func() (temps []int) {
		cmd := exec.Command("sensors")
		out, err := cmd.CombinedOutput()
		outstr := string(out)
		wrkutils.ExecError(cmd, outstr, err)
		restr := "Core\\s\\d+:\\s+[\\+\\-](?P<temp>\\d+)\\.*\\d*°C"
		re := regexp.MustCompile(restr)
		for _, line := range regexp.MustCompile("\n+").Split(outstr, -1) {
			if re.MatchString(line) {
				// submatch captures only the integer part of the temperature
				matchDict := wrkutils.SubmatchMap(re, line)
				if _, ok := matchDict["temp"]; !ok {
					log.WithFields(log.Fields{
						"regexp":    re.String(),
						"matchDict": matchDict,
						"output":    outstr,
					}).Fatal("Couldn't find any temperatures in `sensors` output")
				}
				tempInt64, err := strconv.ParseInt(matchDict["temp"], 10, 64)
				if err != nil {
					log.WithFields(log.Fields{
						"regexp":    re.String(),
						"matchDict": matchDict,
						"output":    outstr,
						"error":     err.Error(),
					}).Fatal("Couldn't parse integer from `sensors` output")
				}
				temps = append(temps, int(tempInt64))
			}
		}
		return temps
	}
	// getCoreTemp returns an integer temperature for a certain core
	getCoreTemp := func(core int) (temp int) {
		temps := allCoreTemps()
		wrkutils.IndexError("No such core available", core, temps)
		return temps[core]
	}
	max := wrkutils.ParseMyInt(parameters[0])
	temp := getCoreTemp(0)
	if temp < max {
		return 0, ""
	}
	msg := "Core temp exceeds defined maximum"
	return wrkutils.GenericError(msg, max, []string{fmt.Sprint(temp)})
}

// module checks to see if a kernel module is installed
func module(parameters []string) (exitCode int, exitMessage string) {
	// kernelModules returns a list of all modules that are currently loaded
	// TODO just read from /proc/modules
	kernelModules := func() (modules []string) {
		cmd := exec.Command("/sbin/lsmod")
		return wrkutils.CommandColumnNoHeader(0, cmd)
	}
	name := parameters[0]
	modules := kernelModules()
	if tabular.StrIn(name, modules) {
		return 0, ""
	}
	return wrkutils.GenericError("Module is not loaded", name, modules)
}

// kernelParameter checks to see if a kernel parameter was set
func kernelParameter(parameters []string) (exitCode int, exitMessage string) {
	// parameterValue returns the value of a kernel parameter
	parameterSet := func(name string) bool {
		cmd := exec.Command("/sbin/sysctl", "-q", "-n", name)
		out, err := cmd.CombinedOutput()
		// failed on incorrect module name
		if err != nil && strings.Contains(err.Error(), "255") {
			return false
		} else if err != nil {
			wrkutils.ExecError(cmd, string(out), err)
		}
		return true
	}
	name := parameters[0]
	if parameterSet(name) {
		return 0, ""
	}
	return 1, "Kernel parameter not set: " + name
}

// phpConfig checks the value of a PHP configuration variable
func phpConfig(parameters []string) (exitCode int, exitMessage string) {
	// getPHPVariable returns the value of a PHP configuration value as a string
	// or just "" if it doesn't exist
	getPHPVariable := func(name string) (val string) {
		quote := func(str string) string {
			return "\"" + str + "\""
		}
		// php -r 'echo get_cfg_var("default_mimetype");
		echo := fmt.Sprintf("echo get_cfg_var(%s);", quote(name))
		cmd := exec.Command("php", "-r", echo)
		out, err := cmd.CombinedOutput()
		if err != nil {
			wrkutils.ExecError(cmd, string(out), err)
		}
		return string(out)
	}
	name := parameters[0]
	value := parameters[1]
	actualValue := getPHPVariable(name)
	if actualValue == value {
		return 0, ""
	} else if actualValue == "" {
		msg := "PHP configuration variable not set"
		return wrkutils.GenericError(msg, value, []string{actualValue})
	}
	msg := "PHP variable did not match expected value"
	return wrkutils.GenericError(msg, value, []string{actualValue})
}
