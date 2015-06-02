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
	Name, Notes                 string
	Command, Installed, Running string
	File, Directory, Symlink    string
	Port, Temp                  int // more check inputs
	Fun                         Thunk
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
	if chk.Command != "" {
		return Command(chk.Command)
	} else if chk.Running != "" {
		return Running(chk.Running)
	} else if chk.File != "" {
		return File(chk.File)
	} else if chk.Directory != "" {
		return Directory(chk.Directory)
	} else if chk.Symlink != "" {
		return Symlink(chk.Symlink)
	} else if chk.Installed != "" {
		return Installed(chk.Installed)
	} else if chk.Temp != 0 {
		return Temp(chk.Temp)
	} else if chk.Port != 0 {
		return Port(chk.Port)
	} else {
		log.Fatal("JSON file didn't include any supported health check types")
	}
	return nil
}

// getChecklist loads a JSON file located at path, and Unmarshals it into a
// Checklist struct, leaving unspecified fields as their zero types.
func getChecklist(path string) (chklst Checklist) {
	//var list []Check
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
