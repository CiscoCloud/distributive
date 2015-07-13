package workers

import (
	"testing"
)

type codeTest func(exitCode int) (passing bool, msg string)
type msgTest func(exitMessage string) (passing bool, msg string)
type worker func(parameters []string) (exitCode int, exitMessage string)

// type codeTest
func gtZero(exitCode int) (passing bool, msg string) {
	if exitCode > 0 {
		return true, ""
	}
	return false, "Expected an exit code > 0"
}

// type codeTest
func eqZero(exitCode int) (passing bool, msg string) {
	if exitCode == 0 {
		return true, ""
	}
	return false, "Expected an exit code == 0"
}

// applies cTest to exitCode and mTest to exitMessage of wrk when called with
// params
func generalTestFunction(wrk worker, params []string, cTest codeTest, mTest msgTest, t *testing.T) {
	exitCode, exitMessage := wrk(params)
	if passing, msg := cTest(exitCode); !passing {
		t.Error("Test on exit code failed with output:\n\t" + msg)
	} else if passing, msg := mTest(exitMessage); !passing {
		t.Error("Test on exit message failed with output:\n\t" + msg)
	}
}
