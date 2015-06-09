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

// Command runs a shell command, and collapses its error code to 0 or 1.
// It outputs stderr and stdout if the command has error code != 0.
func Command(toExec string) Thunk {
	return func() (exitCode int, exitMessage string) {
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
}

// Running checks if a process is running using `ps aux`, and searching for the
// process name, excluding this process (in case the process name is in the JSON
// file name)
func Running(proc string) Thunk {
	// getRunningCommands returns the entries in the "COMMAND" column of `ps aux`
	getRunningCommands := func() (commands []string) {
		cmd := exec.Command("ps", "aux")
		return commandColumnNoHeader(10, cmd)
	}
	return func() (exitCode int, exitMessage string) {
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
}

// Temp parses the output of lm_sensors and determines if Core 0 (all cores) are
// over a certain threshold as specified in the JSON.
func Temp(max int) Thunk {
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
	return func() (exitCode int, exitMessage string) {
		temp := getCoreTemp(0)
		if temp < max {
			return 0, ""
		}
		msg := "Core temp exceeds defined maximum"
		return notInError(msg, fmt.Sprint(temp), []string{fmt.Sprint(max)})
	}
}

// Module checks to see if a kernel module is installed
func Module(name string) Thunk {
	// kernelModules returns a list of all modules that are currently loaded
	kernelModules := func() (modules []string) {
		cmd := exec.Command("/sbin/lsmod")
		return commandColumnNoHeader(0, cmd)
	}
	return func() (exitCode int, exitMessage string) {
		modules := kernelModules()
		if strIn(name, modules) {
			return 0, ""
		}
		return notInError("Module is not loaded", name, modules)
	}
}

// KernelParameter checks to see if a kernel parameter was set
func KernelParameter(name string) Thunk {
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
	return func() (exitCode int, exitMessage string) {
		if parameterSet(name) {
			return 0, ""
		}
		return 1, "Kernel parameter not set: " + name
	}
}
