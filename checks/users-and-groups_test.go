package checks

import (
	"fmt"
	"testing"
)

var validUsernamesOrUIDs = append(append(names, smallInts...), bigIntsUnder100...)

func TestGetGroups(t *testing.T) {
	t.Parallel()
	groups := getGroups()
	if len(groups) < 1 {
		t.Error("Couldn't find any groups in /etc/group")
	}
}

func TestGroupNotFound(t *testing.T) {
	t.Parallel()
	code, message, _ := groupNotFound("dummyGroup")
	if code <= 0 || message == "" {
		msg := "groupNotFound isn't properly reporting errors as such"
		msg += "\n\tCode: " + fmt.Sprint(code)
		msg += "\n\tMessage: " + message
		t.Error(msg)
	}
}

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

func TestLookupUser(t *testing.T) {
	t.Parallel()
	user, err := lookupUser("root")
	user2, err2 := lookupUser("0")
	msg := "Couldn't successfully lookup root user"
	if user == nil || err != nil {
		msg += "\n\tUsername: root"
		msg += "\n\tError: " + err.Error()
		t.Error(msg)
	} else if user2 == nil || err2 != nil {
		msg += "\n\tUID: 0"
		msg += "\n\tError: " + err2.Error()
		t.Error(msg)
	}
}

func TestUserExists(t *testing.T) {
	t.Parallel()
	testParameters(validUsernamesOrUIDs, notLengthOne, UserExists{}, t)
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

func TestUserHasUsername(t *testing.T) {
	t.Parallel()
	validInputs := reverseAppendParameter(names, "0")
	invalidInputs := notLengthTwo
	goodEggs := [][]string{[]string{"0", "root"}} // not always true
	badEggs := appendParameter(names, "nonsense")
	testParameters(validInputs, invalidInputs, UserHasUsername{}, t)
	testCheck(goodEggs, badEggs, UserHasUsername{}, t)
}

func TestUserHasHomeDir(t *testing.T) {
	t.Parallel()
	validInputs := appendParameter(names, "/home/")
	goodEggs := [][]string{[]string{"0", "/root"}} // not always true
	badEggs := appendParameter(names, "/proc")
	testParameters(validInputs, notLengthTwo, UserHasHomeDir{}, t)
	testCheck(goodEggs, badEggs, UserHasHomeDir{}, t)
}
