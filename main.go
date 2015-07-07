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
	Origin      string // where did it come from?
}

// makeReport returns a string used for a checklist.Report attribute, printed
// after all the checks have been run
// TODO transform use of makeReport into a logReport that uses logrus
func (chklst *Checklist) makeReport() (report string) {
	if chklst == nil {
		return ""
	}
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
	report += "\nTotal: " + fmt.Sprint(total)
	report += "\nPassed: " + fmt.Sprint(passed)
	report += "\nFailed: " + fmt.Sprint(failed)
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

// checklistFromBytes takes a bytestring of utf8 encoded JSON and turns it into
// a checklist struct. Used by all checklist constructors below.
func checklistFromBytes(data []byte) (chklst Checklist) {
	err := json.Unmarshal(data, &chklst)
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err.Error(),
			"content": string(data),
		}).Fatal("Couldn't parse checklist JSON.")
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
	var listOfChecks []Check
	for chk := range out2 {
		listOfChecks = append(listOfChecks, chk)
	}
	chklst.Checklist = listOfChecks
	return
}

// checklistFromFile reads the file at the path and parses its utf8 encoded json
// data, turning it into a checklist struct.
func checklistFromFile(path string) (chklst Checklist) {
	return checklistFromBytes(wrkutils.FileToBytes(path))
}

// checklistFromStdin reads the stdin pipe and parses its utf8 encoded json
// data, turning it into a checklist struct.
func checklistFromStdin() (chklst Checklist) {
	stdinAsBytes := func() []byte {
		bytes, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Fatal("Couldn't read from stdin")
		}
		return bytes
	}
	return checklistFromBytes(stdinAsBytes())
}

// checklistsFromDir reads all of the files in the path and parses their utf8
// encoded json data, turning it into a checklist struct.
func checklistsFromDir(dirpath string) (chklsts []Checklist) {
	paths := wrkutils.GetFilesWithExtension(dirpath, ".json")
	for _, path := range paths {
		chklsts = append(chklsts, checklistFromFile(path))
	}
	return chklsts
}

// checklistsFromDir reads data retrieved from the URL and parses its utf8
// encoded json data, turning it into a checklist struct. It also caches this
// data at remoteCheckDir, currently "/var/run/distributive/"
func checklistFromURL(urlstr string) (chklst Checklist) {
	// ensure temp files dir exists
	log.Debug("Creating/checking remote checklist dir")
	if err := os.MkdirAll(remoteCheckDir, 0775); err != nil {
		log.WithFields(log.Fields{
			"dir":   remoteCheckDir,
			"error": err.Error(),
		}).Fatal("Could not create temporary file directory:")
	}

	// write out the response to a file
	// filter these (path illegal) chars: /?%*:|<^>. \
	// TODO use a golang loop with straight up strings, instead of regexp
	pathRegex := regexp.MustCompile("[\\/\\?%\\*:\\|\"<\\^>\\.\\ ]")
	filename := pathRegex.ReplaceAllString(urlstr, "") + ".json"
	fullpath := remoteCheckDir + filename
	// only create it if it doesn't exist
	if _, err := os.Stat(fullpath); err != nil {
		log.Info("Fetching remote checklist")
		body := wrkutils.URLToBytes(urlstr, true) // secure connection
		log.Debug("Writing remote checklist to cache")
		wrkutils.BytesToFile(body, fullpath)
		return checklistFromBytes(body)
	}
	log.WithFields(log.Fields{
		"path": fullpath,
	}).Info("Using local copy of remote checklist")
	return checklistFromFile(fullpath)
}

// getChecklists returns a list of checklists based on the supplied sources
func getChecklists(file string, dir string, url string, stdin bool) (checklists []Checklist) {
	msg := "Creating checklist(s)..."
	switch {
	// checklists from file are already tagged with their origin
	// this applies to FromFile, FromDir, FromURL
	case file != "":
		log.WithFields(log.Fields{
			"type": "file",
			"path": file,
		}).Info(msg)
		checklists = append(checklists, checklistFromFile(file))
	case dir != "":
		log.WithFields(log.Fields{
			"type": "dir",
			"path": dir,
		}).Info(msg)
		checklists = append(checklists, checklistsFromDir(dir)...)
	case url != "":
		log.WithFields(log.Fields{
			"type": "url",
			"path": url,
		}).Info(msg)
		checklists = append(checklists, checklistFromURL(url))
	case stdin == true:
		log.WithFields(log.Fields{
			"type": "url",
			"path": url,
		}).Info(msg)
		checklist := checklistFromStdin()
		checklist.Origin = "stdin"
		checklists = append(checklists, checklist)
	default:
		log.Fatal("Neither file, URL, directory, nor stdin specified. Try --help.")
	}
	return checklists
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
		// were errors encountered?
		no := ""
		if code == 0 {
			no = " no"
		}
		// warn log happens later
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
	file, URL, directory, stdin := getFlags()
	validateFlags(file, URL, directory)

	// add workers to workers, parameterLength
	workers.RegisterAll()
	chklsts := getChecklists(file, directory, URL, stdin)
	// run all checklists
	// TODO: use concurrent pipe here
	exitStatus := 0
	for i, chklst := range chklsts {
		// run checks, populate error codes and messages
		log.Info("Running checklist: " + chklsts[i].Name)
		chklsts[i] = runChecks(chklsts[i])
		// make a printable report
		report := chklsts[i].makeReport()
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
				"report":    report,
			}).Warn("Some checks failed, printing checklist report")
		} else {
			log.WithFields(log.Fields{
				"checklist": chklst.Name,
				"report":    report,
			}).Info("All checks passed, printing checklist report")
		}
	}
	// see if any checks failed, exit accordingly
	os.Exit(exitStatus)
}
