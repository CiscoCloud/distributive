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
		exitCode = 0
		if err != nil {
			exitCode = 1
		}
		// Create output message
		exitMessage = ""
		if exitCode != 0 {
			exitMessage = "Command " + toExec + " executed "
			exitMessage += "with exit code " + fmt.Sprint(exitCode)
			exitMessage += "\n\n"
			exitMessage += "output: \n" + fmt.Sprint(out)
		}
		return exitCode, exitMessage
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

// Installed detects whether the OS is using dpkg, rpm, or pacman, queries
// a package accoringly, and returns an error if it is not installed.
func Installed(pkg string) Thunk {
	// getManager returns the program to use for the query
	getManager := func(managers []string) string {
		for _, program := range managers {
			cmd := exec.Command(program, "--version")
			err := cmd.Start()
			// as long as the command was found, return that manager
			message := ""
			if err != nil {
				message = err.Error()
			}
			if strings.Contains(message, "not found") == false {
				return program
			}
		}
		log.Fatal("No package manager found")
		return "No package manager found"
	}
	// getQuery returns the command that should be used to query the pkg
	getQuery := func(program string) (name string, options string) {
		switch program {
		case "dpkg":
			return "dpkg", "-s"
		case "rpm":
			return "rpm", "-q"
		case "pacman":
			return "pacman", "-Qs"
		default:
			log.Fatal("Unsupported package manager")
			return "echo " + program + " is not supported. ", ""
		}
	}

	managers := []string{"dpkg", "rpm", "pacman"}

	return func() (exitCode int, exitMessage string) {
		name, options := getQuery(getManager(managers))
		out, _ := exec.Command(name, options, pkg).Output()
		if strings.Contains(string(out), pkg) == false {
			return 1, "Package " + pkg + " was not found with " + name + "\n"
		}
		return 0, ""
	}
}

// Temp parses the output of lm_sensors and determines if Core 0 (all cores) are
// over a certain threshold as specified in the JSON.
func Temp(max int) Thunk {
	// getCoreTemp returns an integer temperature for a certain core
	getCoreTemp := func(core int) (temp int) {
		out, err := exec.Command("sensors").Output()
		fatal(err)
		// get all-core line up to paren
		lineRegex := regexp.MustCompile("Core " + fmt.Sprint(core) + ":?(.*)\\(")
		line := lineRegex.Find(out)
		// get temp from that line
		tempRegex := regexp.MustCompile("\\d+\\.\\d*")
		tempString := string(tempRegex.Find(line))
		tempFloat, err := strconv.ParseFloat(tempString, 64)
		fatal(err)
		return int(tempFloat)

	}
	return func() (exitCode int, exitMessage string) {
		temp := getCoreTemp(0)
		if temp < max {
			return 0, ""
		}
		return 1, "Core temp exceeds defined maximum: " + fmt.Sprint(temp)
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
		if strIn(name, kernelModules()) {
			return 0, ""
		}
		return 1, "Module is not loaded: " + name
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
		}
		fatal(err)
		return true
	}
	return func() (exitCode int, exitMessage string) {
		if parameterSet(name) {
			return 0, ""
		}
		return 1, "Kernel parameter not set: " + name
	}
}
