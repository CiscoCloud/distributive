package main

import (
	"bytes"
	"fmt"
	"github.com/CiscoCloud/distributive/tabular"
	"io/ioutil"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// Worker is the type of function that takes a list of string params and returns
// an error code and an exit message to be printed to stdout.
// Generally, if exitCode == 0, exitMessage == "".
type Worker func(parameters []string) (exitCode int, exitMessage string)

//// STRING UTILITIES

// commandColumnNoHeader returns a specified column of the output of a command,
// without that column's header. Useful for parsing the output of shell commands,
// which many of the Checks require.
// TODO for some reason, this + route -n doesn't work with probabalistic.
func commandColumnNoHeader(col int, cmd *exec.Cmd) []string {
	out, err := cmd.CombinedOutput()
	outstr := string(out)
	if strings.Contains(outstr, "permission denied") {
		log.Fatal("Permission denied when running: " + cmd.Path)
	} else if err != nil {
		execError(cmd, outstr, err)
	}
	return tabular.GetColumnNoHeader(col, tabular.StringToSlice(string(out)))
}

//// ERROR UTILITIES

// pathError is an abstraction of couldntReadError and couldntWriteError
func pathError(path string, err error, read bool) {
	// is it a read or write error?
	readOrWrite := "write"
	if read {
		readOrWrite = "read"
	}

	if err != nil {
		msg := "Couldn't " + readOrWrite + " file:"
		msg += "\n\tPath: " + path
		msg += "\n\tError: " + err.Error()
		if path == "" {
			msg += "\n\t No path specified. Try using the -f flag."
		}
		log.Fatal(msg)
	}
}

// couldntWriteError logs.Fatal an error relating to writing a file
func couldntWriteError(path string, err error) {
	pathError(path, err, false)
}

// couldntReadError logs.Fatal an error related to reading a file
func couldntReadError(path string, err error) {
	pathError(path, err, true)
}

// genericError is a general error where the requested variable was not found in
// a given list of variables. This is pure DRY.
func genericError(msg string, name string, actual []string) (exitCode int, exitMessage string) {
	// with low verbosity, we don't need to specify the check in too much detail
	if verbosity <= minVerbosity {
		return 1, msg
	}
	msg += ":\n\tSpecified: " + name
	// this is the number of list items to be output at verbosities strictly
	// in between maximum and minimum verbosity.
	lengthThreshold := 10 * (verbosity + 1)
	if len(actual) == 1 {
		msg += "\n\tActual: " + fmt.Sprint(actual[0])
	} else if verbosity >= maxVerbosity || len(actual) < lengthThreshold {
		msg += "\n\tActual: " + fmt.Sprint(actual)
	} else {
		msg += "\n\tActual (truncated - increase verbosity to see more): "
		msg += fmt.Sprint(actual[1:lengthThreshold])
	}
	return 1, msg
}

func execError(cmd *exec.Cmd, out string, err error) {
	if err != nil {
		msg := "Failed to execute command:"
		msg += "\n\tCommand: " + fmt.Sprint(cmd.Args)
		msg += "\n\tPath: " + cmd.Path
		msg += "\n\tOutput: " + out
		msg += "\n\tError: " + err.Error()
	}
}

// IO UTILITIES

// parseUserRegex either returns a regex from a string, or displays an
// appropriate error to the user
func parseUserRegex(regexString string) *regexp.Regexp {
	re, err := regexp.Compile(regexString)
	if err != nil {
		msg := "Bad configuration - couldn't parse golang regexp:"
		msg += "\n\tRegex text: " + regexString
		msg += "\n\tError: " + err.Error()
		log.Fatal(msg)
	}
	return re
}

// fileToBytes reads a file and handles the error
func fileToBytes(path string) []byte {
	data, err := ioutil.ReadFile(path)
	couldntReadError(path, err)
	return data
}

// fileToString reads in a file at a path, handles errors, and returns that file
// as a string
func fileToString(path string) string {
	return string(fileToBytes(path))
}

// fileToLines reads in a file at a path, handles errors, splits it into lines,
// and returns those lines as byte slices
func fileToLines(path string) [][]byte {
	return bytes.Split(fileToBytes(path), []byte("\n"))
}

// parseMyInt parses an int or logs the error
func parseMyInt(str string) int {
	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		msg := "Probable configuration error - could not parse integer: "
		msg += "\n\tGiven string: " + str
		log.Fatal(msg)
	}
	return int(i)
}

// getFilesWithExtension returns the paths to all the files in the given dir
// that end with the given file extension (with or without dot)
func getFilesWithExtension(path string, ext string) (paths []string) {
	finfos, err := ioutil.ReadDir(path) // list of os.FileInfo
	if err != nil {
		couldntReadError(path, err)
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

/*
// byteSliceToStrSlice takes a slice of byte slices and returns the equivalent
// slice of strings
func byteSliceToStrSlice(byteSlice [][]byte) (strSlice []string) {
	for _, word := range byteSlice {
		strSlice = append(strSlice, string(word))
	}
	return strSlice
}
*/
