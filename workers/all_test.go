package workers

import (
	"fmt"
	"testing"
)

// names is a dummy set of parameters for tests to fail on
var names = []parameters{
	[]string{"incandenza"},
	[]string{"van dyne"},
	[]string{"pemulis"},
}

type worker func(parameters []string) (exitCode int, exitMessage string)
type parameters []string

// function for manipulating parameters
type paramAlt func(param parameters, str string) parameters

// map the given function over params, return the result. Function must take
// two inputs: a string and a slice of strings, and must return a new slice.
func manipulateParameters(alt paramAlt, params []parameters, str string) (toReturn []parameters) {
	for _, param := range params {
		toReturn = append(toReturn, alt(param, str))
	}
	return toReturn
}

// prefixes the first parameter with a string
func prefixParameter(params []parameters, str string) (toReturn []parameters) {
	alt := func(p parameters, s string) parameters { return []string{s + p[0]} }
	return manipulateParameters(alt, params, str)
}

func suffixParameter(params []parameters, str string) (toReturn []parameters) {
	alt := func(p parameters, s string) parameters { return []string{p[0] + s} }
	return manipulateParameters(alt, params, str)
}

// add another parameter on the back of the parameter slices
func appendParameter(params []parameters, str string) (toReturn []parameters) {
	alt := func(p parameters, s string) parameters { return append(p, s) }
	return manipulateParameters(alt, params, str)
}

// add another parameter on the front of the parameter slices
func reverseAppendParameter(params []parameters, str string) (toReturn []parameters) {
	alt := func(p parameters, s string) parameters { return append([]string{s}, p...) }
	return manipulateParameters(alt, params, str)
}

// Expects two sets of inputs; one working and one failing. Tests the check
// against all of those inputs and reports an error if the check succeeds or
// fails when it's not expected to.
func testInputs(t *testing.T, wrk worker, winners []parameters, losers []parameters) {
	for _, winner := range winners {
		exitCode, exitMessage := wrk(winner)
		if exitMessage != "" || exitCode != 0 {
			msg := "Check was expected to succeed, it failed."
			msg += "\n\tExit code: " + fmt.Sprint(exitCode)
			msg += " (expected 0)"
			msg += "\n\tMessage: " + exitMessage
			msg += " (expected '')"
			msg += "\n\tParameters: " + fmt.Sprint(winner)
			t.Error(msg)
		}
	}
	for _, loser := range losers {
		exitCode, exitMessage := wrk(loser)
		if exitMessage == "" || exitCode == 0 {
			msg := "Check was expected to fail, it succeeded."
			msg += "\n\tExit code: " + fmt.Sprint(exitCode)
			msg += " (expected >0)"
			msg += "\n\tMessage: " + exitMessage
			msg += " (expected non-empty)"
			msg += "\n\tParameters: " + fmt.Sprint(loser)
			t.Error(msg)
		}
	}
}
