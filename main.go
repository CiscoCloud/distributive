// Distributive is a tool for running distributed health checks in server clusters.
// It was designed with Consul in mind, but is platform agnostic.
// The idea is that the checks are run locally, but executed by a central server
// that records and logs their output. This model distributes responsibility to
// each node, instead of one central server, and allows for more types of checks.
package main

import (
	"os"

	"github.com/CiscoCloud/distributive/checklists"
	log "github.com/Sirupsen/logrus"
	"github.com/mitchellh/panicwrap"
    _ "github.com/CiscoCloud/distributive/checks"
)

var useCache bool // should remote checks be run from the cache when possible?

const Version = "v0.2.5"
const Name = "distributive"

// getChecklists returns a list of checklists based on the supplied sources
func getChecklists(file string, dir string, url string, stdin bool) (lsts []checklists.Checklist) {
	parseError := func(src string, err error) {
		if err != nil {
			log.WithFields(log.Fields{
				"origin": src,
				"error":  err,
			}).Fatal("Couldn't parse checklist.")
		}
	}
	msg := "Creating checklist(s)..."
	switch {
	// checklists from file are already tagged with their origin
	// this applies to FromFile, FromDirectory, FromURL
	case file != "":
		log.WithFields(log.Fields{
			"type": "file",
			"path": file,
		}).Info(msg)
		chklst, err := checklists.FromFile(file)
		parseError(file, err)
		lsts = append(lsts, chklst)
	case dir != "":
		log.WithFields(log.Fields{
			"type": "dir",
			"path": dir,
		}).Info(msg)
		chklsts, err := checklists.FromDirectory(dir)
		parseError(dir, err)
		lsts = append(lsts, chklsts...)
	case url != "":
		log.WithFields(log.Fields{
			"type": "url",
			"path": url,
		}).Info(msg)
		chklst, err := checklists.FromURL(url, useCache)
		parseError(url, err)
		lsts = append(lsts, chklst)
	case stdin == true:
		log.WithFields(log.Fields{
			"type": "url",
			"path": url,
		}).Info(msg)
		checklist, err := checklists.FromStdin()
		checklist.Origin = "stdin" // TODO put this in the method
		parseError("stdin", err)
		lsts = append(lsts, checklist)
	default:
		log.Fatal("Neither file, URL, directory, nor stdin specified. Try --help.")
	}
	return lsts
}

// main reads the command line flag -f, runs the Check specified in the YAML,
// and exits with the appropriate message and exit code.
func main() {
	// Set up global panic handling
	exitStatus, err := panicwrap.BasicWrap(panicHandler)
	if err != nil {
		reportURL := "https://github.com/mitchellh/panicwrap"
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Fatal("Please report this error to " + reportURL)
	}
	// If exitStatus >= 0, then we're the parent process and the panicwrap
	// re-executed ourselves and completed. Just exit with the proper status.
	if exitStatus >= 0 {
		os.Exit(exitStatus)
	}
	// Otherwise, exitStatus < 0 means we're the child. Continue executing as
	// normal...

	// Set up and parse flags
	log.Debug("Parsing flags")
	file, URL, directory, stdin := getFlags()
	log.Debug("Validating flags")
	validateFlags(file, URL, directory)
	// add workers to workers, parameterLength
	log.Debug("Running checklists")
	exitCode := 0
	for _, chklst := range getChecklists(file, directory, URL, stdin) {
		anyFailed, report := chklst.MakeReport()
		if anyFailed {
			exitCode = 1
		}
		log.WithFields(log.Fields{
			"checklist": chklst.Name,
			"report":    report,
		}).Info("Report from checklist")
	}
	os.Exit(exitCode)
}
