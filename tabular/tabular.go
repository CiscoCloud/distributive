// tabular is a package that simplifies the reading of tabular data.
// It is especially intended for use when finding certain values in certain
// rows or columns, but can be useful for simply handling lists of strings.
package tabular

import (
	"fmt"
	"regexp"
	"strings"
)

// Table, Column, and Row are just type synonyms that make it easier to reason
// about some of the below functions
type Table [][]string
type Column []string
type Row []string

func PrintTable(tab Table) {
	for _, row := range tab {
		fmt.Println(row)
	}
}

// separateString is an abstraction of stringToSlice that takes two kinds of
// separators, and splits a string into a 2D slice based on those separators
func SeparateString(rowSep *regexp.Regexp, colSep *regexp.Regexp, str string) (output Table) {
	lines := rowSep.Split(str, -1)
	for _, line := range lines {
		output = append(output, colSep.Split(line, -1))
	}
	return output
}

// stringToSlice takes in a string and returns a 2D slice of its output,
// separated on whitespace and newlines
func StringToSlice(str string) (output Table) {
	rowSep := regexp.MustCompile("\n+")
	colSep := regexp.MustCompile("\\s+")
	return SeparateString(rowSep, colSep, str)
}

// stringToSliceMultispace is for commands that have spaces within their columns
// but more than one space between columns
func StringToSliceMultispace(str string) (output Table) {
	rowSep := regexp.MustCompile("\n+")
	colSep := regexp.MustCompile("\\s{2,}")
	return SeparateString(rowSep, colSep, str)
}

// getColumn isolates the entries of a single column from a 2D slice
func GetColumn(col int, slice [][]string) (column Column) {
	for _, line := range slice {
		if len(line) > col {
			column = append(column, line[col])
		}
	}
	return column
}

// getColumnNoHeader safely removes the first element from a column
func GetColumnNoHeader(col int, tab Table) Column {
	column := GetColumn(col, tab)
	if len(column) < 1 {
		return column
	}
	return column[1:]
}

// stringPredicate is a function that filters a list of strings
type StringPredicate func(str string) bool

// anySatisfies checks to see whether any string in a given slice satisfies the
// provided stringPredicate
func AnySatisfies(pred StringPredicate, slice []string) bool {
	for _, sliceString := range slice {
		if pred(sliceString) {
			return true
		}
	}
	return false
}

// strIn checks to see if a given string is in a slice of strings
func StrIn(str string, slice []string) bool {
	pred := func(strx string) bool { return (strx == str) }
	return AnySatisfies(pred, slice)
}

// strContainedIn works like strIn, but checks for substring containing rather
// than whole string equality.
func StrContainedIn(str string, slice []string) bool {
	pred := func(strx string) bool { return strings.Contains(strx, str) }
	return AnySatisfies(pred, slice)
}

// reIn is like strIn, but matches regexps instead
func ReIn(re *regexp.Regexp, slice []string) bool {
	pred := func(strx string) bool { return re.MatchString(strx) }
	return AnySatisfies(pred, slice)
}
