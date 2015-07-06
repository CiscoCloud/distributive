// Package tabular simplifies the reading of tabular data.  It is especially
// intended for use when finding certain values in certain rows or columns, but
// can be useful for simply handling lists of strings.
package tabular

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// Table is a 2D slice of strings, for representing tabular data as cells, in
// rows and columns.
type Table [][]string

// Column is a slice of strings, a vertical slice of a Table.
type Column []string

// Row is one element of a Table, a horizontal slice.
type Row []string

// ToString returns a table ([][]string) as a nicely formatted string
func ToString(table Table) string {
	// maxInt returns the maximum of a list of integers
	maxInt := func(slc []int) (max int) {
		for _, i := range slc {
			if i > max {
				max = i
			}
		}
		return max
	}
	// longestInColumn returns the length of the longest string in the given
	// slice
	longestInColumn := func(col Column) int {
		strLengths := func(slc []string) (lengths []int) {
			for _, str := range slc {
				lengths = append(lengths, utf8.RuneCountInString(str))
			}
			return lengths
		}
		return maxInt(strLengths(col))
	}
	rows := []string{}
	for _, row := range table {
		rowStr := "\n"
		for i, item := range row {
			columnWidth := longestInColumn(GetColumn(i, table))
			rowStr = rowStr + "| " + padString(item, " ", columnWidth) + " "
		}
		rows = append(rows, rowStr+"|")
	}
	maxLength := 0
	for _, row := range rows {
		length := utf8.RuneCountInString(row)
		if length > maxLength {
			maxLength = length
		}
	}
	divider := repeatString("-", maxLength)
	return divider + strings.Join(rows, "\n"+divider) + "\n" + divider
}

// SeparateString is an abstraction of stringToSlice that takes two kinds of
// separators, and splits a string into a 2D slice based on those separators
func SeparateString(rowSep *regexp.Regexp, colSep *regexp.Regexp, str string) (output Table) {
	lines := rowSep.Split(str, -1)
	for _, line := range lines {
		row := colSep.Split(line, -1)
		if len(row) > 0 && HasNonEmpty(row) {
			output = append(output, row)
		}
	}
	return output
}

// StringToSlice takes in a string and returns a 2D slice of its output,
// separated on whitespace and newlines
// TODO: this should be depreciated by probabalisticSplit
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

// HasNonEmpty checks to see if there is a single string with a non-whitespace
// character in the list
func HasNonEmpty(slice []string) bool {
	// regexp matches: beginning of string, >0 non-whitespace char, eof
	return ReIn(regexp.MustCompile("\\S+"), slice)
}
