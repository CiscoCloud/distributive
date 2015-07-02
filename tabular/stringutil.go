package tabular

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
