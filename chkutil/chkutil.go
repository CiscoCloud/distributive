package chkutil

import (
	"bytes"
	"crypto/tls"
	"errors"
	"github.com/CiscoCloud/distributive/errutil"
	"github.com/CiscoCloud/distributive/tabular"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// Check is a unified interface for health checks, it defines only the minimal
// necessary behaviors, while allowing each check to parse and store type-safe
// parameters.
type Check interface {
	// New both validates the YAML-provided parameters (list of strings),
	// and parses and stores them in an internal, typed field for later access.
	New(parameters []string) (Check, error)

	// Status returns the status of the check at the instant it is called.
	//
	// msg is a descriptive, human-readable description of the status.
	//
	// code is exit code defining whether or not this check is passing.  0 is
	// considered passing, 1 is failing, with other values reserved for later
	// use.
	Status() (code int, msg string, err error)
}

/// Checks registry

type MakeCheckT func() Check

var registry = map[string]MakeCheckT{}

func Register(name string, check MakeCheckT) {
    lname := strings.ToLower(name)
    registry[lname] = check
}

func LookupCheck(name string) Check {
    lname := strings.ToLower(name)
    if makeCheckFn, ok := registry[lname]; ok {
        return makeCheckFn()
    }
    return nil
}

//// STRING UTILITIES

// CommandOutput returns a string version of the ouput of a given command,
// and reports errors effectively.
func CommandOutput(cmd *exec.Cmd) string {
	out, err := cmd.CombinedOutput()
	outStr := string(out)
	if err != nil {
		errutil.ExecError(cmd, outStr, err)
	}
	return outStr
}

// CommandColumnNoHeader returns a specified column of the output of a command,
// without that column's header. Useful for parsing the output of shell commands,
// which many of the Checks require.
func CommandColumnNoHeader(col int, cmd *exec.Cmd) []string {
	out := CommandOutput(cmd)
	return tabular.GetColumnNoHeader(col, tabular.StringToSlice(out))
}

// SeparateByteUnits: The integer part of a string representing a size unit,
// the unit: b | kb | mb | gb | tb, and an error if applicable.
// 90KB -> (90, kb, nil), 800ads -> (0, "", error)
// NOTE: this doesn't differentiate between kb and kib, and I don't know how
// `free` does.
func SeparateByteUnits(str string) (int, string, error) {
	// getByteUnits takes an arbitrary string containing any crazy possible
	// string representing byte units and returns something normal like kb, mb.
	getByteUnits := func(str string) (string, error) {
		regexps := map[string]*regexp.Regexp{
			// because "bytes" is harder, use it last
			"tb": regexp.MustCompile(`tera(bytes){0,1}|[tT]{1}[iI]{0,1}[bB]{1}`),
			"gb": regexp.MustCompile(`giga(bytes){0,1}|[gG]{1}[iI]{0,1}[bB]{1}`),
			"mb": regexp.MustCompile(`mega(bytes){0,1}|[mM]{1}[iI]{0,1}[bB]{1}`),
			"kb": regexp.MustCompile(`kilo(bytes){0,1}|[kK]{1}[iI]{0,1}[bB]{1}`),
			"b":  regexp.MustCompile(`[^oa]bytes{0,1}|[^kKmMgGtT][bB]{1}`),
		}
		for unit, re := range regexps {
			if re.MatchString(str) {
				return unit, nil
			}
		}
		return "", errors.New("Couldn't extract byte units from string " + str)
	}
	// integerFromString extracts the leftmost integer from str
	integerFromString := func(str string) (int, error) {
		digitRe := regexp.MustCompile(`\d+`)
		intStr := digitRe.FindString(str)
		if intStr == "" {
			return 0, errors.New("Couldn't find integer in string")
		}
		actualInt, err := strconv.ParseInt(intStr, 10, 64)
		if err != nil {
			return 0, errors.New("Couldn't extract integer from string")
		}
		return int(actualInt), nil
	}
	// Warn the user upon failure, but should be shut down later
	unit, err := getByteUnits(str)
	if err != nil {
		return 1, "", err
	}
	scalar, err := integerFromString(str)
	if err != nil {
		return 1, "", err
	}
	return scalar, unit, nil
}

// SubmatchMap returns a map of submatch names to their captures, if any.
// If no matches are found, it returns an empty dict.
// Submatch names are specified using (?P<name>[matchme])
func SubmatchMap(re *regexp.Regexp, str string) (dict map[string]string) {
	dict = make(map[string]string)
	names := re.SubexpNames()
	matches := re.FindStringSubmatch(str)
	if len(names) > 0 && len(matches) > 0 {
		names = names[1:]
		matches = matches[1:]
		for i, name := range names {
			dict[name] = matches[i] // offset from names[1:]
		}
	}
	return dict
}

// IO UTILITIES

// FileToBytes reads a file and handles the error
func FileToBytes(path string) []byte {
	data, err := ioutil.ReadFile(path)
	errutil.CouldntReadError(path, err)
	return data
}

// FileToString reads in a file at a path, handles errors, and returns that file
// as a string
func FileToString(path string) string {
	return string(FileToBytes(path))
}

// FileToLines reads in a file at a path, handles errors, splits it into lines,
// and returns those lines as byte slices
func FileToLines(path string) [][]byte {
	return bytes.Split(FileToBytes(path), []byte("\n"))
}

// BytesToFile writes the given data to the file at the path, and handles errors
func BytesToFile(data []byte, path string) {
	err := ioutil.WriteFile(path, data, 0755)
	if err != nil {
		errutil.CouldntWriteError(path, err)
	}
}

// URLToBytes gets the response from urlstr and returns it as a byte string
// TODO wait on a goroutine w/ timeout, instead of blocking main thread
func URLToBytes(urlstr string, secure bool) []byte {
	// create http client
	transport := &http.Transport{}
	if !secure {
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	client := &http.Client{Transport: transport}
	// get response from URL
	resp, err := client.Get(urlstr)
	if err != nil {
		errutil.CouldntReadError(urlstr, err)
	}
	defer resp.Body.Close()

	// read response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"URL":   urlstr,
			"Error": err.Error(),
		}).Fatal("Bad response, couldn't read body")
	} else if body == nil || bytes.Equal(body, []byte{}) {
		log.WithFields(log.Fields{
			"URL": urlstr,
		}).Warn("Body of response was empty")
	}
	return body
}

// GetFilesWithExtension returns the paths to all the files in the given dir
// that end with the given file extension (with or without dot)
func GetFilesWithExtension(path string, ext string) (paths []string) {
	finfos, err := ioutil.ReadDir(path) // list of os.FileInfo
	if err != nil {
		errutil.CouldntReadError(path, err)
	}
	for _, finfo := range finfos {
		name := finfo.Name()
		if strings.HasSuffix(name, ext) {
			// TODO path.Join these suckers
			paths = append(paths, path+"/"+name)
		}
	}
	return paths
}
