package workers

import (
	"fmt"
	"github.com/CiscoCloud/distributive/chkutil"
	"github.com/CiscoCloud/distributive/errutil"
	"github.com/CiscoCloud/distributive/tabular"
	log "github.com/Sirupsen/logrus"
	"os/user"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// validGroupName asks: Is this a valid POSIX+Linux group name?
func validGroupName(name string) bool {
	return !strings.Contains(name, ":")
}

// validUserName asks: Is this a valid POSIX+Linux username?
func validUsername(name string) bool {
	if len(name) > 32 {
		return false
	} else if strings.Contains(name, ":") {
		return false
	}
	return true
}

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
	data := chkutil.FileToString(path)
	rowSep := regexp.MustCompile(`\n`)
	colSep := regexp.MustCompile(`:`)
	lines := tabular.SeparateString(rowSep, colSep, data)
	commaRegexp := regexp.MustCompile(`,`)
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
// UserInGroup, and GroupID
func groupNotFound(name string) (int, string, error) {
	// get a nicely formatted list of groups that do exist
	var existing []string
	for _, group := range getGroups() {
		existing = append(existing, group.Name)
	}
	return errutil.GenericError("Group not found", name, existing)
}

/*
#### GroupExists
Description: Does this group exist?
Parameters:
  - Name (group name): Name of the group
Example parameters:
  - sudo, wheel, www, storage
*/

type GroupExists struct{ name string }

func (chk GroupExists) ID() string { return "GroupExists" }

func (chk GroupExists) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	} else if !validGroupName(params[0]) {
		return chk, errutil.ParameterTypeError{params[0], "group name"}
	}
	chk.name = params[0]
	return chk, nil
}

func (chk GroupExists) Status() (int, string, error) {
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
	if doesGroupExist(chk.name) {
		return errutil.Success()
	}
	return groupNotFound(chk.name)
}

/*
#### UserInGroup
Description: Is this user in this group?
Parameters:
  - User (user name): Name of the group
  - Group (group name): Name of the group
Example parameters:
  - lb, siddharthist, root, centos
  - sudo, wheel, www, storage
*/

type UserInGroup struct{ user, group string }

func (chk UserInGroup) ID() string { return "UserInGroup" }

func (chk UserInGroup) New(params []string) (chkutil.Check, error) {
	if len(params) != 2 {
		return chk, errutil.ParameterLengthError{2, params}
	} else if !validUsername(params[0]) {
		return chk, errutil.ParameterTypeError{params[0], "username"}
	} else if !validGroupName(params[1]) {
		return chk, errutil.ParameterTypeError{params[1], "group name"}
	}
	chk.user = params[0]
	chk.group = params[1]
	return chk, nil
}

func (chk UserInGroup) Status() (int, string, error) {
	groups := getGroups()
	for _, g := range groups {
		if g.Name == chk.group {
			if tabular.StrIn(chk.user, g.Users) {
				return errutil.Success()
			}
			msg := "User not found in group"
			return errutil.GenericError(msg, chk.user, g.Users)
		}
	}
	return groupNotFound(chk.group)
}

/*
#### GroupID
Description: Does this group have this integer ID?
Parameters:
  - Group (group name): Name of the group
  - ID (int): Group ID
Example parameters:
  - sudo, wheel, www, storage
  - 0, 20, 50, 38
*/

type GroupID struct {
	name string
	id   int
}

func (chk GroupID) ID() string { return "GroupID" }

func (chk GroupID) New(params []string) (chkutil.Check, error) {
	if len(params) != 2 {
		return chk, errutil.ParameterLengthError{2, params}
	} else if !validGroupName(params[0]) {
		return chk, errutil.ParameterTypeError{params[0], "group name"}
	}
	chk.name = params[0]
	id64, err := strconv.ParseInt(params[1], 10, 64)
	if err != nil {
		return chk, errutil.ParameterTypeError{params[1], "int"}
	}
	chk.id = int(id64)
	return chk, nil
}

