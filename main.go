// Distributive is a tool for running distributed health checks in server clusters.
// It was designed with Consul in mind, but is platform agnostic.
// The idea is that the checks are run locally, but executed by a central server
// that records and logs their output. This model distributes responsibility to
// each node, instead of one central server, and allows for more types of checks.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

// constructors are registered, have their parameter length checked, and then
// are passed all of Parameters
var workers map[string]Worker = make(map[string]Worker)

// a dictionary with the number of parameters that each method takes
var parameterLength map[string]int = make(map[string]int)

// verbosity settings, provided on the command line
var maxVerbosity int = 2
var minVerbosity int = 0
var verbosity int

// Check is a struct for a unified interface for health checks
// It passes its check-specific fields to that check's Worker
type Check struct {
	Name, Notes string
	Check       string // type of check to run
	Parameters  []string
	Work        Worker
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
	// countInt counts the occurences of int in this []int
	countInt := func(i int, slice []int) (counter int) {
		for _, in := range slice {
			if in == i {
				counter++
			}
		}
		return counter
	}
	// get fail messages
	failMessages := []string{}
	for i, code := range chklst.Codes {
		if code != 0 {
			failMessages = append(failMessages, "\n"+chklst.Messages[i])
		}
	}
	// output global stats
	passed := countInt(0, chklst.Codes)
	failed := countInt(1, chklst.Codes)
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
		if given == 0 {
			msg := "Invalid check:"
			msg += "\n\tCheck type: " + chk.Check
			log.Fatal(msg)
		}
		if given != expected {
			msg := "Invalid check parameters: "
			msg += "\n\tName: " + chk.Name
			msg += "\n\tCheck type: " + chk.Check
			msg += "\n\tExpected: " + fmt.Sprint(expected)
			msg += "\n\tGiven: " + fmt.Sprint(given)
			msg += "\n\tParameters: " + fmt.Sprint(chk.Parameters)
			log.Fatal(msg)
		}
	}
	checkParameterLength(chk, parameterLength[strings.ToLower(chk.Check)])
}

// getWorker returns a Worker based on the Check's name. It also makes sure that
// the correct number of parameters were specified.
func getWorker(chk Check) Worker {
	validateParameters(chk)
	thun := workers[strings.ToLower(chk.Check)]
	if thun == nil {
		msg := "JSON file included one or more unsupported health checks: "
		msg += "\n\tName: " + chk.Name
		msg += "\n\tCheck type: " + chk.Check
		msg += "\n\tParameters: " + fmt.Sprint(chk.Parameters)
		log.Fatal(msg)
		return nil
	}
	return thun
}

// getChecklist loads a JSON file located at path, and Unmarshals it into a
// Checklist struct, leaving unspecified fields as their zero types.
func getChecklist(path string) (chklst Checklist) {
	fileJSON := fileToBytes(path)
	err := json.Unmarshal(fileJSON, &chklst)
	if err != nil {
		log.Fatal("Could not parse JSON at " + path + ":\n\t" + err.Error())
	}
	// Go concurrent pipe - one stage to the next
	// send all checks in checklist to the channel
	out := make(chan Check)
	go func() {
		for _, chk := range chklst.Checklist {
			out <- chk
		}
		close(out)
	}()
	// get Workers for each check
	out2 := make(chan Check)
	go func() {
		for chk := range out {
			chk.Work = getWorker(chk)
			out2 <- chk
		}
		close(out2)
	}()
	// collect data, reassign check list
	var newChecklist []Check
	for chk := range out2 {
		newChecklist = append(newChecklist, chk)
	}
	chklst.Checklist = newChecklist
	return
}

// getVerbosity returns the verbosity specifed by the -v flag, and checks to
// see that it is in a valid range
func getFlags() string {
	verbosityMsg := "Output verbosity level (valid values are "
	verbosityMsg += "[" + fmt.Sprint(minVerbosity) + "-" + fmt.Sprint(maxVerbosity) + "])"
	verbosityMsg += "\n\t 0: Display only errors, with no other output."
	verbosityMsg += "\n\t 1: Display errors and some information."
	verbosityMsg += "\n\t 2: Display everything that's happening."
	pathMsg := "Use the health check JSON located at this path"

	verbosityFlag := flag.Int("v", 1, verbosityMsg)
	path := flag.String("f", "", pathMsg)
	flag.Parse()

	verbosity = *verbosityFlag
	// check for invalid options
	if *path == "" {
		log.Fatal("No path specified. Use -f option.")
	}
	// check for invalid options
	if verbosity > maxVerbosity || verbosity < minVerbosity {
		log.Fatal("Invalid option for verbosity: " + fmt.Sprint(verbosity))
	} else if verbosity >= maxVerbosity {
		fmt.Println("Running with verbosity level " + fmt.Sprint(verbosity))
	}
	return *path
}

// verbosityPrint only prints its message if verbosity is above the given value
func verbosityPrint(str string, minVerb int) {
	if verbosity >= minVerb {
		fmt.Println(str)
	}
}

func runChecks(chklst Checklist) Checklist {
	for _, chk := range chklst.Checklist {
		if chk.Work == nil {
			msg := "Check had a nil function associated with it!"
			msg += " Please submit a bug report with this message."
			msg += "\n\tCheck:" + chk.Check
			msg += "\n\tCheck map: " + fmt.Sprint(workers)
			log.Fatal(msg)
		}
		code, msg := chk.Work(chk.Parameters)
		chklst.Codes = append(chklst.Codes, code)
		chklst.Messages = append(chklst.Messages, msg)
		if code == 0 {
			message := "Check exited with no errors: "
			message += "\n\tName: " + chk.Name
			message += "\n\tType: " + chk.Check
			verbosityPrint(message, maxVerbosity)
		}
	}
	return chklst
}

// main reads the command line flag -f, runs the Check specified in the JSON,
// and exits with the appropriate message and exit code.
func main() {
	// Set up and parse flags
	path := getFlags()

	// add workers to workers, parameterLength
	registerChecks()
	verbosityPrint("Creating checklist...", minVerbosity+1)
	chklst := getChecklist(path)
	// run checks, populate error codes and messages
	verbosityPrint("Running checks...", minVerbosity+1)
	chklst = runChecks(chklst)
	// make a printable report
	chklst.Report = makeReport(chklst)
	// see if any checks failed, exit accordingly
	for _, code := range chklst.Codes {
		if code != 0 {
			verbosityPrint(chklst.Report, minVerbosity)
			os.Exit(1)
		}
	}
	verbosityPrint(chklst.Report, maxVerbosity)
	os.Exit(0)
}
