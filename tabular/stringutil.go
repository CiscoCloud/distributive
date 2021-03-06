package tabular

import (
	"strings"
)

// padString pads the string with terminal spaces if its length is less than
// the provided lengths
func padString(str string, pad string, length int) string {
	for len(str) < length {
		str = str + pad
	}
	return str
}

// repeatString repeats a given string a number of times
func repeatString(str string, length int) (result string) {
	for i := 0; i < length; i++ {
		result = result + str
	}
	return result
}

// IndiciesOf takes a list of needles and a haystack and returns the first
// locations of all the needles
func IndiciesOf(needles []string, haystack string) (indicies []int) {
	for _, needle := range needles {
		index := strings.Index(haystack, needle)
		indicies = append(indicies, index)
	}
	return indicies
}

// Lines splits a string on newlines
func Lines(str string) []string {
	return strings.Split(str, "\n")
}

// Unlines is the inverse operation of lines
func Unlines(slc []string) string {
	return strings.Join(slc, "\n")
}

func SliceEqual(slc1 []string, slc2 []string) bool {
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
