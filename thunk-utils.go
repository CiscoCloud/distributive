package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"regexp"
	"strings"
)

// Thunk is the type of function that runs without parameters and returns
// an error code and an exit message to be printed to stdout.
// Generally, if exitCode == 0, exitMessage == "".
type Thunk func() (exitCode int, exitMessage string)

// separateString is an abstraction of stringToSlice that takes two kinds of
// separators, and splits a string into a 2D slice based on those separators
func separateString(rowSep *regexp.Regexp, colSep *regexp.Regexp, str string) (output [][]string) {
	lines := rowSep.Split(str, -1)
	for _, line := range lines {
		output = append(output, colSep.Split(line, -1))
	}
	return output
}

// stringToSlice takes in a string and returns a 2D slice of its output,
// separated on whitespace and newlines
func stringToSlice(str string) (output [][]string) {
	rowSep := regexp.MustCompile("\n+")
	colSep := regexp.MustCompile("\\s+")
	return separateString(rowSep, colSep, str)
}

// getColumn isolates the entries of a single column from a 2D slice
// it is currently only used by PPA for reading /etc/apt/sources.list
func getColumn(col int, slice [][]string) (column []string) {
	for _, line := range slice {
		if len(line) > col {
			column = append(column, line[col])
		}
	}
	return column
}

// getColumnNoHeader safely removes the first element from a column
func getColumnNoHeader(col int, slice [][]string) []string {

	column := getColumn(col, slice)
	if len(column) < 1 {
		return column
	}
	return column[1:]
}

// commandColumnNoHeader returns a specified column of the output of a command,
// without that column's header. Useful for parsing the output of shell commands,
// which many of the Checks require.
func commandColumnNoHeader(col int, cmd *exec.Cmd) []string {
	out, err := cmd.CombinedOutput()
	outstr := string(out)
	if strings.Contains(outstr, "permission denied") {
		log.Fatal("Permission denied when running: " + cmd.Path)
	} else if err != nil {
		msg := "Error while executing command:"
		msg += "\n\tCommand: " + cmd.Path
		msg += "\n\tArguments: " + fmt.Sprint(cmd.Args)
		msg += "\n\tError: " + err.Error()
		log.Fatal(msg)
	}
	return getColumnNoHeader(col, stringToSlice(string(out)))
}

// strIn checks to see if a given string is in a slice of strings
func strIn(str string, slice []string) bool {
	for _, sliceString := range slice {
		if str == sliceString {
			return true
		}
	}
	return false
}

// couldntReadError logs.Fatal an error related to reading a file
func couldntReadError(path string, err error) {
	if err != nil {
		msg := "Couldn't read file:"
		msg += "\n\tPath: " + path
		msg += "\n\tError: " + err.Error()
	}
}

// fileToBytes reads a file and handles the error
// TODO run through all checks, use this where appropriate
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

// notInError is a general error where the requested variable was not found in
// a given list of variables. This is pure DRY.
// TODO make the output of this function depend on a verbosity level, as made
// a global state variable when command line flags are parsed
func notInError(msg string, name string, available []string) (exitCode int, exitMessage string) {
	msg += "\n\tSpecified: " + name
	msg += "\n\tActual: "
	if len(available) == 1 {
		msg += fmt.Sprint(available[0])
	}
	msg += fmt.Sprint(available)
	return 1, msg
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
/*
// anyContains checks to see whether any of the strings in the given slice
// contain the substring str
func anyContains(str string, slice []string) bool {
	for _, sliceString := range slice {
		if strings.Contains(sliceString, str) {
			return true
		}
	}
	return false
}
*/
