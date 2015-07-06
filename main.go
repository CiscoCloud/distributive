// Distributive is a tool for running distributed health checks in server clusters.
// It was designed with Consul in mind, but is platform agnostic.
// The idea is that the checks are run locally, but executed by a central server
// that records and logs their output. This model distributes responsibility to
// each node, instead of one central server, and allows for more types of checks.
package main

import (
	"encoding/json"
	"fmt"
	"github.com/CiscoCloud/distributive/workers"
	"github.com/CiscoCloud/distributive/wrkutils"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

// where remote checks are downloaded to
var remoteCheckDir = "/var/run/distributive/"

// Check is a struct for a unified interface for health checks
// It passes its check-specific fields to that check's workers.Worker
type Check struct {
	Name, Notes string
	Check       string // type of check to run
	Parameters  []string
	Work        wrkutils.Worker
}

// Checklist is a struct that provides a concise way of thinking about doing
// several checks and then returning some kind of output.
type Checklist struct {
	Name, Notes string
	Checklist   []Check // list of Checks to run
	Codes       []int
	Messages    []string
	Report      string
	Origin      string // where did it come from?
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
	total := len(chklst.Codes)
	passed := countInt(0, chklst.Codes)
	failed := countInt(1, chklst.Codes)
	report += "Total: " + fmt.Sprint(total) + "\n"
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
		if given == 0 || expected == 0 {
			log.WithFields(log.Fields{
				"check type": chk.Check,
			}).Fatal("Invalid check")
		}
		if given != expected {
			log.WithFields(log.Fields{
				"name":       chk.Name,
				"check type": chk.Check,
				"expected":   fmt.Sprint(expected),
				"given":      fmt.Sprint(given),
				"parameters": fmt.Sprint(chk.Parameters),
			}).Fatal("Invalid check parameters")
		}
	}
	checkParameterLength(chk, wrkutils.ParameterLength[strings.ToLower(chk.Check)])
}

// getworkers.Worker returns a workers.Worker based on the Check's name. It also makes sure that
// the correct number of parameters were specified.
func getWorker(chk Check) wrkutils.Worker {
	validateParameters(chk)
	work := wrkutils.Workers[strings.ToLower(chk.Check)]
	if work == nil {
		log.WithFields(log.Fields{
			"name":       chk.Name,
			"type":       chk.Check,
			"parameters": chk.Parameters,
		}).Fatal("JSON file included one or more unsupported health checks")
		return nil
	}
	return work
}

// loadRemoteChecklist either downloads a checklist from a remote URL and puts
// it in /etc/distributive/url.json
func loadRemoteChecklist(urlstr string) (chklst Checklist) {
	// urlToFile gets the response from urlstr and writes it to path
	urlToFile := func(urlstr string, path string) error {
		body := workers.URLToBytes(urlstr, true) // secure connection
		// write to file
		err := ioutil.WriteFile(path, body, 0755)
		if err != nil {
			wrkutils.CouldntWriteError(path, err)
		}
		return nil
	}
	// ensure temp files dir exists
	log.Info("Creating/checking remote checklist dir")
	if err := os.MkdirAll(remoteCheckDir, 0775); err != nil {
		log.WithFields(log.Fields{
			"dir":   remoteCheckDir,
			"error": err.Error(),
		}).Fatal("Could not create temporary file directory:")
	}

	// write out the response to a file
	// filter these chars: /?%*:|<^>. \
	pathRegex := regexp.MustCompile("[\\/\\?%\\*:\\|\"<\\^>\\.\\ ]")
	filename := pathRegex.ReplaceAllString(urlstr, "") + ".json"
	fullpath := remoteCheckDir + filename
	// only create it if it doesn't exist
	if _, err := os.Stat(fullpath); err != nil {
		log.Info("Fetching remote checklist")
		urlToFile(urlstr, fullpath)
	} else {
		log.Debug("Using local copy of remote checklist")
	}
	// return a real checklist
	return getChecklist(fullpath)
}

