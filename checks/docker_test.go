package checks

import (
	"testing"
)

func TestDockerImage(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("Skipping docker tests in short mode")
	} else {
		validInputs := names
		invalidInputs := notLengthOne
		// inputs that should lead to success
		goodEggs := [][]string{}
		// inputs that should lead to failure
		badEggs := [][]string{{"lkjbdakjsd"}, {"failme"}}
		badEggs = append(badEggs, names...)
		testParameters(validInputs, invalidInputs, DockerImage{}, t)
		testCheck(goodEggs, badEggs, DockerImage{}, t)
	}
}

func TestDockerImageRegexp(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("Skipping docker tests in short mode")
	} else {
		validInputs := [][]string{
			{"name"}, {"test*"}, {`win\d{1}`},
		}
		validInputs = append(validInputs, names...)
		// TODO invalid regexps
		invalidInputs := notLengthOne
		goodEggs := [][]string{}
		badEggs := [][]string{{"lkjbdakjsd{3}"}, {"failme+"}}
		badEggs = append(badEggs, names...)
		testParameters(validInputs, invalidInputs, DockerImageRegexp{}, t)
		testCheck(goodEggs, badEggs, DockerImageRegexp{}, t)
	}
}

func TestDockerRunning(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("Skipping docker tests in short mode")
	} else {
		validInputs := names
		invalidInputs := notLengthOne
		goodEggs := [][]string{}
		badEggs := [][]string{{"lkjbdakjsd{3}"}, {"failme+"}}
		badEggs = append(badEggs, names...)
		testParameters(validInputs, invalidInputs, DockerRunning{}, t)
		testCheck(goodEggs, badEggs, DockerRunning{}, t)
	}
}

func TestDockerRunningRegexp(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("Skipping docker tests in short mode")
	} else {
		validInputs := names
		// TODO invalid regexps
		invalidInputs := notLengthOne
		goodEggs := [][]string{}
		badEggs := [][]string{{"lkjbdakjsd{3}"}, {"failme+"}}
		badEggs = append(badEggs, names...)
		testParameters(validInputs, invalidInputs, DockerRunning{}, t)
		testCheck(goodEggs, badEggs, DockerRunning{}, t)
	}
}

func TestDockerDaemonTimeout(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("Skipping docker tests in short mode")
	} else {
		validInputs := [][]string{{"5s"}, {"0m"}, {"1h"}}
		invalidInputs := append(notLengthOne, []string{"1fail", "sec"})
		badEggs := [][]string{{"0s"}, {".1μs"}}
		testParameters(validInputs, invalidInputs, DockerDaemonTimeout{}, t)
		testCheck([][]string{}, badEggs, DockerDaemonTimeout{}, t)
	}
}
