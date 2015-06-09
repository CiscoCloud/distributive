// Distributive is a tool for running distributed health checks in server clusters.
// It was designed with Consul in mind, but is platform agnostic.
// The idea is that the checks are run locally, but executed by a central server
// that records and logs their output. This model distributes responsibility to
// each node, instead of one central server, and allows for more types of checks.
package main

// TODO rework public and private methods. Likely, nothing needs to be public
// from the thunk side of things

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

var maxVerbosity int = 2
var minVerbosity int = 0
var verbosity int // global program verbosity

// Check is a struct for a unified interface for health checks
// It passes its check-specific fields to that check's Thunk constructor
type Check struct {
	Name, Notes string
	Check       string // type of check to run
	Parameters  []string
	Fun         Thunk
}

// Checklist is a struct that provides a concise way of thinking about doing
// several checks and then returning some kind of output.
type Checklist struct {
	Name, Notes string
	Checklist   []Check // list of Checks to run
	Codes       []int
	Messages    []string
	Report      string
}

// makeReport returns a string used for a checklist.Report attribute, printed
// after all the checks have been run
func makeReport(chklst Checklist) (report string) {
	failMessages := []string{}
	passed := 0
	failed := 0
	for i, code := range chklst.Codes {
		if code != 0 {
			failed++
			failMessages = append(failMessages, "\n"+chklst.Messages[i])
		} else {
			passed++
		}
	}
	// No output on all sucess - the Unix way!
	if failed == 0 {
		return ""
	}
	report += "Passed: " + fmt.Sprint(passed) + "\n"
	report += "Failed: " + fmt.Sprint(failed) + "\n"
	for _, msg := range failMessages {
		report += msg
	}
	return report
}

// validateParameters asks whether or not this check has the correct number of
// parameters specified
func validateParameters(chk Check) {
	// checkParameterLength ensures that the Check has the proper number of
	// parameters, and exits otherwise. Can't do much with a broken check!
	checkParameterLength := func(chk Check, expected int) {
		given := len(chk.Parameters)
		if given != expected {
			msg := "Invalid check parameters for check: " + chk.Name
			msg += "\nCheck type: " + chk.Check
			msg += "\nExpected: " + fmt.Sprint(expected)
			msg += "\nGiven: " + fmt.Sprint(given)
			msg += "\nParameters: " + fmt.Sprint(chk.Parameters)
			log.Fatal(msg)
		}
	}
	// a dictionary with the number of parameters that each method takes
	numParameters := map[string]int{
		"command": 1, "running": 1, "file": 1, "directory": 1, "symlink": 1,
		"installed": 1, "ppa": 1, "checksum": 3, "temp": 1, "port": 1,
		"interface": 1, "up": 1, "ip4": 2, "ip6": 2, "gateway": 1,
		"gatewayinterface": 1, "host": 1, "tcp": 1, "udp": 1, "module": 1,
		"kernelparameter": 1, "dockerimage": 1, "dockerrunning": 1,
		"groupexists": 1, "useringroup": 2, "groupid": 2, "userexists": 1,
		"userhasuid": 2, "userhasgid": 2, "userhasusername": 2, "userhasname": 2,
		"userhashomedir": 2, "yumrepo": 1, "yumrepourl": 1,
	}
	checkParameterLength(chk, numParameters[strings.ToLower(chk.Check)])
}