// getChecklist loads a JSON file located at path, and Unmarshals it into a
// Checklist struct, leaving unspecified fields as their zero types.
func getChecklist(path string) (chklst Checklist) {
	fileJSON := wrkutils.FileToBytes(path)
	err := json.Unmarshal(fileJSON, &chklst)
	if err != nil {
		log.WithFields(log.Fields{
			"path":    path,
			"error":   err.Error(),
			"content": string(fileJSON),
		}).Fatal("Couldn't parse JSON file")
	}
	//// Go concurrent pipe - one stage to the next
	// send all checks in checklist to the channel
	out := make(chan Check)
	go func() {
		for _, chk := range chklst.Checklist {
			out <- chk
		}
		close(out)
	}()
	// get workers.Workers for each check
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
	chklst.Origin = path
	return
}

// getChecklistsInDir uses getChecklist to construct a checklist struct for
// every .json file in a directory
func getChecklistsInDir(dirpath string) (chklsts []Checklist) {
	paths := wrkutils.GetFilesWithExtension(dirpath, ".json")
	for _, path := range paths {
		chklsts = append(chklsts, getChecklist(path))
	}
	return chklsts
}

// runChecks takes a checklist, performs every worker, and collects the results
// in that checklist's Codes and Messages fields. TODO: Create concurrent pipe.
func runChecks(chklst Checklist) Checklist {
	for _, chk := range chklst.Checklist {
		if chk.Work == nil {
			msg := "Nil function associated with this check."
			msg += " Please submit a bug report with this message."
			log.WithFields(log.Fields{
				"check":     chk.Check,
				"check map": fmt.Sprint(wrkutils.Workers),
			}).Fatal(msg)
		}
		code, msg := chk.Work(chk.Parameters)
		chklst.Codes = append(chklst.Codes, code)
		chklst.Messages = append(chklst.Messages, msg)
		no := ""
		if code == 0 {
			no = " no"
		}
		log.WithFields(log.Fields{
			"name": chk.Name,
			"type": chk.Check,
		}).Info("Check exited with" + no + " errors")
	}
	return chklst
}

// main reads the command line flag -f, runs the Check specified in the JSON,
// and exits with the appropriate message and exit code.
func main() {
	// Set up and parse flags
	file, URL, directory := getFlags()
	validateFlags(file, URL, directory)

	// add workers to workers, parameterLength
	workers.RegisterAll()
	log.Info("Creating checklist(s)...")
	// load checklists according to flags
	var chklsts []Checklist
	if file != "" {
		chklsts = append(chklsts, getChecklist(file))
	} else if URL != "" {
		chklsts = append(chklsts, loadRemoteChecklist(URL))
	} else if directory != "" {
		chklsts = append(chklsts, getChecklistsInDir(directory)...)
	} else {
		log.Fatal("Neither file nor URL nor directory specified. Try --help.")
	}
	// run all checklists
	// TODO: use concurrent pipe here
	for i := range chklsts {
		// run checks, populate error codes and messages
		log.Info("Running checklist: " + chklsts[i].Name)
		chklsts[i] = runChecks(chklsts[i])
		// make a printable report
		report := makeReport(chklsts[i])
		chklsts[i].Report = report
	}
	// see if any checks failed, exit accordingly
	exitStatus := 0
	for _, chklst := range chklsts {
		failed := false
		for _, code := range chklst.Codes {
			if code != 0 {
				failed = true
				exitStatus = 1
			}
		}
		if failed {
			log.WithFields(log.Fields{
				"checklist": chklst.Name,
				"report":    chklst.Report,
			}).Warn("Some checks failed, printing checklist report")
		} else {
			log.WithFields(log.Fields{
				"checklist": chklst.Name,
				"report":    chklst.Report,
			}).Info("All checks passed, printing checklist report")
		}
	}
	os.Exit(exitStatus)
}
