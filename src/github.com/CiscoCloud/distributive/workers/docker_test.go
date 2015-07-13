package workers

import (
	"testing"
)

func TestDockerImage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping docker tests in short mode")
	} else {
		t.Parallel()
		losers := []parameters{[]string{"failme"}, []string{""}}
		testInputs(t, dockerImage, []parameters{}, losers)
	}
}

func TestDockerImageRegexp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping docker tests in short mode")
	} else {
		t.Parallel()
		losers := []parameters{[]string{"failme"}, []string{"regexp"}}
		testInputs(t, dockerImageRegexp, []parameters{}, losers)
	}
}

func TestDockerRunning(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping docker tests in short mode")
	} else {
		t.Parallel()
		losers := []parameters{[]string{"failme"}, []string{""}}
		testInputs(t, dockerRunning, []parameters{}, losers)
	}
}

func TestDockerRunningAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping docker tests in short mode")
	} else {
		t.Parallel()
		path := "/var/run/docker.sock"
		losers := []parameters{[]string{path, "failme"}, []string{path, ""}}
		testInputs(t, dockerRunningAPI, []parameters{}, losers)
	}
}

func TestDockerRunningRegexp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping docker tests in short mode")
	} else {
		t.Parallel()
		losers := []parameters{[]string{"failme"}, []string{""}}
		testInputs(t, dockerRunningRegexp, []parameters{}, losers)
	}
}
