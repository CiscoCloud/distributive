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
	"path/filepath"
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
	Report      string // printable status update
	Failed      bool   // did any of the checks fail?
}

// makeReport returns a string used for a checklist.Report attribute, printed
// after all the checks have been run
// TODO transform use of makeReport into a logReport that uses logrus
func (chklst *Checklist) makeReport() {
	if chklst == nil {
		return
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
	report := "↴\nTotal: " + fmt.Sprint(total)
	report += "\nPassed: " + fmt.Sprint(passed)
	report += "\nFailed: " + fmt.Sprint(total-passed)
	for _, msg := range failMessages {
		report += msg
	}
	chklst.Report = report
}

// validateParameters asks whether or not this check has the correct number of
// parameters specified
func validateParameters(chk Check) {
	// checkParameterLength ensures that the Check has the proper number of
	// parameters, and exits otherwise. Can't do much with a broken check!
	checkParameterLength := func(chk Check, expected int) {
		given := len(chk.Parameters)
		if expected == 0 {
			log.WithFields(log.Fields{
				"name":       chk.Name,
				"check type": chk.Check,
				"parameters": chk.Parameters,
			}).Fatal("Invalid check")
		} else if given != expected {
			log.WithFields(log.Fields{
				"name":       chk.Name,
				"check type": chk.Check,
				"expected":   expected,
				"given":      given,
				"parameters": chk.Parameters,
			}).Fatal("Invalid check parameters")
		}
	}
	// for testing this independently of main, shouldn't run outside of testing
	if len(wrkutils.ParameterLength) < 1 {
		workers.RegisterAll()
	}
	if len(wrkutils.ParameterLength) < 1 {
		log.WithFields(log.Fields{
			"name":       chk.Name,
			"check type": chk.Check,
			"given":      len(chk.Parameters),
			"parameters": chk.Parameters,
		}).Fatal("wrkutils.ParameterLength table is empty")
	}
	expected := wrkutils.ParameterLength[strings.ToLower(chk.Check)]
	checkParameterLength(chk, expected)
}

// getworkers.Worker returns a workers.Worker based on the Check's name. It
// ensures that any invalid checks are reported appropriately.
func getWorker(chk Check) wrkutils.Worker {
	work := wrkutils.Workers[strings.ToLower(chk.Check)]
	if work == nil {
		msg := "JSON file included one or more unsupported health checks"
		msg2 := "(check lookup returned nil function)"
		log.WithFields(log.Fields{
			"name":       chk.Name,
			"type":       chk.Check,
			"parameters": chk.Parameters,
		}).Fatal(msg + " " + msg2)
		return nil
	}
	return work
}

// checklistFromBytes takes a bytestring of utf8 encoded JSON and turns it into
// a checklist struct. Used by all checklist constructors below. It validates
// the number of parameters that each check has.
func checklistFromBytes(data []byte) (chklst Checklist) {
	err := json.Unmarshal(data, &chklst)
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err.Error(),
			"content": string(data),
		}).Fatal("Couldn't parse checklist JSON.")
	}
	// get workers for each check
	out := make(chan Check)
	defer close(out)
	for _, chk := range chklst.Checklist {
		go func(chk Check, out chan Check) {
			validateParameters(chk)
			chk.Work = getWorker(chk)
			out <- chk
		}(chk, out)
	}
	// grab all the data from the channel, mutating the checklist
	for i := range chklst.Checklist {
		chklst.Checklist[i] = <-out
	}
	return chklst
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
	// send one checklist per path to the channel
	out := make(chan Checklist)
	for _, path := range paths {
		go func(path string, out chan Checklist) {
			out <- checklistFromFile(path)
		}(path, out)
	}
	// get all values from the channel, return them
	for _ = range paths {
		chklsts = append(chklsts, <-out)
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
		}).Warn("Could not create remote check directory")
		remoteCheckDir = "./.remote-checks"
		if err := os.MkdirAll(remoteCheckDir, 0755); err != nil {
			wrkutils.CouldntWriteError(remoteCheckDir, err)
		}
	}
	log.Debug("Using " + remoteCheckDir + " for remote check storage")

	// pathSanitize filters these (path illegal) chars: /?%*:|<^>. \
	pathSanitize := func(str string) (filename string) {
		filename = str
		disallowed := []string{
			`/`, `?`, `%`, `*`, `:`, `|`, `"`, `<`, `^`, `>`, `.`, `\`, ` `,
		}
		for _, c := range disallowed {
			filename = strings.Replace(filename, c, "", -1)
		}
		return filename
	}
	filename := pathSanitize(urlstr) + ".json"
	fullpath := filepath.Join(remoteCheckDir, filename)

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
// in that checklist's Codes and Messages fields.
func (chklst *Checklist) runChecks() {
	codes := make(chan int)
	msgs := make(chan string)
	for _, chk := range chklst.Checklist {
		// concurrently execute the checklist's checks, passing their
		go func(chk Check, codes chan int, msgs chan string) {
			if chk.Work == nil {
				msg := "Nil function associated with this check."
				msg += " Please submit a bug report with this message."
				log.WithFields(log.Fields{
					"check":     chk.Check,
					"check map": fmt.Sprint(wrkutils.Workers),
				}).Fatal(msg)
			}
			code, msg := chk.Work(chk.Parameters)
			// Log an informational message on the check's status
			if code == 0 {
				log.WithFields(log.Fields{
					"name": chk.Name,
					"type": chk.Check,
				}).Info("Check passed")
			} else {
				log.WithFields(log.Fields{
					"name": chk.Name,
					"type": chk.Check,
				}).Warn("Check failed")
				chklst.Failed = true
			}
			// send back results
			codes <- code
			msgs <- msg
		}(chk, codes, msgs)
	}
	// consume codes and messages, adding them to the Checklist struct
	for _ = range chklst.Checklist {
		code := <-codes
		msg := <-msgs
		chklst.Codes = append(chklst.Codes, code)
		chklst.Messages = append(chklst.Messages, msg)
	}
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
	out := make(chan Checklist)
	defer close(out)
	for i, chklst := range chklsts {
		go func(chklst Checklist, out chan Checklist) {
			// run checks, populate error codes and messages
			log.Info("Running checklist: " + chklsts[i].Name)
			chklst.runChecks()
			chklst.makeReport()
			// If any of the checks failed, mark this checklist as failed
			for _, code := range chklsts[i].Codes {
				if code != 0 {
					chklst.Failed = true
				}
			}
			// send out results
			out <- chklst
		}(chklst, out)
	}
	failed := false
	for _ = range chklsts {
		chklst := <-out
		if chklst.Failed {
			log.WithFields(log.Fields{
				"checklist": chklst.Name,
				"report":    chklst.Report,
			}).Warn("Check(s) failed, printing checklist report")
			failed = true
		} else {
			log.WithFields(log.Fields{
				"checklist": chklst.Name,
				"report":    chklst.Report,
			}).Info("All checks passed, printing checklist report")
		}
	}
	// see if any checks failed, exit accordingly
	if failed {
		os.Exit(1)
	}
	os.Exit(0)
}
