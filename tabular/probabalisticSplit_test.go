package tabular

import (
	"fmt"
	"regexp"
	"testing"
)

func pureFunctionError(t *testing.T, input interface{}, expected interface{}, actual interface{}) {
	msg := "Actual output did not match expected"
	msg += "\n\tInput: " + fmt.Sprint(input)
	msg += "\n\tExpected: " + fmt.Sprint(expected)
	msg += "\n\tActual: " + fmt.Sprint(actual)
	t.Error(msg)
}

func TestMeanInt(t *testing.T) {
	t.Parallel()
	inputs := [][]int{
		[]int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		[]int{17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17},
		[]int{1, 2, 3},
		[]int{1, 2, 3, 4, 5, 6},
		[]int{1500, 3500},
		[]int{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		[]int{-17, -17, -17, -17, -17, -17, -17, -17, -17, -17, -17, -17},
		[]int{-1, -2, -3},
		[]int{-1, -2, -3, -4, -5, -6},
		[]int{-1500, -3500},
		[]int{-1, 1},
		[]int{-0, 0},
		[]int{0},
	}
	outputs := []float64{
		1, 17, 2, 3.5, 2500, -1, -17, -2, -3.5, -2500, 0, 0, 0,
	}
	for i := range inputs {
		input := inputs[i]
		expected := outputs[i]
		actual := meanInt(input)
		if expected != actual {
			pureFunctionError(t, input, expected, actual)
		}
	}
}

func TestMeanFloat(t *testing.T) {
	t.Parallel()
	inputs := [][]float64{
		[]float64{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		[]float64{17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17},
		[]float64{1, 2, 3},
		[]float64{1, 2, 3, 4, 5, 6},
		[]float64{1500, 3500},
		[]float64{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		[]float64{-17, -17, -17, -17, -17, -17, -17, -17, -17, -17, -17, -17},
		[]float64{-1, -2, -3},
		[]float64{-1, -2, -3, -4, -5, -6},
		[]float64{-1500, -3500},
		[]float64{-1, 1},
		[]float64{-0, 0},
		[]float64{0},
	}
	outputs := []float64{
		1, 17, 2, 3.5, 2500, -1, -17, -2, -3.5, -2500, 0, 0, 0,
	}
	for i := range inputs {
		input := inputs[i]
		expected := outputs[i]
		actual := meanFloat(input)
		if expected != actual {
			pureFunctionError(t, input, expected, actual)
		}
	}
}

func TestExtremaIndex(t *testing.T) {
	t.Parallel()
	funs := []compare{
		minFunc, maxFunc,
		minFunc, maxFunc,
		minFunc, maxFunc,
	}
	inputs := [][]float64{
		[]float64{1, 2, 3, 4, 5, 6, 7, 8, 9}, []float64{1, 2, 3, 4, 5, 6, 7, 8, 9},
		[]float64{-1, 0, 1}, []float64{-1, 0, 1},
		[]float64{2, -18, 9999, -123981, 134}, []float64{2, -18, 9999, -123981, 134},
	}
	outputs := []int{
		0, 8,
		0, 2,
		3, 2,
	}
	for i := range funs {
		fun := funs[i]
		input := inputs[i]
		expected := outputs[i]
		actual := extremaIndex(fun, input)
		if expected != actual {
			pureFunctionError(t, input, expected, actual)
		}
	}
}

func TestChauvenet(t *testing.T) {
	t.Parallel()
	intSliceEqual := func(slc1 []int, slc2 []int) bool {
		if len(slc1) != len(slc2) {
			return false
		}
		for i := range slc1 {
			if slc1[i] != slc2[i] {
				return false
			}
		}
		return true
	}
	inputs := [][]int{
		[]int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		[]int{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		[]int{20, 21, 20, 20, 19, 20, 20, 1000},
		[]int{-234, -230, 5, -226, -228, -230},
		//[]int{0, 1, 2, 3, 4, 5}, TODO
	}
	outputs := [][]int{
		[]int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		[]int{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		[]int{20, 21, 20, 20, 19, 20, 20},
		[]int{-234, -230, -226, -228, -230},
		//[]int{0, 1, 2, 3, 4, 5},
	}
	for i := range inputs {
		input := inputs[i]
		expected := outputs[i]
		actual := chauvenet(input)
		if !intSliceEqual(expected, actual) {
			pureFunctionError(t, input, expected, actual)
		}
	}
}

// TODO two digit numbers
// tables to be split
var inputTables = []string{
	// single space
	`1 2 3 4
1 2 3 4
1 2 3 4`,
	// double space
	`
1  2  3  4
1  2  3  4
1  2  3  4  5
1  2  3  4
1  2  3  4`,
	// tabs
	"1\t2\n1\t2\n1\t2\n1\t2\n",
	// four space inconsistent (choose 2)
	`1    2    3    4    5    6    7    8    9
1    2    3    4    5    6    7    8    9
1    2    3    4    5    6    7    8    9
1    2    3    4    5    6    7    8    9
1    2    3    4    5    6    7    8    9
1   2    3   4    5   6    7   8
1    2    3    4    5    6    7    8    9
1    2    3    4    5    6    7    8    9
1    2    3    4    5    6    7    8    9 `,
}

var outputTables = []Table{
	[][]string{
		[]string{"1", "2", "3", "4"},
		[]string{"1", "2", "3", "4"},
		[]string{"1", "2", "3", "4"},
	},
	[][]string{
		[]string{"1", "2", "3", "4"},
		[]string{"1", "2", "3", "4"},
		[]string{"1", "2", "3", "4", "5"},
		[]string{"1", "2", "3", "4"},
		[]string{"1", "2", "3", "4"},
	},
	[][]string{
		[]string{"1", "2"},
		[]string{"1", "2"},
		[]string{"1", "2"},
		[]string{"1", "2"},
	},
	[][]string{
		[]string{"1", "2", "3", "4", "5", "6", "7", "8", "9"},
		[]string{"1", "2", "3", "4", "5", "6", "7", "8", "9"},
		[]string{"1", "2", "3", "4", "5", "6", "7", "8", "9"},
		[]string{"1", "2", "3", "4", "5", "6", "7", "8", "9"},
		[]string{"1", "2", "3", "4", "5", "6", "7", "8", "9"},
		[]string{"1", "2", "3", "4", "5", "6", "7", "8"},
		[]string{"1", "2", "3", "4", "5", "6", "7", "8", "9"},
		[]string{"1", "2", "3", "4", "5", "6", "7", "8", "9"},
		[]string{"1", "2", "3", "4", "5", "6", "7", "8", "9"},
	},
}

func TestGetColumnRegex(t *testing.T) {
	t.Parallel()
	// regexpEqual tests whether or not two regexps are the same by examining
	// the strings that constructed them
	regexpEqual := func(re1 *regexp.Regexp, re2 *regexp.Regexp) bool {
		return re1.String() == re2.String()
	}
	outputs := []*regexp.Regexp{
		regexp.MustCompile(`\s+`),
		regexp.MustCompile(`\s{2,}`),
		regexp.MustCompile(`\t+`),
		regexp.MustCompile(`\s{2,}`),
	}
	rowSep := regexp.MustCompile(`\n+`)
	for i := range inputTables {
		input := inputTables[i]
		expected := outputs[i]
		actual := getColumnRegex(input, rowSep)
		if !regexpEqual(expected, actual) {
			pureFunctionError(t, input, expected, actual)
		}
	}
}

func TestProbabalisticSplit(t *testing.T) {
	t.Parallel()
	for i := range inputTables {
		input := inputTables[i]
		expected := outputTables[i]
		actual := ProbabalisticSplit(input)
		if !TableEqual(expected, actual) {
			msg := "Actual output did not match expected"
			msg += "\n\tInput: " + input
			msg += "\n\tExpected:\n" + ToString(expected)
			msg += "\n\tActual:\n" + ToString(actual)
			t.Error(msg)
		}
	}
}
