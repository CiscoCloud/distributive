// Package tabular simplifies the reading of tabular data.  It is especially
// intended for use when finding certain values in certain rows or columns, but
// can be useful for simply handling lists of strings.
package tabular

import (
	log "github.com/Sirupsen/logrus"
	"regexp"
	"strings"
	"unicode/utf8"
)

var rowSep = regexp.MustCompile("\n+")

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
	// columnWidths gets the lengths of the widest strings in each column of a
	// table
	columnWidths := func(table Table) []int {
		columnWidths := []int{}
		for _, row := range table {
			for i, item := range row {
				itemLength := utf8.RuneCountInString(item)
				if len(columnWidths) <= i {
					columnWidths = append(columnWidths, itemLength)
				} else if columnWidths[i] < itemLength {
					columnWidths[i] = itemLength
				}
			}
		}
		return columnWidths
	}

	rows := []string{}
	for _, row := range table {
		rowStr := "\n| "
		for i, item := range row {
			columnWidth := longestInColumn(GetColumn(i, table))
			rowStr = rowStr + padString(item, " ", columnWidth) + " |"
		}
		rows = append(rows, rowStr)
	}
	maxLength := 0
	for _, row := range rows {
		length := utf8.RuneCountInString(row)
		if length > maxLength {
			maxLength = length
		}
	}
	// construct RST style row dividers
	normalDivider := "+"
	headerDivider := "+"
	for _, width := range columnWidths(table) {
		normalDivider += repeatString("-", width) + "--+"
		headerDivider += repeatString("=", width) + "==+"
	}
	// intersperse those dividers
	separatedRows := []string{}
	for i, row := range rows {
		divider := normalDivider
		if i == 0 {
			divider = headerDivider
		}
		separatedRows = append(separatedRows, row, "\n"+divider)
	}
	return normalDivider + strings.Join(separatedRows, "")
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

// SeparateOnAlignment splits a table based on the indicies of its headers,
// assuming all columns are left-aligned and all headers are separated by
// whitespace
func SeparateOnAlignment(str string) (table Table) {
	// wordAfterIndex gets the first whitespace-delimited word of a string
	// that occurs after the given index
	wordAfterIndex := func(i int, str string) string {
		if i < 0 {
			log.WithFields(log.Fields{
				"index": i,
			}).Fatal("Internal error: negative index passed to wordAfterIndex")
		}
		msg := "Couldn't get wordAfterIndex, "
		if len(str) < i {
			log.WithFields(log.Fields{
				"index": i,
				"str":   str,
			}).Warn(msg + "string was too short")
			return ""
		}
		fields := strings.Fields(str[i:])
		if len(fields) < 1 {
			log.WithFields(log.Fields{
				"index": i,
				"str":   str,
			}).Warn(msg + "string had only whitespace")
		}
		return strings.TrimSpace(fields[0])
	}
	// getHeaders returns a list of table headers from a unseparated string,
	// assumed to be separated by a unicode.IsSpace character
	getHeaders := func(str string) []string {
		rows := regexp.MustCompile("\n+").Split(str, -1)
		if len(rows) < 2 {
			log.WithFields(log.Fields{
				"row regexp": rowSep.String(),
				"length":     len(rows),
				"string":     str,
			}).Fatal("Couldn't split table based on headers")
		}
		headers := strings.Fields(rows[0])
		if len(headers) < 1 {
			log.WithFields(log.Fields{
				"string": str,
			}).Fatal("Couldn't get headers of table, row 0 was only whitespace")
		}
		return headers
	}
	rows := regexp.MustCompile("\n+").Split(str, -1)
	headers := getHeaders(str)
	headerIndicies := IndiciesOf(headers, rows[0])
	// error out on negative indicies, headers should always all be found.
	for i, index := range headerIndicies {
		if index < 0 {
			header := ""
			if len(headers) > i {
				header = headers[i]
			}
			log.WithFields(log.Fields{
				"index":    index,
				"indicies": headerIndicies,
				"header":   header,
				"row":      rows[0],
			}).Fatal("Internal error: negative header index")
		}
	}
	// split string based on indicies
	//table = append(table, headers)
	for i, row := range rows {
		table = append(table, []string{}) // make a new row
		for _, index := range headerIndicies {
			table[i] = append(table[i], wordAfterIndex(index, row))
		}
	}
	return table
}

// StringToSlice takes in a string and returns a 2D slice of its output,
// separated on whitespace and newlines
// TODO: this should be depreciated by probabalisticSplit
func StringToSlice(str string) (output Table) {
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
	if len(column) < 1 {
		log.WithFields(log.Fields{
			"column": column,
			"length": len(column),
		}).Fatal("Column too short to remove header from")
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