func (chk GroupID) Status() (int, string, error) {
	groups := getGroups()
	for _, g := range groups {
		if g.Name == chk.name {
			if g.ID == chk.id {
				return errutil.Success()
			}
			msg := "Group does not have expected ID"
			return errutil.GenericError(msg, chk.id, []int{g.ID})
		}
	}
	return groupNotFound(chk.name)
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
func userHasField(usernameOrUID string, fieldName string, expected string) (bool, error) {
	// get user to look at their info
	user, err := lookupUser(usernameOrUID)
	if err != nil || user == nil {
		return false, err
	}
	// reflect and get values
	val := reflect.ValueOf(*user)
	fieldVal := val.FieldByName(fieldName)
	// check to see if the field is a string
	errutil.ReflectError(fieldVal, reflect.Struct, "userHasField")
	actualValue := fieldVal.String()
	return actualValue == expected, nil
}

// genericUserField constructs (int, string, error)s that check if a given field of a User
// object found by lookupUser has a given value
func genericUserField(usernameOrUID string, fieldName string, fieldValue string) (int, string, error) {
	boolean, err := userHasField(usernameOrUID, fieldName, fieldValue)
	if err != nil {
		return 1, "User does not exist: " + usernameOrUID, nil
	} else if boolean {
		return errutil.Success()
	}
	msg := "User does not have expected " + fieldName + ": "
	msg += "\nUser: " + usernameOrUID
	msg += "\nGiven: " + fieldValue
	return 1, msg, nil
}

/*
#### UserExists
Description: Does this user exist?
Parameters:
  - Username/UID (username or UID)
Example parameters:
  - lb, root, user, 10
*/

type UserExists struct{ usernameOrUID string }

func (chk UserExists) ID() string { return "UserExists" }

func (chk UserExists) New(params []string) (chkutil.Check, error) {
	// TODO validate usernameOrUID
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	chk.usernameOrUID = params[0]
	return chk, nil
}

func (chk UserExists) Status() (int, string, error) {
	if _, err := lookupUser(chk.usernameOrUID); err == nil {
		return errutil.Success()
	}
	return 1, "User does not exist: " + chk.usernameOrUID, nil
}

/*
#### UserHasUID
Description: Does this user have this UID?
Parameters:
  - Username/UID (username or UID)
  - Expected UID (UID)
Example parameters:
  - lb, root, user, 10
  - 11, 13, 17
*/

type UserHasUID struct {
	usernameOrUID string
	desiredUID    string
}

func (chk UserHasUID) ID() string { return "UserHasUID" }

func (chk UserHasUID) New(params []string) (chkutil.Check, error) {
	if len(params) != 2 {
		return chk, errutil.ParameterLengthError{2, params}
	}
	// simply check that it is an integer, no need to store it as such
	// since it is converted back to a string in the comparison
	_, err := strconv.ParseInt(params[1], 10, 32)
	if err != nil {
		return chk, errutil.ParameterTypeError{params[1], "int32"}
	}
	chk.desiredUID = params[1]
	// TODO validate usernameOrUID
	chk.usernameOrUID = params[0]
	return chk, nil
}

func (chk UserHasUID) Status() (int, string, error) {
	return genericUserField(chk.usernameOrUID, "Uid", chk.desiredUID)
}

/*
#### UserHasGID
Description: Does this user have this GID?
Parameters:
  - Username/UID (username or UID)
  - Expected GID (GID)
Example parameters:
  - lb, root, user, 10
  - 11, 13, 17
*/

type UserHasGID struct {
	usernameOrUID string
	desiredGID    string
}

func (chk UserHasGID) ID() string { return "UserHasGID" }

func (chk UserHasGID) New(params []string) (chkutil.Check, error) {
	if len(params) != 2 {
		return chk, errutil.ParameterLengthError{2, params}
	}
	// store it as a string for comparison
	_, err := strconv.ParseInt(params[1], 10, 32)
	if err != nil {
		return chk, errutil.ParameterTypeError{params[1], "int32"}
	}
	chk.desiredGID = params[1]
	// TODO validate usernameOrUID
	chk.usernameOrUID = params[0]
	return chk, nil
}

func (chk UserHasGID) Status() (int, string, error) {
	return genericUserField(chk.usernameOrUID, "Gid", string(chk.desiredGID))
}

/*
#### UserHasUsername
Description: Does this user have this username?
Parameters:
  - Username/UID (username or UID)
  - Expected Username (username)
Example parameters:
  - lb, 0, 12
  - lb, root, user
*/

type UserHasUsername struct{ usernameOrUID, expectedUsername string }

func (chk UserHasUsername) ID() string { return "UserHasUsername" }

func (chk UserHasUsername) New(params []string) (chkutil.Check, error) {
	if len(params) != 2 {
		return chk, errutil.ParameterLengthError{2, params}
	}
	chk.usernameOrUID = params[0]
	chk.expectedUsername = params[1]
	return chk, nil
}

func (chk UserHasUsername) Status() (int, string, error) {
	return genericUserField(chk.usernameOrUID, "Username", chk.expectedUsername)
}

/*
#### UserHasName
Description: Does this user have this real name?
Parameters:
  - Username/UID (username or UID)
  - Expected real name (string)
Example parameters:
  - lb, root, 0
  - langston, steve, brian
*/

type UserHasName struct{ usernameOrUID, expectedName string }

func (chk UserHasName) ID() string { return "UserHasName" }

func (chk UserHasName) New(params []string) (chkutil.Check, error) {
	if len(params) != 2 {
		return chk, errutil.ParameterLengthError{2, params}
	}
	// TODO validate username
	chk.usernameOrUID = params[0]
	chk.expectedName = params[1]
	return chk, nil
}

func (chk UserHasName) Status() (int, string, error) {
	return genericUserField(chk.usernameOrUID, "Name", chk.expectedName)
}

/*
#### UserHasHomeDir
Description: Does this user have this home directory?
Parameters:
  - Username/UID (username or UID)
  - Expected home directory (path)
Example parameters:
  - lb, root, 0
  - /home/lb, /root, /mnt/my/custom/dir
*/

type UserHasHomeDir struct{ usernameOrUID, expectedHomeDir string }

func (chk UserHasHomeDir) ID() string { return "UserHasHomeDir" }

func (chk UserHasHomeDir) New(params []string) (chkutil.Check, error) {
	if len(params) != 2 {
		return chk, errutil.ParameterLengthError{2, params}
	}
	// TODO validate username
	chk.usernameOrUID = params[0]
	chk.expectedHomeDir = params[1]
	return chk, nil
}

func (chk UserHasHomeDir) Status() (int, string, error) {
	return genericUserField(chk.usernameOrUID, "HomeDir", chk.expectedHomeDir)
}
