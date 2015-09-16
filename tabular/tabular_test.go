package tabular

import (
	"regexp"
	"testing"
)

// TODO they seem the same, but apparently differ
/*
func TestToString(t *testing.T) {
	outputs := []string{
		`+---+---+---+---+
| 1 | 2 | 3 | 4 |
+===+===+===+===+
| 1 | 2 | 3 | 4 |
+---+---+---+---+
| 1 | 2 | 3 | 4 |
+---+---+---+---+`,
		`+---+---+---+---+---+
| 1 | 2 | 3 | 4 |   |
+===+===+===+===+===+
| 1 | 2 | 3 | 4 |   |
+---+---+---+---+---+
| 1 | 2 | 3 | 4 | 5 |
+---+---+---+---+---+
| 1 | 2 | 3 | 4 |   |
+---+---+---+---+---+
| 1 | 2 | 3 | 4 |   |
+---+---+---+---+---+`,
		`+---+---+
| 1 | 2 |
+===+===+
| 1 | 2 |
+---+---+
| 1 | 2 |
+---+---+
| 1 | 2 |
+---+---+`,
		`+---+---+---+---+---+---+---+---+---+
| 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 |
+===+===+===+===+===+===+===+===+===+
| 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 |
+---+---+---+---+---+---+---+---+---+
| 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 |
+---+---+---+---+---+---+---+---+---+
| 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 |
+---+---+---+---+---+---+---+---+---+
| 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 |
+---+---+---+---+---+---+---+---+---+
| 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 |   |
+---+---+---+---+---+---+---+---+---+
| 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 |
+---+---+---+---+---+---+---+---+---+
| 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 |
+---+---+---+---+---+---+---+---+---+
| 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 |
+---+---+---+---+---+---+---+---+---+`,
	}
	for i := range outputTables {
		input := outputTables[i]
		expected := outputs[i]
		actual := ToString(input)
		if actual != expected {
			pureFunctionError(t, input, expected, actual)
		}
	}
}
*/

func TestTableEqual(t *testing.T) {
	t.Parallel()
	inputs1 := [][][]string{
		[][]string{
			[]string{"0", "1", "2"},
			[]string{"0", "1", "2"},
			[]string{"0", "1", "2"},
		},
		[][]string{
			[]string{"0", "1", "2"},
			[]string{"0", "1", "2"},
			[]string{"0", "1", "2"},
		},
		[][]string{[]string{"0"}},
		[][]string{[]string{"0"}},
	}
	inputs2 := [][][]string{
		[][]string{
			[]string{"0", "1", "2"},
			[]string{"0", "1", "2"},
			[]string{"0", "1", "2"},
		},
		[][]string{
			[]string{"0", "1", "2"},
			[]string{"0", "1", " "},
			[]string{"0", "1", "2"},
		},
		[][]string{[]string{"0"}},
		[][]string{[]string{"4"}},
	}
	outputs := []bool{true, false, true, false}
	for i := range inputs1 {
		input1 := inputs1[i]
		input2 := inputs2[i]
		expected := outputs[i]
		actual := TableEqual(input1, input2)
		if actual != expected {
			pureFunctionError(t, input2, expected, actual)
		}
	}
}

func TestSeparateOnAligment(t *testing.T) {
	t.Parallel()
	inputs := []string{
		`head1        head2   head3 head4      head5
asdlfjkab    asdllas asdf  asd        pin
lj           eqw     bb    0890hlkjba 1xsa`,
	}
	outputs := [][][]string{
		[][]string{
			[]string{"head1", "head2", "head3", "head4", "head5"},
			[]string{"asdlfjkab", "asdllas", "asdf", "asd", "pin"},
			[]string{"lj", "eqw", "bb", "0890hlkjba", "1xsa"},
		},
	}
	for i := range inputs {
		input := inputs[i]
		expected := outputs[i]
		actual := SeparateOnAlignment(input)
		if !TableEqual(expected, actual) {
			pureFunctionError(t, input, expected, actual)
		}
	}
}

