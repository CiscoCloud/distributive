package checks

import (
	"fmt"
	"github.com/CiscoCloud/distributive/tabular"
	"strings"
	"testing"
)

func TestGetManager(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	// simply make sure we're not panicing
	_ = getRepos(getManager())
}

// all the belowe are empty, only failing tests included. This is because we
// can't make any assumptions about the system these tests are being run on.

func TestRepoExists(t *testing.T) {
	t.Parallel()
	// dpkg will fail on invalid package name
	validPackageNames := [][]string{}
	for _, name := range names {
		newName := strings.Replace(name[0], " ", "-", -1)
		validPackageNames = append(validPackageNames, []string{newName})
	}
	validInputs := reverseAppendParameter(validPackageNames, getManager())
	invalidInputs := reverseAppendParameter(names, "nonsense")
	goodEggs := [][]string{}
	badEggs := reverseAppendParameter(validPackageNames, getManager())
	invalidInputs = append(invalidInputs, notLengthTwo...)
	testParameters(validInputs, invalidInputs, RepoExists{}, t)
	testCheck(goodEggs, badEggs, RepoExists{}, t)
}

func TestRepoExistsURI(t *testing.T) {
	t.Parallel()
	// dpkg will fail on invalid package name
	validPackageNames := [][]string{}
	for _, name := range names {
		newName := strings.Replace(name[0], " ", "-", -1)
		validPackageNames = append(validPackageNames, []string{newName})
	}
	validInputs := reverseAppendParameter(validPackageNames, getManager())
	invalidInputs := reverseAppendParameter(names, "nonsense")
	invalidInputs = append(invalidInputs, notLengthTwo...)
	goodEggs := [][]string{}
	badEggs := reverseAppendParameter(validPackageNames, getManager())
	testParameters(validInputs, invalidInputs, RepoExistsURI{}, t)
	testCheck(goodEggs, badEggs, RepoExistsURI{}, t)
}

func TestPacmanIgnore(t *testing.T) {
	t.Parallel()
	testParameters(names, notLengthOne, PacmanIgnore{}, t)
	if getManager() == "pacman" {
		testCheck([][]string{}, names, PacmanIgnore{}, t)
	}
}

func TestInstalled(t *testing.T) {
	t.Parallel()
	// dpkg will fail on invalid package name
	validPackageNames := [][]string{}
	for _, name := range names {
		newName := strings.Replace(name[0], " ", "-", -1)
		validPackageNames = append(validPackageNames, []string{newName})
	}
	testParameters(validPackageNames, notLengthOne, Installed{}, t)
	testCheck([][]string{}, names, Installed{}, t)
}
