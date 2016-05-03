package checklists

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/CiscoCloud/distributive/chkutil"
	"github.com/CiscoCloud/distributive/errutil"
	log "github.com/Sirupsen/logrus"
	"github.com/ghodss/yaml"
)

// where remote checks are downloaded to
var remoteCheckDir = "/var/run/distributive/"

/***************** Checklist type *****************/

// Checklist is a struct that provides a concise way of thinking about doing
// several checks and then returning some kind of output.
type Checklist struct {
	Name   string
	Checks []chkutil.Check // list of chkutil.Checks to run
	Origin string          // where did it come from?
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
		log.Info("Running check " + chk.ID())
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

/***************** Checklist YAML structs *****************/

// CheckYAML is the check that gets unmarshalled out of the YAML configuration
// file.
type CheckYAML struct {
	// matches the ID() of the chkutil.Check object, used in construction process
	ID string `json:"id"`
	// the parameters to the check. To be validated upon check construction.
	Parameters []string `json:"parameters"`
}

// ChecklistYAML is the representation of a checklist that's parsed from the
// YAML, before being converted into an internal representation.
type ChecklistYAML struct {
	Name      string      `json:"name"`
	Checklist []CheckYAML `json"checklist"`
}

/***************** Checklist constructors *****************/

// FromBytes takes a bytestring of utf8 encoded YAML and turns it into
// a checklist struct. Used by all checklist constructors below. It validates
// the number of parameters that each check has.
func FromBytes(data []byte) (chklst Checklist, err error) {
	var chklstYAML ChecklistYAML
	err = yaml.Unmarshal(data, &chklstYAML)
	if err != nil {
		return chklst, err
	}
	chklst.Name = chklstYAML.Name
	// get workers for each check
	out := make(chan chkutil.Check)
	defer close(out)
	for _, chk := range chklstYAML.Checklist {
		go func(chkYAML CheckYAML, out chan chkutil.Check) {
			chkStruct := constructCheck(chkYAML)
			if chkStruct == nil {
				log.Fatal("Check had nil struct: " + chkYAML.ID)
			}
			newChk, err := chkStruct.New(chkYAML.Parameters)
			if err != nil {
				log.WithFields(log.Fields{
					"check":  chkYAML.ID,
					"params": chkYAML.Parameters,
					"error":  err.Error(),
				}).Fatal("Error while constructing check")
			}
			out <- newChk
		}(chk, out)
	}
	// grab all the data from the channel, mutating the checklist
	for _ = range chklstYAML.Checklist {
		chklst.Checks = append(chklst.Checks, <-out)
	}
	if len(chklst.Checks) < 1 {
		log.WithFields(log.Fields{
			"checklist": chklst.Name,
		}).Fatal("Checklist had no checks associated with it!")
	}
	return chklst, nil
}

// FromFile reads the file at the path and parses its utf8 encoded yaml
// data, turning it into a checklist struct.
func FromFile(path string) (chklst Checklist, err error) {
	log.Debugf("Creating checklist from %s", path)
	return FromBytes(chkutil.FileToBytes(path))
}

// FromStdin reads the stdin pipe and parses its utf8 encoded yaml
// data, turning it into a checklist struct.
func FromStdin() (chklst Checklist, err error) {
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
	return FromBytes(stdinAsBytes())
}

// FromDirectory reads all of the files in the path and parses their utf8
// encoded yaml data, turning it into a checklist struct.
func FromDirectory(dirpath string) (chklsts []Checklist, err error) {
	log.Debug("Creating checklist(s) from " + dirpath)
	paths := chkutil.GetFilesWithExtension(dirpath, ".yaml")
	paths = append(paths, chkutil.GetFilesWithExtension(dirpath, ".yml")...)
	paths = append(paths, chkutil.GetFilesWithExtension(dirpath, ".json")...)
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
		chklst, err := FromFile(path)
		if err != nil {
			return chklsts, err
		}
		chklsts = append(chklsts, chklst)
	}
	return chklsts, nil
}

// FromURL reads data retrieved from the URL and parses its utf8
// encoded yaml data, turning it into a checklist struct. It also optionally
// caches this data at remoteCheckDir, currently "/var/run/distributive/".
func FromURL(urlstr string, cache bool) (chklst Checklist, err error) {
	log.Debug("Creating/checking remote checklist dir")
	if err := os.MkdirAll(remoteCheckDir, 0775); err != nil {
		log.WithFields(log.Fields{
			"dir":   remoteCheckDir,
			"error": err.Error(),
		}).Warn("Could not create remote check directory")
		// attempt a more local directory
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
	filename := pathSanitize(urlstr) + ".yaml"
	fullpath := filepath.Join(remoteCheckDir, filename)

	// only create it if it doesn't exist or we aren't using the cached copy
	if _, err := os.Stat(fullpath); err != nil || !cache {
		log.Info("Fetching remote checklist")
		body := chkutil.URLToBytes(urlstr, true) // secure connection
		log.Debug("Writing remote checklist to cache")
		chkutil.BytesToFile(body, fullpath)
		return FromBytes(body)
	}
	log.WithFields(log.Fields{
		"path": fullpath,
	}).Info("Using local copy of remote checklist")
	return FromFile(fullpath)
}
