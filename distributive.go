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
	Id, Name, Notes, Service_id         string
	Command, Running, Exists, Installed string // hold inputs for checks
	Port, Temp                          int    // more check inputs
	Fun                                 Thunk
}

// getThunk passes a Check to the proper Thunk constructor based on which
// of its fields were filled when it was read from JSON.
// Fields that weren't specified in the JSON take on zero values for their type
func getThunk(chk Check) Thunk {
	if chk.Command != "" {
		return Command(chk.Command)
	} else if chk.Running != "" {
		return Running(chk.Running)
	} else if chk.Exists != "" {
		return Exists(chk.Exists)
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

// getCheck loads a JSON file located at path, and Unmarshals it into a Check
// struct, leaving unspecified fields as their zero types.
func getCheck(path string) (chk Check) {
	fileJSON, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal("Couldn't find .json at specified location: " + path)
	}
	err = json.Unmarshal(fileJSON, &chk)
	fatal(err)
	chk.Fun = getThunk(chk)
	return chk
}

// main reads the command line flag -f, runs the Check specified in the JSON,
// and exits with the appropriate message and exit code.
func main() {
	// run a check, print output, and exit
	run := func(chk Check) {
		c, m := chk.Fun()
		fmt.Printf(m)
		os.Exit(c) // exit with Consul-compatible error code : 0 | 1
	}

	fn := flag.String("f", "", "Use the health check JSON located at this path")
	flag.Parse()
	chk := getCheck(*fn)
	run(chk)
}
