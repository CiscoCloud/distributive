package workers

import (
	"fmt"
	"github.com/CiscoCloud/distributive/tabular"
	"testing"
)

func TestGetManager(t *testing.T) {
	//t.Parallel()
	man := getManager()
	supported := []string{"pacman", "dpkg", "rpm"}
	if !tabular.StrIn(man, supported) {
		msg := "getManager returned an unsupported package manager"
		msg += "\n\tReturned: " + man
		msg += "\n\tSupported: " + fmt.Sprint(supported)
		t.Error(msg)
	}
}

func TestGetRepos(t *testing.T) {
	//t.Parallel()
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
	//t.Parallel()
	validInputs := reverseAppendParameter(names, getManager())
	invalidInputs := reverseAppendParameter(names, "nonsense")
	goodEggs := [][]string{}
	badEggs := reverseAppendParameter(names, getManager())
	invalidInputs = append(invalidInputs, notLengthTwo...)
	testParameters(validInputs, invalidInputs, RepoExists{}, t)
	testCheck(goodEggs, badEggs, RepoExists{}, t)
}

func TestRepoExistsURI(t *testing.T) {
	//t.Parallel()
	validInputs := reverseAppendParameter(names, getManager())
	invalidInputs := reverseAppendParameter(names, "nonsense")
	invalidInputs = append(invalidInputs, notLengthTwo...)
	goodEggs := [][]string{}
	badEggs := reverseAppendParameter(names, getManager())
	testParameters(validInputs, invalidInputs, RepoExistsURI{}, t)
	testCheck(goodEggs, badEggs, RepoExistsURI{}, t)
}

func TestPacmanIgnore(t *testing.T) {
	//t.Parallel()
	testParameters(names, notLengthOne, PacmanIgnore{}, t)
	if getManager() == "pacman" {
		testCheck([][]string{}, names, PacmanIgnore{}, t)
	}
}

func TestInstalled(t *testing.T) {
	//t.Parallel()
	testParameters(names, notLengthOne, Installed{}, t)
	testCheck([][]string{}, names, Installed{}, t)
}
