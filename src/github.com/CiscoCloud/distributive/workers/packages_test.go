package workers

import (
	"fmt"
	"github.com/CiscoCloud/distributive/tabular"
	"testing"
)

func TestGetManager(t *testing.T) {
	t.Parallel()
	man := getManager()
	supported := []string{"pacman", "dpkg", "yum"}
	if !tabular.StrIn(man, supported) {
		msg := "getManager returned an unsupported package manager"
		msg += "\n\tReturned: " + man
		msg += "\n\tSupported: " + fmt.Sprint(supported)
		t.Error(msg)
	}
}

func TestGetRepos(t *testing.T) {
	t.Parallel()
	man := getManager()
	repos := getRepos(man)
	if len(repos) < 1 {
		msg := "getRepos couldn't didn't return any repos"
		msg += "\n\tManager: " + man
		t.Error(msg)
	}
}

// all the belowe are empty, only failing tests included. This is because we
// can't make any assumptions about the system these tests are being run on.

func TestRepoExists(t *testing.T) {
	t.Parallel()
	losers := reverseAppendParameter(names, getManager())
	testInputs(t, repoExists, []parameters{}, losers)
}

func TestRepoExistsURI(t *testing.T) {
	t.Parallel()
	losers := reverseAppendParameter(names, getManager())
	testInputs(t, repoExistsURI, []parameters{}, losers)
}

func TestPacmanIgnore(t *testing.T) {
	t.Parallel()
	testInputs(t, pacmanIgnore, []parameters{}, names)
}

func TestInstalled(t *testing.T) {
	t.Parallel()
	testInputs(t, installed, []parameters{}, names)
}
