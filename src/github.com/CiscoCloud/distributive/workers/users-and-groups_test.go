package workers

import (
	"fmt"
	"testing"
)

func TestGetGroups(t *testing.T) {
	t.Parallel()
	groups := getGroups()
	if len(groups) < 1 {
		t.Error("Couldn't find any groups in /etc/group")
	}
}

func TestGroupNotFound(t *testing.T) {
	t.Parallel()
	code, message := groupNotFound("dummyGroup")
	if code <= 0 || message == "" {
		msg := "groupNotFound isn't properly reporting errors as such"
		msg += "\n\tCode: " + fmt.Sprint(code)
		msg += "\n\tMessage: " + message
		t.Error(msg)
	}
}

func TestGroupExists(t *testing.T) {
	t.Parallel()
	winners := []parameters{
		[]string{"root"},
		[]string{"bin"},
		[]string{"daemon"},
		[]string{"sys"},
		[]string{"adm"},
		[]string{"tty"},
	}
	testInputs(t, groupExists, winners, names)
}

func TestGroupID(t *testing.T) {
	t.Parallel()
	winners := []parameters{
		[]string{"root", "0"},
		[]string{"bin", "1"},
		[]string{"daemon", "2"},
		[]string{"sys", "3"},
		[]string{"adm", "4"},
		[]string{"tty", "5"},
	}
	losers := appendParameter(names, "17389")
	testInputs(t, groupID, winners, losers)
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
	testInputs(t, userExists, []parameters{[]string{"0"}}, names)
}

func TestUserHasUID(t *testing.T) {
	t.Parallel()
	winners := []parameters{[]string{"root", "0"}} // not always true
	losers := appendParameter(names, "0")
	testInputs(t, userHasUID, winners, losers)
}

func TestUserHasGID(t *testing.T) {
	t.Parallel()
	winners := []parameters{[]string{"0", "0"}}
	losers := appendParameter(names, "0")
	testInputs(t, userHasGID, winners, losers)
}

func TestUserHasUsername(t *testing.T) {
	t.Parallel()
	winners := []parameters{[]string{"0", "root"}} // not always true
	losers := appendParameter(names, "dummyUsername")
	testInputs(t, userHasUsername, winners, losers)
}

func TestUserHasHomeDir(t *testing.T) {
	t.Parallel()
	winners := []parameters{[]string{"0", "/root"}} // not always true
	losers := appendParameter(names, "dummyHomeDir")
	testInputs(t, userHasHomeDir, winners, losers)
}
