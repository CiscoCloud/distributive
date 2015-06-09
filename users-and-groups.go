package main

import (
	"fmt"
	"log"
	"os/user"
	"reflect"
	"regexp"
	"strconv"
)

// Group is a struct that contains all relevant information that can be parsed
// from an entry in /etc/group
type Group struct {
	Name  string
	Id    int
	Users []string
}

// getGroups returns a list of Group structs, as parsed from /etc/group
func getGroups() (groups []Group) {
	data := fileToString("/etc/group")
	rowSep := regexp.MustCompile("\n")
	colSep := regexp.MustCompile(":")
	lines := separateString(rowSep, colSep, data)
	commaRegexp := regexp.MustCompile(",")
	for _, line := range lines {
		if len(line) > 3 { // only lines that have all fields (non-empty)
			gid, err := strconv.ParseInt(line[2], 10, 64)
			if err != nil {
				log.Fatal("Could not parse ID for group: " + line[0])
			}
			userSlice := commaRegexp.Split(line[3], -1)
			group := Group{Name: line[0], Id: int(gid), Users: userSlice}
			groups = append(groups, group)
		}
	}
	return groups
}

// groupNotFound creates generic error messages and exit codes for GroupExits,
// UserInGroup, and GroupId
func groupNotFound(name string) (int, string) {
	// get a nicely formatted list of groups that do exist
	var existing []string
	for _, group := range getGroups() {
		existing = append(existing, group.Name)
	}
	return notInError("Group not found:", name, existing)
}

// GroupExists determines whether a certain UNIX user group exists
func GroupExists(name string) Thunk {
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
	return func() (exitCode int, exitMessage string) {
		if doesGroupExist(name) {
			return 0, ""
		}
		return groupNotFound(name)
	}
}

// UserInGroup checks whether or not a given user is in a given group
func UserInGroup(user string, group string) Thunk {
	return func() (exitCode int, exitMessage string) {
		groups := getGroups()
		for _, g := range groups {
			if g.Name == group {
				if strIn(user, g.Users) {
					return 0, ""
				}
				return notInError("User not found in group", user, g.Users)
			}
		}
		return groupNotFound(group)
	}
}

// GroupId checks to see if a group of a certain name has a given integer id
func GroupId(name string, id int) Thunk {
	return func() (exitCode int, exitMessage string) {
		groups := getGroups()
		for _, g := range groups {
			if g.Name == name {
				if g.Id == id {
					return 0, ""
				}
				msg := "Group does not have expected ID:"
				return notInError(msg, fmt.Sprint(id), []string{fmt.Sprint(g.Id)})
			}
		}
		return groupNotFound(name)
	}
}

// lookupUser: Does the user with either the given username or given user id
// exist? Given argument can either be a string that can be parsed as an int
// (UID) or just a username
func lookupUser(usernameOrUid string) (*user.User, error) {
	usr, err := user.LookupId(usernameOrUid)
	if err != nil {
		usr, err = user.Lookup(usernameOrUid)
	}
	if err != nil {
		return usr, fmt.Errorf("Couldn't find user: " + usernameOrUid)
	}
	return usr, nil
}

// userHasField checks to see if the user of a given username or uid's struct
// field "fieldName" matches the given value. An abstraction of HasUID, HasGID,
// HasName, HasHomeDir, and UserExists
func userHasField(usernameOrUid string, fieldName string, givenValue string) (bool, error) {
	// get user to look at their info
	user, err := lookupUser(usernameOrUid)
	if err != nil || user == nil {
		return false, err
	}
	// reflect and get values
	val := reflect.ValueOf(*user)
	fieldVal := val.FieldByName(fieldName)
	// check to see if the field is a string
	if fieldVal.Kind() != reflect.String {
		msg := "Failure during reflection: Field is not a string:"
		msg += "\n\tField name: " + fieldName
		msg += "\n\tField Kind: " + fmt.Sprint(fieldVal.Kind())
		msg += "\n\tUser: " + user.Username
		log.Fatal(msg)
	}
	actualValue := fieldVal.String()
	return actualValue == givenValue, nil
}

// genericUserField constructs Thunks that check if a given field of a User
// object found by lookupUser has a given value
func genericUserField(usernameOrUid string, fieldName string, fieldValue string) Thunk {
	return func() (exitCode int, exitMessage string) {
		boolean, err := userHasField(usernameOrUid, fieldName, fieldValue)
		if err != nil {
			return 1, "User does not exist: " + usernameOrUid
		} else if boolean {
			return 0, ""
		}
		msg := "User does not have expected " + fieldName + ": "
		msg += "\nUser: " + usernameOrUid
		msg += "\nGiven: " + fieldValue
		return 1, msg
	}

}

// UserExists checks to see if a given user exists by looking up their username
// or UID.
func UserExists(usernameOrUid string) Thunk {
	return func() (exitCode int, exitMessage string) {
		if _, err := lookupUser(usernameOrUid); err == nil {
			return 0, ""
		}
		return 1, "User does not exist: " + usernameOrUid
	}
}

// UserHasUID checks if the user of the given username or uid has the given
// UID.
func UserHasUID(usernameOrUid string, uid string) Thunk {
	return genericUserField(usernameOrUid, "Uid", uid)
}

// UserHasUsername checks if the user of the given username or uid has the given
// GID.
func UserHasGID(usernameOrUid string, gid string) Thunk {
	return genericUserField(usernameOrUid, "Gid", gid)
}

// UserHasUsername checks if the user of the given username or uid has the given
// username.
func UserHasUsername(usernameOrUid string, username string) Thunk {
	return genericUserField(usernameOrUid, "Username", username)
}

// UserHasName checks if the user of the given username or uid has the given
// name.
func UserHasName(usernameOrUid string, name string) Thunk {
	return genericUserField(usernameOrUid, "Name", name)
}

// UserHasHomeDir checks if the user of the given username or uid has the given
// home directory.
func UserHasHomeDir(usernameOrUid string, homeDir string) Thunk {
	return genericUserField(usernameOrUid, "HomeDir", homeDir)
}
