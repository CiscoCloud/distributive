package checks

import (
	"testing"
)

func TestDockerImage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping docker tests in short mode")
	} else {
		t.Parallel()
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
	if testing.Short() {
		t.Skip("Skipping docker tests in short mode")
	} else {
		t.Parallel()
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
	if testing.Short() {
		t.Skip("Skipping docker tests in short mode")
	} else {
		t.Parallel()
		validInputs := names
		invalidInputs := notLengthOne
		goodEggs := [][]string{}
		badEggs := [][]string{{"lkjbdakjsd{3}"}, {"failme+"}}
		badEggs = append(badEggs, names...)
		testParameters(validInputs, invalidInputs, DockerRunning{}, t)
		testCheck(goodEggs, badEggs, DockerRunning{}, t)
	}
}

func TestDockerRunningAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping docker tests in short mode")
	} else {
		t.Parallel()
		validInputs := [][]string{
			{"/proc/cpuinfo", "name"},
			{"/proc/cpuinfo", "test"},
			{"/proc/cpuinfo", "win"},
		}
		invalidInputs := notLengthOne
		invalidInputs = append(invalidInputs, names...)
		goodEggs := [][]string{}
		badEggs := [][]string{
			{"/var/run/docker.sock", "failme"},
			{"/var/run/docker.sock", "fail"},
			{"/var/run/docker.sock", "loser"},
		}
		testParameters(validInputs, invalidInputs, DockerRunningAPI{}, t)
		testCheck(goodEggs, badEggs, DockerRunningAPI{}, t)
	}
}

func TestDockerRunningRegexp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping docker tests in short mode")
	} else {
		t.Parallel()
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
