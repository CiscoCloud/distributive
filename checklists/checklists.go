package checklists

import (
	"encoding/json"
	"fmt"
	"github.com/CiscoCloud/distributive/chkutil"
	"github.com/CiscoCloud/distributive/errutil"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// where remote checks are downloaded to
var remoteCheckDir = "/var/run/distributive/"

/***************** Checklist type *****************/

// Checklist is a struct that provides a concise way of thinking about doing
// several checks and then returning some kind of output.
type Checklist struct {
	Name, Notes string
	Checks      []chkutil.Check // list of chkutil.Checks to run
	Origin      string          // where did it come from?
}

// MakeReport runs all checks concurrently, and produces a user-facing string
// summary of their run.
func (chklst *Checklist) MakeReport() (anyFailed bool, report string) {
	if chklst == nil { // pointers can always be nil
		log.Warn("Nil checklist passed to makeReport. Please report this bug.")
		return
	}
	log.Debug("Making report for " + chklst.Name)
	// run checklist concurrently, reporting errors along the way
	// channels store status information for the report creation
	codes := make(chan int)
	msgs := make(chan string)
	for _, chk := range chklst.Checks {
		log.Info("Running checklist " + chk.ID())
		go func(chk chkutil.Check, codes chan int, msgs chan string) {
			log.Debug("Running check " + chk.ID())
			code, msg, err := chk.Status()
			if err != nil {
				log.WithFields(log.Fields{
					"ID":    chk.ID(),
					"error": err.Error(),
				}).Warn("There was an error running a check")
			}
			codes <- code
			msgs <- msg
		}(chk, codes, msgs)
	}
	// aggregate statistics
	total := len(chklst.Checks)
	passed := 0
	failed := 0
	other := 0
	for _ = range chklst.Checks {
		code := <-codes
		switch code {
		case 0:
			passed++
		case 1:
			failed++
		default:
			other++
		}
	}
	close(codes)
	// output global stats
	report += "↴\nTotal: " + fmt.Sprint(total)
	report += "\nPassed: " + fmt.Sprint(passed)
	report += "\nFailed: " + fmt.Sprint(failed)
	report += "\nOther: " + fmt.Sprint(other)
	// append specific check reports
	for _ = range chklst.Checks {
		if msg := <-msgs; msg != "" {
			report += "\n" + msg
		}
	}
	close(msgs)
	return (failed > 0), report
}

/***************** Checklist JSON structs *****************/

// chkutil.CheckJSON is the check that gets unmarshalled out of the JSON configuration
// file. Because this prohibits type safety
type CheckJSON struct {
	// ignored, included because JSON doesn't have comments
	Notes string
	// matches the ID() of the chkutil.Check object, used in construction process
	ID string
	// the parameters to the check. To be validated upon check construction.
	Parameters []string
}

// chkutil.ChecklsitJSON
type ChecklistJSON struct {
	Name, Notes string
	Checklist   []CheckJSON
}

/***************** Checklist constructors *****************/

// ChecklistFromBytes takes a bytestring of utf8 encoded JSON and turns it into
// a checklist struct. Used by all checklist constructors below. It validates
// the number of parameters that each check has.
func ChecklistFromBytes(data []byte) (chklst Checklist, err error) {
	var chklstJSON ChecklistJSON
	err = json.Unmarshal(data, &chklstJSON)
	if err != nil {
		return chklst, err
	}
	chklst.Name = chklstJSON.Name
	chklst.Notes = chklstJSON.Notes
	// get workers for each check
	out := make(chan chkutil.Check)
	defer close(out)
	for _, chk := range chklstJSON.Checklist {
		go func(chkJSON CheckJSON, out chan chkutil.Check) {
			chkStruct := constructCheck(chkJSON)
			if chkStruct == nil {
				log.Fatal("Check had nil struct: " + chkJSON.ID)
			}
			newChk, err := chkStruct.New(chkJSON.Parameters)
			if err != nil {
				log.WithFields(log.Fields{
					"check":  chkJSON.ID,
					"params": chkJSON.Parameters,
					"error":  err.Error(),
				}).Fatal("Error while constructing check")
			}
			out <- newChk
		}(chk, out)
	}
	// grab all the data from the channel, mutating the checklist
	for _ = range chklstJSON.Checklist {
		chklst.Checks = append(chklst.Checks, <-out)
	}
	if len(chklst.Checks) < 1 {
		log.WithFields(log.Fields{
			"checklist": chklst.Name,
		}).Fatal("Checklist had no checks associated with it!")
	}
	return chklst, nil
}

// ChecklistFromFile reads the file at the path and parses its utf8 encoded json
// data, turning it into a checklist struct.
func ChecklistFromFile(path string) (chklst Checklist, err error) {
	log.Debug("Creating checklist from " + path)
	return ChecklistFromBytes(chkutil.FileToBytes(path))
}

// ChecklistFromStdin reads the stdin pipe and parses its utf8 encoded json
// data, turning it into a checklist struct.
func ChecklistFromStdin() (chklst Checklist, err error) {
	stdinAsBytes := func() (data []byte) {
		data, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Fatal("Couldn't read from stdin")
		} else if len(data) < 1 {
			log.Fatal("Stdin was empty!")
		}
		return data
	}
	log.Debug("Creating checklist from stdin")
	return ChecklistFromBytes(stdinAsBytes())
}

// ChecklistsFromDir reads all of the files in the path and parses their utf8
// encoded json data, turning it into a checklist struct.
func ChecklistsFromDir(dirpath string) (chklsts []Checklist, err error) {
	log.Debug("Creating checklist(s) from " + dirpath)
	paths := chkutil.GetFilesWithExtension(dirpath, ".json")
	// send one checklist per path to the channel
	/*
		out := make(chan Checklist)
		errs := make(chan error)
		for _, path := range paths {
			go func(path string, out chan Checklist, errs chan error) {
				chklst, err := ChecklistFromFile(path)
				out <- chklst
				errs <- err
			}(path, out, errs)
		}
		// get all values from the channel, return them
		for _ = range paths {
			err := <-errs
			if err != nil {
				return chklsts, err
			}
			chklsts = append(chklsts, <-out)
		}
		close(out)
		close(errs)
	*/
	for _, path := range paths {
		chklst, err := ChecklistFromFile(path)
		if err != nil {
			return chklsts, err
		}
		chklsts = append(chklsts, chklst)
	}
	return chklsts, nil
}

// checklistsFromDir reads data retrieved from the URL and parses its utf8
// encoded json data, turning it into a checklist struct. It also caches this
// data at remoteCheckDir, currently "/var/run/distributive/"
func ChecklistFromURL(urlstr string) (chklst Checklist, err error) {
	// ensure temp files dir exists
	log.Debug("Creating/checking remote checklist dir")
	if err := os.MkdirAll(remoteCheckDir, 0775); err != nil {
		log.WithFields(log.Fields{
			"dir":   remoteCheckDir,
			"error": err.Error(),
		}).Warn("Could not create remote check directory")
		remoteCheckDir = "./.remote-checks"
		if err := os.MkdirAll(remoteCheckDir, 0755); err != nil {
			errutil.CouldntWriteError(remoteCheckDir, err)
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
		body := chkutil.URLToBytes(urlstr, true) // secure connection
		log.Debug("Writing remote checklist to cache")
		chkutil.BytesToFile(body, fullpath)
		return ChecklistFromBytes(body)
	}
	log.WithFields(log.Fields{
		"path": fullpath,
	}).Info("Using local copy of remote checklist")
	return ChecklistFromFile(fullpath)
}