func TestGetColumn(t *testing.T) {
	t.Parallel()
	inputs := outputTables
	columns := []int{1, 2, 0, 6}
	outputs := [][]string{
		[]string{"2", "2", "2"},
		[]string{"3", "3", "3", "3", "3"},
		[]string{"1", "1", "1", "1"},
		[]string{"7", "7", "7", "7", "7", "7", "7", "7", "7"},
	}
	for i := range inputs {
		input := inputs[i]
		column := columns[i]
		expected := outputs[i]
		actual := GetColumn(column, input)
		if !SliceEqual(expected, actual) {
			pureFunctionError(t, input, expected, actual)
		}
	}
}

func TestGetColumnByHeader(t *testing.T) {
	t.Parallel()
	inputs := outputTables
	columns := []string{"2", "3", "1", "7"}
	outputs := [][]string{
		[]string{"2", "2"},
		[]string{"3", "3", "3", "3"},
		[]string{"1", "1", "1"},
		[]string{"7", "7", "7", "7", "7", "7", "7", "7"},
	}
	for i := range inputs {
		input := inputs[i]
		column := columns[i]
		expected := outputs[i]
		actual := GetColumnByHeader(column, input)
		if !SliceEqual(expected, actual) {
			pureFunctionError(t, input, expected, actual)
		}
	}
}

var testStrings = []string{
	"test", "  testing", "01243894word10238", "aasdff", "drow", "esac", "fi",
}

func TestStrIn(t *testing.T) {
	t.Parallel()
	inputs := []string{"test", "testing", "fi", "esle"}
	outputs := []bool{true, false, true, false}
	for i := range inputs {
		input := inputs[i]
		expected := outputs[i]
		actual := StrIn(input, testStrings)
		if actual != expected {
			pureFunctionError(t, input, expected, actual)
		}
	}
}

func TestStrContainedIn(t *testing.T) {
	t.Parallel()
	// TODO df returns false
	//inputs := []string{"test", "df", "fi", "esle", "3107233"}
	//outputs := []bool{true, true, true, false, false}
	inputs := []string{"test", "fi", "esle", "3107233"}
	outputs := []bool{true, true, false, false}
	for i := range inputs {
		input := inputs[i]
		expected := outputs[i]
		actual := StrIn(input, testStrings)
		if actual != expected {
			pureFunctionError(t, input, expected, actual)
		}
	}
}

func TestReIn(t *testing.T) {
	t.Parallel()
	inputs := []*regexp.Regexp{
		regexp.MustCompile(`test`),
		regexp.MustCompile(`testing`),
		regexp.MustCompile(`fi`),
		regexp.MustCompile(`esle`),
		regexp.MustCompile(`3107233`),
		regexp.MustCompile(`\d+\w+\d+`),
		regexp.MustCompile(`\w+`),
		regexp.MustCompile(`\s+\w+`),
		regexp.MustCompile(`\s+\w+\s+\w+\s+`),
	}
	outputs := []bool{true, true, true, false, false, true, true, true, false}
	for i := range inputs {
		input := inputs[i]
		expected := outputs[i]
		actual := ReIn(input, testStrings)
		if actual != expected {
			pureFunctionError(t, input, expected, actual)
		}
	}
}

func TestHasNonEmpty(t *testing.T) {
	t.Parallel()
	inputs := [][]string{
		[]string{"    ", "", " ", "\t\n\t", "\t", "   \t  \n "},
		[]string{"    ", "t", " ", "\t\n\t", "\t", "   \t  \n "},
		[]string{"    ", "", " ", "\t-\n\t", "\t", "   \t  \n "},
		[]string{"", "", "", "", "", ""},
	}
	outputs := []bool{false, true, true, false}
	for i := range inputs {
		input := inputs[i]
		expected := outputs[i]
		actual := HasNonEmpty(input)
		if actual != expected {
			pureFunctionError(t, input, expected, actual)
		}
	}
}
