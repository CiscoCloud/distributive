package workers

import (
	"fmt"
	"github.com/CiscoCloud/distributive/chkutil"
	"testing"
)

// names is a dummy set of parameters for tests to fail on
var names = [][]string{
	// infinite jest
	[]string{"incandenza"},
	[]string{"van dyne"},
	[]string{"pemulis"},
	[]string{"Lenz"},
	[]string{"Stice"},
	[]string{"Schitt"},
	// glass bead game
	[]string{"Knecht"},
	[]string{"Designori"},
	[]string{"Tegularius"},
	[]string{"Jacobus"},
	// steppenwolf
	[]string{"harry haller"},
	[]string{"loering"},
	[]string{"hermine"},
}

var ints = [][]string{
	[]string{"0"},
	[]string{"-1"},
	[]string{"1000"},
	[]string{"17"},
	[]string{"-23"},
	[]string{"127"},
	[]string{"314"},
	[]string{"2147483647"},
	[]string{"-2147483648"},
}

var notInts = [][]string{
	[]string{"1/2"},
	[]string{"-999999999999999999999999999999999999"},
	[]string{"999999999999999999999999999999999999"},
	[]string{"0/0"},
	[]string{"0+0"},
	[]string{"12hi31"},
}

var notLengthOne = [][]string{
	[]string{}, []string{"", ""}, []string{"", "", ""}, []string{"one", "two"},
}

var notLengthTwo = append(names,
	[]string{}, []string{""}, []string{"", "", ""}, []string{"one"},
)

// TODO invalid regexps

// function for manipulating parameters
type paramAlt func(param []string, str string) []string

// map the given function over params, return the result. Function must take
// two inputs: a string and a slice of strings, and must return a new slice.
func manipulateParameters(alt paramAlt, params [][]string, str string) (toReturn [][]string) {
	for _, param := range params {
		toReturn = append(toReturn, alt(param, str))
	}
	return toReturn
}

// prefixes the first parameter with a string
func prefixParameter(params [][]string, str string) (toReturn [][]string) {
	alt := func(p []string, s string) []string { return []string{s + p[0]} }
	return manipulateParameters(alt, params, str)
}

func suffixParameter(params [][]string, str string) (toReturn [][]string) {
	alt := func(p []string, s string) []string { return []string{p[0] + s} }
	return manipulateParameters(alt, params, str)
}

// add another parameter on the back of the parameter slices
func appendParameter(params [][]string, str string) (toReturn [][]string) {
	alt := func(p []string, s string) []string { return append(p, s) }
	return manipulateParameters(alt, params, str)
}

// add another parameter on the front of the parameter slices
func reverseAppendParameter(params [][]string, str string) (toReturn [][]string) {
	alt := func(p []string, s string) []string { return append([]string{s}, p...) }
	return manipulateParameters(alt, params, str)
}

// The two functions below form the basis of all check testing. They validate
// parameters, and then see if some sets pass or fail. These parameter sets are
// hardcoded, and should always work as expected (except for extreme cases).

func testParameters(goodEggs [][]string, badEggs [][]string, chk chkutil.Check, t *testing.T) {
	for _, goodEgg := range goodEggs {
		newChk, err := chk.New(goodEgg)
		if err != nil {
			msg := "Supposedly valid parameters were invalid"
			msg = "\n\tParameters: " + fmt.Sprint(goodEgg)
			msg = "\n\tError: " + err.Error()
			t.Error(msg)
		} else if newChk == nil {
			msg := "chk.New returned nil!"
			msg = "\n\tParameters: " + fmt.Sprint(goodEgg)
			t.Error(msg)
		}
	}
	for _, badEgg := range badEggs {
		_, err := chk.New(badEgg)
		if err == nil {
			msg := "Supposedly invalid parameters were valid"
			msg = "\n\tParameters: " + fmt.Sprint(badEgg)
			t.Error(msg)
		}
	}
}

func testCheck(goodEggs [][]string, badEggs [][]string, chk chkutil.Check, t *testing.T) {
	getNewChk := func(chk chkutil.Check, params []string, t *testing.T) chkutil.Check {
		newChk, err := chk.New(params)
		if err != nil {
			msg := "Supposedly valid parameters were invalid in testCheck"
			msg = "\n\tParameters: " + fmt.Sprint(params)
			msg = "\n\tError: " + err.Error()
			t.Error(msg)
		} else if newChk == nil {
			msg := "chk.New returned nil!"
			msg = "\n\tParameters: " + fmt.Sprint(params)
			t.Error(msg)
		}
		return newChk
	}
	for _, goodEgg := range goodEggs {
		newChk := getNewChk(chk, goodEgg, t)
		code, exitMsg, err := newChk.Status()
		if err != nil {
			msg := "Unexpected error while running check"
			msg += "\n\tError: " + err.Error()
			t.Error(msg)
		} else if code != 0 {
			msg := "Parameter set returned unexpected exit code"
			msg += "\n\tParameters: " + fmt.Sprint(goodEgg)
			msg += "\n\tExpected: 0"
			msg += "\n\tActual: " + fmt.Sprint(code)
			t.Error(msg)
		} else if exitMsg != "" {
			msg := "Parameter set returned unexpected message"
			msg += "\n\tParameters: " + fmt.Sprint(goodEgg)
			msg += "\n\tExpected: ''"
			msg += "\n\tActual: " + exitMsg
			t.Error(msg)
		}
	}
	for _, badEgg := range badEggs {
		newChk := getNewChk(chk, badEgg, t)
		code, exitMsg, err := newChk.Status()
		if err != nil {
			msg := "Unexpected error while running check"
			msg += "\n\tError: " + err.Error()
			t.Error(msg)
		} else if code == 0 {
			msg := "Parameter set returned unexpected exit code"
			msg += "\n\tParameters: " + fmt.Sprint(badEgg)
			msg += "\n\tExpected: != 0"
			msg += "\n\tActual: " + fmt.Sprint(code)
			t.Error(msg)
		} else if exitMsg == "" {
			msg := "Parameter set returned unexpected message"
			msg += "\n\tParameters: " + fmt.Sprint(badEgg)
			msg += "\n\tExpected: != ''"
			msg += "\n\tActual: " + exitMsg
			t.Error(msg)
		}
	}
}
