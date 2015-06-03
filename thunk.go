package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// TODO: eliminate thunks that are simple if/else statements by abstracting that
// interaction: create a new type called boolThunk, and pass it to a method

// Thunk is the type of function that runs without parameters and returns
// an error code and an exit message to be printed to stdout.
// Generally, if exitCode == 0, exitMessage == "".
type Thunk func() (exitCode int, exitMessage string)

// Command runs a shell command, and collapses its error code to 0 or 1.
// It outputs stderr and stdout if the command has error code != 0.
func Command(toExec string) Thunk {
	return func() (exitCode int, exitMessage string) {
		params := strings.Split(toExec, " ")
		cmd := exec.Command(params[0], params[1:]...)
		// capture outputs
		stdout, err := cmd.StdoutPipe()
		fatal(err)
		stderr, err := cmd.StderrPipe()
		fatal(err)
		// run the command
		err = cmd.Start()
		if strings.Contains(err.Error(), "not found in $PATH") {
			return 1, "Executable not found: " + params[0]
		}
		err = cmd.Wait()
		exitCode = 0
		if err != nil {
			exitCode = 1
		}
		stdoutBytes, err := ioutil.ReadAll(stdout)
		fatal(err)
		stderrBytes, err := ioutil.ReadAll(stderr)
		// Create output message
		exitMessage = ""
		if exitCode != 0 {
			exitMessage = "Command " + toExec + " executed "
			exitMessage += "with exit code " + fmt.Sprint(exitCode)
			exitMessage += "\n\n"
			exitMessage += "stdout: \n" + fmt.Sprint(stdoutBytes)
			exitMessage += "\n\n"
			exitMessage += "stderr: \n" + fmt.Sprint(stderrBytes)
		}
		return exitCode, exitMessage
	}
}

// Running checks if a process is running using `ps aux`, and searching for the
// process name, excluding this process (in case the process name is in the JSON
// file name)
func Running(proc string) Thunk {
	return func() (exitCode int, exitMessage string) {
		cmd := exec.Command("ps", "aux")
		stdoutBytes, err := cmd.Output()
		fatal(err)
		// this regex matches: flag, space, quote, path, filename.json, quote
		re := regexp.MustCompile("-f\\s+\"*?.*?(health-checks/)*?[^/]*.json\"*")
		// remove this process from consideration
		filtered := re.ReplaceAllString(string(stdoutBytes), "")
		if strings.Contains(filtered, proc) {
			return 0, ""
		} else {
			return 1, "Process not running: " + proc
		}
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
	kernelModules := func() (modules [][]byte) {
		out, err := exec.Command("/sbin/lsmod").Output()
		fatal(err)
		for _, line := range bytes.Split(out, []byte("\n"))[1:] {
			module := bytes.Split(line, []byte(" "))[0]
			modules = append(modules, module)
		}
		return modules
	}
	// isLoaded returns whether or not a kernel module is currently loaded
	isLoaded := func(name string) bool {
		for _, module := range kernelModules() {
			if string(module) == name {
				return true
			}
		}
		return false
	}
	return func() (exitCode int, exitMessage string) {
		if isLoaded(name) {
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
