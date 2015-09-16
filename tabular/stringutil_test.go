package tabular

import (
	"testing"
)

func TestPadString(t *testing.T) {
	t.Parallel()
	inputs := []string{"test1", "test2", "test3", "test4"}
	pads := []string{" ", "~", "-", " "}
	lengths := []int{10, 8, 9, 2}
	outputs := []string{"test1     ", "test2~~~", "test3----", "test4"}
	for i := range inputs {
		input := inputs[i]
		pad := pads[i]
		length := lengths[i]
		expected := outputs[i]
		actual := padString(input, pad, length)
		if expected != actual {
			pureFunctionError(t, input, expected, actual)
		}
	}
}

func TestRepeatString(t *testing.T) {
	t.Parallel()
	inputs := []string{"-", "~", "l", "lb", "1001"}
	lengths := []int{5, 2, 3, 4, 3}
	outputs := []string{"-----", "~~", "lll", "lblblblb", "100110011001"}
	for i := range inputs {
		input := inputs[i]
		length := lengths[i]
		expected := outputs[i]
		actual := repeatString(input, length)
		if expected != actual {
			pureFunctionError(t, input, expected, actual)
		}
	}
}

func TestIndiciesOf(t *testing.T) {
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
	needles := [][]string{
		[]string{"l", "b"},
		[]string{"a", "s", "t"},
		[]string{"r", "e", "p"},
	}
	haystacks := []string{"langston barrett", "asteris", "gopher"}
	outputs := [][]int{[]int{0, 9}, []int{0, 1, 2}, []int{5, 4, 2}}
	for i := range needles {
		needle := needles[i]
		haystack := haystacks[i]
		expected := outputs[i]
		actual := IndiciesOf(needle, haystack)
		if !intSliceEqual(expected, actual) {
			pureFunctionError(t, needle, expected, actual)
		}
	}
}

func TestSliceEqual(t *testing.T) {
	t.Parallel()
	inputs1 := [][]string{
		[]string{"a", "b", "&", "*", "0"},
		[]string{"a", "b", "&", "*", "0"},
		[]string{"a", "b", "&", "*", "0"},
		[]string{"-"},
	}
	inputs2 := [][]string{
		[]string{"a", "b", "&", "*", "0"},
		[]string{"a", "b", "*", "0"},
		[]string{"r", "b", "&", "*", "0"},
		[]string{"-"},
	}
	outputs := []bool{true, false, false, true, true}
	for i := range inputs1 {
		input1 := inputs1[i]
		input2 := inputs2[i]
		expected := outputs[i]
		actual := SliceEqual(input1, input2)
		if actual != expected {
			pureFunctionError(t, inputs2, expected, actual)
		}
	}
}
