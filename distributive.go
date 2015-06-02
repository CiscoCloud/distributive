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
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

// fatal simplifies error handling (instead of an if err != nil block)
func fatal(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

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
	report += "Passed: " + fmt.Sprint(passed) + "\n"
	report += "Failed: " + fmt.Sprint(failed) + "\n"
	for _, msg := range failMessages {
		report += msg
	}
	return report
}

// getThunk passes a Check to the proper Thunk constructor based on which
// of its fields were filled when it was read from JSON.
// Fields that weren't specified in the JSON take on zero values for their type
func getThunk(chk Check) Thunk {
	// TODO handle errors here, log effectively (separate function?)
	switch chk.Check {
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
	case "installed":
		return Installed(chk.Parameters[0])
	case "temp":
		tempInt, err := strconv.ParseInt(chk.Parameters[0], 10, 32)
		fatal(err)
		return Temp(int(tempInt))
	case "port":
		portInt, err := strconv.ParseInt(chk.Parameters[0], 10, 32)
		fatal(err)
		return Port(int(portInt))
	case "interface":
		return Interface(chk.Parameters[0])
	default:
		msg := "JSON file included one or more unsupported health checks: "
		log.Fatal(msg + chk.Check)
		return nil
	}
}

// getChecklist loads a JSON file located at path, and Unmarshals it into a
// Checklist struct, leaving unspecified fields as their zero types.
func getChecklist(path string) (chklst Checklist) {
	fileJSON, err := ioutil.ReadFile(path)
	if err != nil {
		if path == "" {
			log.Fatal("No path specified (use -f option)")
		}
		log.Fatal("Couldn't read JSON at specified location: " + path)
	}
	err = json.Unmarshal(fileJSON, &chklst)
	fatal(err)
	for i, _ := range chklst.Checklist {
		chklst.Checklist[i].Fun = getThunk(chklst.Checklist[i])
	}
	return
}

// main reads the command line flag -f, runs the Check specified in the JSON,
// and exits with the appropriate message and exit code.
func main() {
	fn := flag.String("f", "", "Use the health check JSON located at this path")
	flag.Parse()
	chklst := getChecklist(*fn)
	// run checks, populate error codes and messages
	for _, chk := range chklst.Checklist {
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
