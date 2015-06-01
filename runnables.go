// Thunks provides functions that construct functions in the format that
// Checkr expects, namely the Thunk type, that can be used as health checks.
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// Thunk is the type of function that runs without parameters and returns
// an error code and an exit message to be printed to stdout.
// Generally, if exitCode == 0, exitMessage == "".
type Thunk func() (exitCode int, exitMessage string)

// Command runs a shell command, and collapses its error code into the range [0-1]
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
		fatal(err)
		err = cmd.Wait()
		exitCode = 0
		if err != nil {
			exitCode = 1
		}
		stdoutText, err := ioutil.ReadAll(stdout)
		stderrText, err := ioutil.ReadAll(stderr)
		// Create output message
		exitMessage = ""
		if exitCode != 0 {
			exitMessage = "Command " + toExec + " executed "
			exitMessage += "with exit code " + fmt.Sprint(exitCode)
			exitMessage += "\n\n"
			exitMessage += "stdout: \n" + fmt.Sprint(stdoutText)
			exitMessage += "\n\n"
			exitMessage += "stderr: \n" + fmt.Sprint(stderrText)
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
		stdoutText, err := cmd.Output()
		fatal(err)
		// this regex matches: flag, space, quote, path, filename.json, quote
		re, e := regexp.Compile("-f\\s+\"*?.*?(health-checks/)*?[^/]*.json\"*")
		fatal(e)
		// remove this process from consideration
		filtered := re.ReplaceAllString(string(stdoutText), "")
		if strings.Contains(filtered, proc) {
			return 0, ""
		} else {
			return 1, proc + " is not running"
		}
	}
}

// Exists checks to see if a file/dir exists at given path
func Exists(path string) Thunk {
	// does the file at this path exist?
	exists := func(path string) bool {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return false
		}
		return true
	}
	return func() (exitCode int, exitMessage string) {
		if exists(path) {
			return 0, ""
		}
		return 1, "File does not exist: " + path
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
		lineRegex, err := regexp.Compile("Core " + fmt.Sprint(core) + ":?(.*)\\(")
		fatal(err)
		line := lineRegex.Find(out)
		// get temp from that line
		tempRegex, err := regexp.Compile("\\d+\\.\\d*")
		fatal(err)
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
		return 1, "Core temp " + fmt.Sprint(temp) + " exceeds defined max of " + fmt.Sprint(max) + "\n"
	}
}

// Port parses /proc/net/tcp to determine if a given port is in an open state
// and returns an error if it is not.
func Port(port int) Thunk {
	// strHexToDecimal converts from string containing hex number to int
	strHexToDecimal := func(hex string) int {
		portInt, err := strconv.ParseInt(hex, 16, 64)
		fatal(err)
		return int(portInt)
	}

	// getHexPorts gets all open ports as hex strings from /proc/net/tcp
	getHexPorts := func() (ports []string) {
		toReturn := []string{}
		tcp, err := ioutil.ReadFile("/proc/net/tcp")
		fatal(err)
		// matches only the beginnings of lines
		lines := bytes.Split(tcp, []byte("\n"))
		portRe, err := regexp.Compile(":([0-9A-F]{4})")
		for _, line := range lines {
			port := portRe.Find(line) // only get first port, which is local
			if port == nil {
				continue
			}
			portString := string(port[1:])
			fatal(err)
			toReturn = append(toReturn, portString)
		}
		return toReturn
	}

	// getOpenPorts gets a list of open/listening ports as integers
	getOpenPorts := func() (ports []int) {
		for _, port := range getHexPorts() {
			ports = append(ports, strHexToDecimal(port))
		}
		return ports

	}

	return func() (exitCode int, exitMessage string) {
		for _, p := range getOpenPorts() {
			if p == port {
				return 0, ""
			}
		}
		return 1, "Port " + fmt.Sprint(port) + " did not respond."
	}
}
