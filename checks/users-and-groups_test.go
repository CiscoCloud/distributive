package checks

import "testing"

func TestGroupExists(t *testing.T) {
	t.Parallel()
	validInputs := names
	invalidInputs := notLengthOne
	goodEggs := [][]string{
		{"root"}, {"bin"}, {"daemon"}, {"sys"}, {"adm"}, {"tty"},
	}
	badEggs := names
	testParameters(validInputs, invalidInputs, GroupExists{}, t)
	testCheck(goodEggs, badEggs, GroupExists{}, t)
}

func TestGroupID(t *testing.T) {
	t.Parallel()
	validInputs := appendParameter(names, "0")
	invalidInputs := append(notLengthTwo, appendParameter(names, "notint")...)
	goodEggs := [][]string{
		{"root", "0"},
		{"bin", "1"},
		{"daemon", "2"},
		{"sys", "3"},
		{"adm", "4"},
		{"tty", "5"},
	}
	badEggs := appendParameter(names, "17389")
	testParameters(validInputs, invalidInputs, GroupID{}, t)
	testCheck(goodEggs, badEggs, GroupID{}, t)
}

func TestUserExists(t *testing.T) {
	t.Parallel()
	testParameters(names, notLengthOne, UserExists{}, t)
	testCheck([][]string{{"root"}}, names, UserExists{}, t)
}

func TestUserHasUID(t *testing.T) {
	t.Parallel()
	validInputs := appendParameter(names, "0")
	invalidInputs := append(notLengthTwo, appendParameter(names, "notint")...)
	goodEggs := [][]string{[]string{"root", "0"}} // not always true
	badEggs := appendParameter(names, "0")
	testParameters(validInputs, invalidInputs, UserHasUID{}, t)
	testCheck(goodEggs, badEggs, UserHasUID{}, t)
}

func TestUserHasGID(t *testing.T) {
	t.Parallel()
	validInputs := appendParameter(names, "0")
	invalidInputs := append(notLengthTwo, appendParameter(names, "notint")...)
	goodEggs := [][]string{[]string{"0", "0"}}
	badEggs := appendParameter(names, "0")
	testParameters(validInputs, invalidInputs, UserHasGID{}, t)
	testCheck(goodEggs, badEggs, UserHasGID{}, t)
}

func TestUserHasHomeDir(t *testing.T) {
	t.Parallel()
	validInputs := appendParameter(names, "/home/")
	goodEggs := [][]string{[]string{"0", "/root"}} // not always true
	badEggs := appendParameter(names, "/proc")
	testParameters(validInputs, notLengthTwo, UserHasHomeDir{}, t)
	testCheck(goodEggs, badEggs, UserHasHomeDir{}, t)
}
