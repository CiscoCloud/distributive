package main

import (
	"log"
	"os/exec"
	"regexp"
	"strings"
)

// Thunk is the type of function that runs without parameters and returns
// an error code and an exit message to be printed to stdout.
// Generally, if exitCode == 0, exitMessage == "".
type Thunk func() (exitCode int, exitMessage string)

/*
func fileToLines(path string) [][]byte {
	data, err := ioutil.ReadFile(path)
	fatal(err)
	return bytes.Split(data, []byte("\n"))
}
*/

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
	}
	fatal(err)
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
