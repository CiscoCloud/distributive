// Package tabular simplifies the reading of tabular data.  It is especially
// intended for use when finding certain values in certain rows or columns, but
// can be useful for simply handling lists of strings.
package tabular

import (
	"fmt"
	"regexp"
	"strings"
)

// Table is a 2D slice of strings, for representing tabular data as cells, in
// rows and columns.
type Table [][]string

// Column is a slice of strings, a vertical slice of a Table.
type Column []string

// Row is one element of a Table, a horizontal slice.
type Row []string

// PrintTable prints a table ([][]string) row by row.
func PrintTable(tab Table) {
	for _, row := range tab {
		fmt.Println(row)
	}
}

// SeparateString is an abstraction of stringToSlice that takes two kinds of
// separators, and splits a string into a 2D slice based on those separators
func SeparateString(rowSep *regexp.Regexp, colSep *regexp.Regexp, str string) (output Table) {
	lines := rowSep.Split(str, -1)
	for _, line := range lines {
		output = append(output, colSep.Split(line, -1))
	}
	return output
}

// StringToSlice takes in a string and returns a 2D slice of its output,
// separated on whitespace and newlines
func StringToSlice(str string) (output Table) {
	rowSep := regexp.MustCompile("\n+")
	colSep := regexp.MustCompile("\\s+")
	return SeparateString(rowSep, colSep, str)
}

// GetColumn isolates the entries of a single column from a 2D slice, specified
// by the column number (counting from zero on the left).
func GetColumn(col int, slice [][]string) (column Column) {
	for _, line := range slice {
		if len(line) > col {
			column = append(column, line[col])
		}
	}
	return column
}

// GetColumnNoHeader safely removes the first element from a column.
func GetColumnNoHeader(col int, tab Table) Column {
	column := GetColumn(col, tab)
	if len(column) < 1 {
		return column
	}
	return column[1:]
}

// GetColumnByHeader returns the body of a column with a header that is equal
// to name (ignoring case differences). It is for developer ease and
// future-proofing, as it doesn't rely on an index.
func GetColumnByHeader(name string, tab Table) (column Column) {
	// strIndexOf returns the index of a string in its slice
	strIndexOf := func(str string, slc []string) int {
		for i, sliceStr := range slc {
			if strings.EqualFold(sliceStr, str) {
				return i
			}
		}
		return -1
	}
	// if the table's empty, the column will be too
	if len(tab) < 1 {
		return column
	}
	// ensure that header is present in the headers
	headerCol := -1
	if headerCol = strIndexOf(name, tab[0]); headerCol == -1 {
		return column
	}
	return GetColumnNoHeader(headerCol, tab)
}

// StringPredicate is a function that filters a list of strings
type StringPredicate func(str string) bool

// AnySatisfies checks to see whether any string in a given slice satisfies the
// provided StringPredicate.
func AnySatisfies(pred StringPredicate, slice []string) bool {
	for _, sliceString := range slice {
		if pred(sliceString) {
			return true
		}
	}
	return false
}

// StrIn checks to see if a given string is in a slice of strings.
func StrIn(str string, slice []string) bool {
	pred := func(strx string) bool { return (strx == str) }
	return AnySatisfies(pred, slice)
}

// StrContainedIn works like strIn, but checks for substring containing rather
// than whole string equality.
func StrContainedIn(str string, slice []string) bool {
	pred := func(strx string) bool { return strings.Contains(strx, str) }
	return AnySatisfies(pred, slice)
}

// ReIn is like StrIn or StrContainedIn, but matches regexps instead.
func ReIn(re *regexp.Regexp, slice []string) bool {
	pred := func(strx string) bool { return re.MatchString(strx) }
	return AnySatisfies(pred, slice)
}
