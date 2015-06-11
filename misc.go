package main

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// php -r 'echo get_cfg_var("default_mimetype");

// command runs a shell command, and collapses its error code to 0 or 1.
// It outputs stderr and stdout if the command has error code != 0.
func command(parameters []string) (exitCode int, exitMessage string) {
	toExec := parameters[0]
	params := strings.Split(toExec, " ")
	out, err := exec.Command(params[0], params[1:]...).CombinedOutput()
	if strings.Contains(err.Error(), "not found in $PATH") {
		return 1, "Executable not found: " + params[0]
	}
	if err == nil {
		return 0, ""
	}
	// Create output message
	exitMessage += "Command exited with non-zero exit code:"
	exitMessage += "\n\tCommand: " + toExec
	exitMessage += "\n\tExit code: " + fmt.Sprint(exitCode)
	exitMessage += "\n\tExit code: " + fmt.Sprint(exitCode)
	exitMessage += "\n\tOutput: " + fmt.Sprint(out)
	return 1, exitMessage
}

// Running checks if a process is running using `ps aux`, and searching for the
// process name, excluding this process (in case the process name is in the JSON
// file name)
func Running(parameters []string) (exitCode int, exitMessage string) {
	proc := parameters[0]
	// getRunningCommands returns the entries in the "COMMAND" column of `ps aux`
	getRunningCommands := func() (commands []string) {
		cmd := exec.Command("ps", "aux")
		return commandColumnNoHeader(10, cmd)
	}
	// remove this process from consideration
	commands := getRunningCommands()
	var filtered []string
	for _, cmd := range commands {
		if !strings.Contains(cmd, "distributive") {
			filtered = append(filtered, cmd)
		}
	}
	if strIn(proc, filtered) {
		return 0, ""
	}
	return 1, "Process not running: " + proc
}

// Temp parses the output of lm_sensors and determines if Core 0 (all cores) are
// over a certain threshold as specified in the JSON.
func Temp(parameters []string) (exitCode int, exitMessage string) {
	// parse string parameters from JSON
	max := parseMyInt(parameters[0])
	// getCoreTemp returns an integer temperature for a certain core
	getCoreTemp := func(core int) (temp int) {
		out, err := exec.Command("sensors").Output()
		if err != nil {
			log.Fatal("Error while executing `sensors`:\n\t" + err.Error())
		}
		// get all-core line up to paren
		lineRegex := regexp.MustCompile("Core " + fmt.Sprint(core) + ":?(.*)\\(")
		line := lineRegex.Find(out)
		// get temp from that line
		tempRegex := regexp.MustCompile("\\d+\\.\\d*")
		tempString := string(tempRegex.Find(line))
		tempFloat, err := strconv.ParseFloat(tempString, 64)
		if err != nil {
			msg := "Error while parsing output from `sensors`:\n\t"
			log.Fatal(msg + err.Error())
		}
		return int(tempFloat)

	}
	temp := getCoreTemp(0)
	if temp < max {
		return 0, ""
	}
	msg := "Core temp exceeds defined maximum"
	return genericError(msg, fmt.Sprint(max), []string{fmt.Sprint(temp)})
}

// Module checks to see if a kernel module is installed
func Module(parameters []string) (exitCode int, exitMessage string) {
	name := parameters[0]
	// kernelModules returns a list of all modules that are currently loaded
	kernelModules := func() (modules []string) {
		cmd := exec.Command("/sbin/lsmod")
		return commandColumnNoHeader(0, cmd)
	}
	modules := kernelModules()
	if strIn(name, modules) {
		return 0, ""
	}
	return genericError("Module is not loaded", name, modules)
}

// KernelParameter checks to see if a kernel parameter was set
func KernelParameter(parameters []string) (exitCode int, exitMessage string) {
	name := parameters[0]
	// parameterValue returns the value of a kernel parameter
	parameterSet := func(name string) bool {
		_, err := exec.Command("/sbin/sysctl", "-q", "-n", name).Output()
		// failed on incorrect module name
		if err != nil && strings.Contains(err.Error(), "255") {
			return false
		} else if err != nil {
			log.Fatal("Error while executing /sbin/systctl:\n\tError: " + err.Error())
		}
		return true
	}
	if parameterSet(name) {
		return 0, ""
	}
	return 1, "Kernel parameter not set: " + name
}