// getThunk passes a Check's parameters to the correct Thunk constructor based
// on the Check's name. It also makes sure that the correct number of parameters
// were specified.
func getThunk(chk Check) Thunk {
	validateParameters(chk)
	switch strings.ToLower(chk.Check) {
	case "command":
		return Command(chk.Parameters[0])
	case "running":
		return Running(chk.Parameters[0])
	case "file":
		return File(chk.Parameters[0])
	case "directory":
		return Directory(chk.Parameters[0])
	case "symlink":
		return Symlink(chk.Parameters[0])
	case "checksum":
		return Checksum(chk.Parameters[0], chk.Parameters[1], chk.Parameters[2])
	case "temp":
		tempInt, err := strconv.ParseInt(chk.Parameters[0], 10, 32)
		if err != nil {
			log.Fatal("Could not parse temperature: " + chk.Parameters[0])
		}
		return Temp(int(tempInt))
	case "port":
		portInt, err := strconv.ParseInt(chk.Parameters[0], 10, 32)
		if err != nil {
			log.Fatal("Could not parse port number: " + chk.Parameters[0])
		}
		return Port(int(portInt))
	case "interface":
		return Interface(chk.Parameters[0])
	case "up":
		return Up(chk.Parameters[0])
	case "ip4":
		return Ip4(chk.Parameters[0], chk.Parameters[1])
	case "ip6":
		return Ip6(chk.Parameters[0], chk.Parameters[1])
	case "gateway":
		return Gateway(chk.Parameters[0])
	case "gatewayinterface":
		return GatewayInterface(chk.Parameters[0])
	case "host":
		return Host(chk.Parameters[0])
	case "tcp":
		return TCP(chk.Parameters[0])
	case "udp":
		return UDP(chk.Parameters[0])
	case "module":
		return Module(chk.Parameters[0])
	case "kernelparameter":
		return KernelParameter(chk.Parameters[0])
	case "dockerimage":
		return DockerImage(chk.Parameters[0])
	case "dockerrunning":
		return DockerRunning(chk.Parameters[0])
	case "groupexists":
		return GroupExists(chk.Parameters[0])
	case "useringroup":
		return UserInGroup(chk.Parameters[0], chk.Parameters[1])
	case "groupid":
		tempInt, err := strconv.ParseInt(chk.Parameters[1], 10, 32)
		if err != nil {
			log.Fatal("Could not parse group ID for group: " + chk.Parameters[0])
		}
		return GroupId(chk.Parameters[0], int(tempInt))
	case "userexists":
		return UserExists(chk.Parameters[0])
	case "userhasuid":
		return UserHasUID(chk.Parameters[0], chk.Parameters[1])
	case "userhasgid":
		return UserHasGID(chk.Parameters[0], chk.Parameters[1])
	case "userhasusername":
		return UserHasUsername(chk.Parameters[0], chk.Parameters[1])
	case "userhasname":
		return UserHasName(chk.Parameters[0], chk.Parameters[1])
	case "userhashomedir":
		return UserHasHomeDir(chk.Parameters[0], chk.Parameters[1])
	case "installed":
		return Installed(chk.Parameters[0])
	case "ppa":
		return PPA(chk.Parameters[0])
	case "yumrepo":
		return YumRepoExists(chk.Parameters[0])
	case "yumrepourl":
		return YumRepoURL(chk.Parameters[0])
	default:
		msg := "JSON file included one or more unsupported health checks: "
		msg += "\n\tName: " + chk.Name
		msg += "\n\tCheck type: " + chk.Check
		msg += "\n\tParameters: " + fmt.Sprint(chk.Parameters)
		log.Fatal(msg)
		return nil
	}
}

// getChecklist loads a JSON file located at path, and Unmarshals it into a
// Checklist struct, leaving unspecified fields as their zero types.
func getChecklist(path string) (chklst Checklist) {
	fileJSON := fileToBytes(path)
	err := json.Unmarshal(fileJSON, &chklst)
	if err != nil {
		log.Fatal("Could not parse JSON at " + path + ":\n\t" + err.Error())
	}
	// TODO make this concurrent
	for i, _ := range chklst.Checklist {
		chklst.Checklist[i].Fun = getThunk(chklst.Checklist[i])
	}
	return
}

// main reads the command line flag -f, runs the Check specified in the JSON,
// and exits with the appropriate message and exit code.
func main() {
	// Set up and parse flags
	verbosityMsg := "Output verbosity level (valid values are "
	verbosityMsg += "[" + fmt.Sprint(minVerbosity) + "-" + fmt.Sprint(maxVerbosity) + "])"
	verbosityMsg += "\n\t 0: (Default) Display only errors, with no other output."
	verbosityMsg += "\n\t 1: Display errors and some information."
	verbosityMsg += "\n\t 2: Display everything that's happening."
	pathMsg := "Use the health check JSON located at this path"
	path := flag.String("f", "", pathMsg)
	verbosityFlag := flag.Int("v", 0, verbosityMsg)
	flag.Parse()
	verbosity = *verbosityFlag
	// check for invalid options
	if verbosity > maxVerbosity || verbosity < minVerbosity {
		log.Fatal("Invalid option for verbosity: " + fmt.Sprint(verbosity))
	} else if verbosity > -5 { // TODO change this to geq max
		fmt.Println("Running with verbosity level " + fmt.Sprint(verbosity))
	} else if *path == "" {
		log.Fatal("No path specified. Use -f option.")
	}

	chklst := getChecklist(*path)
	if verbosity > minVerbosity {
		fmt.Println("Creating checklist...")
	}
	// run checks, populate error codes and messages
	// TODO make this concurrent
	for _, chk := range chklst.Checklist {
		if verbosity > 2 {
			name := ": " + chk.Name
			if chk.Name == "" {
				name = ""
			}
			fmt.Println("Running check" + name + " of type: " + chk.Check)
		}
		code, message := chk.Fun()
		chklst.Codes = append(chklst.Codes, code)
		chklst.Messages = append(chklst.Messages, message)
	}
	// make a printable report
	chklst.Report = makeReport(chklst) // run tests, get messages
	fmt.Println(chklst.Report)
	// exit with the proper code
	for _, code := range chklst.Codes {
		if code != 0 {
			os.Exit(1)
		}
	}
	os.Exit(0)
}
