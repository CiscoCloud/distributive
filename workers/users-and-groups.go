package workers

import (
	"fmt"
	"github.com/CiscoCloud/distributive/tabular"
	"github.com/CiscoCloud/distributive/wrkutils"
	log "github.com/Sirupsen/logrus"
	"os/user"
	"reflect"
	"regexp"
	"strconv"
)

// Group is a struct that contains all relevant information that can be parsed
// from an entry in /etc/group
type Group struct {
	Name  string
	ID    int
	Users []string
}

// getGroups returns a list of Group structs, as parsed from /etc/group
func getGroups() (groups []Group) {
	path := "/etc/group"
	data := wrkutils.FileToString(path)
	rowSep := regexp.MustCompile("\n")
	colSep := regexp.MustCompile(":")
	lines := tabular.SeparateString(rowSep, colSep, data)
	commaRegexp := regexp.MustCompile(",")
	for _, line := range lines {
		if len(line) > 3 { // only lines that have all fields (non-empty)
			gid, err := strconv.ParseInt(line[2], 10, 64)
			if err != nil {
				log.WithFields(log.Fields{
					"group": line[0],
					"path":  path,
				}).Fatal("Could not parse ID for group")
			}
			userSlice := commaRegexp.Split(line[3], -1)
			group := Group{Name: line[0], ID: int(gid), Users: userSlice}
			groups = append(groups, group)
		}
	}
	return groups
}

// groupNotFound creates generic error messages and exit codes for groupExits,
// userInGroup, and groupID
func groupNotFound(name string) (int, string) {
	// get a nicely formatted list of groups that do exist
	var existing []string
	for _, group := range getGroups() {
		existing = append(existing, group.Name)
	}
	return wrkutils.GenericError("Group not found", name, existing)
}

// groupExists determines whether a certain UNIX user group exists
func groupExists(parameters []string) (exitCode int, exitMessage string) {
	// doesGroupExist preforms all the meat of GroupExists
	doesGroupExist := func(name string) bool {
		groups := getGroups()
		for _, group := range groups {
			if group.Name == name {
				return true
			}
		}
		return false
	}
	name := parameters[0]
	if doesGroupExist(name) {
		return 0, ""
	}
	return groupNotFound(name)
}

// userInGroup checks whether or not a given user is in a given group
func userInGroup(parameters []string) (exitCode int, exitMessage string) {
	user := parameters[0]
	group := parameters[0]
	groups := getGroups()
	for _, g := range groups {
		if g.Name == group {
			if tabular.StrIn(user, g.Users) {
				return 0, ""
			}
			return wrkutils.GenericError("User not found in group", user, g.Users)
		}
	}
	return groupNotFound(group)
}

// groupID checks to see if a group of a certain name has a given integer id
func groupID(parameters []string) (exitCode int, exitMessage string) {
	name := parameters[0]
	id := wrkutils.ParseMyInt(parameters[1])
	groups := getGroups()
	for _, g := range groups {
		if g.Name == name {
			if g.ID == id {
				return 0, ""
			}
			msg := "Group does not have expected ID"
			return wrkutils.GenericError(msg, fmt.Sprint(id), []string{fmt.Sprint(g.ID)})
		}
	}
	return groupNotFound(name)
}

// lookupUser: Does the user with either the given username or given user id
// exist? Given argument can either be a string that can be parsed as an int
// (UID) or just a username
func lookupUser(usernameOrUID string) (*user.User, error) {
	usr, err := user.LookupId(usernameOrUID)
	if err != nil {
		usr, err = user.Lookup(usernameOrUID)
	}
	if err != nil {
		return usr, fmt.Errorf("Couldn't find user: " + usernameOrUID)
	}
	return usr, nil
}

// userHasField checks to see if the user of a given username or UID's struct
// field "fieldName" matches the given value. An abstraction of hasUID, hasGID,
// hasName, hasHomeDir, and userExists
func userHasField(usernameOrUID string, fieldName string, givenValue string) (bool, error) {
	// get user to look at their info
	user, err := lookupUser(usernameOrUID)
	if err != nil || user == nil {
		return false, err
	}
	// reflect and get values
	val := reflect.ValueOf(*user)
	fieldVal := val.FieldByName(fieldName)
	// check to see if the field is a string
	wrkutils.ReflectError(fieldVal, reflect.String, "userHasField")
	actualValue := fieldVal.String()
	return actualValue == givenValue, nil
}

// genericUserField constructs (exitCode int, exitMessage string)s that check if a given field of a User
// object found by lookupUser has a given value
func genericUserField(usernameOrUID string, fieldName string, fieldValue string) (exitCode int, exitMessage string) {
	boolean, err := userHasField(usernameOrUID, fieldName, fieldValue)
	if err != nil {
		return 1, "User does not exist: " + usernameOrUID
	} else if boolean {
		return 0, ""
	}
	msg := "User does not have expected " + fieldName + ": "
	msg += "\nUser: " + usernameOrUID
	msg += "\nGiven: " + fieldValue
	return 1, msg
}

// userExists checks to see if a given user exists by looking up their username
// or UID.
func userExists(parameters []string) (exitCode int, exitMessage string) {
	usernameOrUID := parameters[0]
	if _, err := lookupUser(usernameOrUID); err == nil {
		return 0, ""
	}
	return 1, "User does not exist: " + usernameOrUID
}

// userHasUID checks if the user of the given username or UID has the given
// UID.
func userHasUID(parameters []string) (exitCode int, exitMessage string) {
	return genericUserField(parameters[0], "Uid", parameters[1])
}

// userHasUsername checks if the user of the given username or UID has the given
// GID.
func userHasGID(parameters []string) (exitCode int, exitMessage string) {
	return genericUserField(parameters[0], "Gid", parameters[1])
}

// userHasUsername checks if the user of the given username or UID has the given
// username.
func userHasUsername(parameters []string) (exitCode int, exitMessage string) {
	return genericUserField(parameters[0], "Username", parameters[1])
}

// userHasName checks if the user of the given username or UID has the given
// name.
func userHasName(parameters []string) (exitCode int, exitMessage string) {
	return genericUserField(parameters[0], "Name", parameters[1])
}

// userHasHomeDir checks if the user of the given username or UID has the given
// home directory.
func userHasHomeDir(parameters []string) (exitCode int, exitMessage string) {
	return genericUserField(parameters[0], "HomeDir", parameters[1])
}
