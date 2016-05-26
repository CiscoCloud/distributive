package checks

// Many of the following checks were copied from/inspired by goss, which has a
// compatible Apache license:
// https://github.com/aelsabbahy/goss/blob/d28f3cc6d708fb012ea614acf712eb56712a7de3/system/group.go
// https://github.com/aelsabbahy/goss/blob/d28f3cc6d708fb012ea614acf712eb56712a7de3/LICENSE

import (
	"strconv"
	"strings"

	"github.com/CiscoCloud/distributive/chkutil"
	"github.com/CiscoCloud/distributive/errutil"
	"github.com/CiscoCloud/distributive/tabular"
	gossuser "github.com/aelsabbahy/goss/system"
	gossutil "github.com/aelsabbahy/goss/util"
	libcontaineruser "github.com/opencontainers/runc/libcontainer/user"
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

/*
#### GroupExists
Description: Does this group exist?
Parameters:
- Name (group name): Name of the group
Example parameters:
- sudo, wheel, www, storage
*/

type GroupExists struct{ name string }

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
	_, err := libcontaineruser.LookupGroup(chk.name)
	if err != nil {
		return 1, "", err
	}
	return errutil.Success()
}

/*
#### UserInGroup
Description: Is this user in this group?
Parameters:
- User (user name): Name of the group
- Group (group name): Name of the group
Example parameters:
- siddharthist, siddharthist, root, centos
- sudo, wheel, www, storage
*/

type UserInGroup struct{ user, group string }

func init() {
    chkutil.Register("UserInGroup", func() chkutil.Check {
        return &UserInGroup{}
    })
    chkutil.Register("GroupID", func() chkutil.Check {
        return &GroupID{}
    })
    chkutil.Register("UserExists", func() chkutil.Check {
        return &UserExists{}
    })
    chkutil.Register("GroupExists", func() chkutil.Check {
        return &GroupExists{}
    })
    chkutil.Register("UserHasUID", func() chkutil.Check {
        return &UserHasUID{}
    })
    chkutil.Register("UserHasHomeDir", func() chkutil.Check {
        return &UserHasHomeDir{}
    })
    chkutil.Register("UserHasGID", func() chkutil.Check {
        return &UserHasGID{}
    })
}

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
	usr := gossuser.NewDefUser(chk.user, nil, gossutil.Config{})
	groups, err := usr.Groups()
	if err != nil {
		return 1, "", err
	} else if tabular.StrIn(chk.group, groups) {
		return errutil.Success()
	}
	return 1, "User not found in group", nil
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

func (chk GroupID) New(params []string) (chkutil.Check, error) {
	if len(params) != 2 {
		return chk, errutil.ParameterLengthError{2, params}
	} else if !validGroupName(params[0]) {
		return chk, errutil.ParameterTypeError{params[0], "group name"}
	}
	chk.name = params[0]
	id64, err := strconv.ParseInt(params[1], 10, 64)
	if err != nil {
		return chk, errutil.ParameterTypeError{params[1], "int64"}
	}
	chk.id = int(id64)
	return chk, nil
}

func (chk GroupID) Status() (int, string, error) {
	group, err := libcontaineruser.LookupGroup(chk.name)
	if err != nil {
		return 1, "", err
	} else if group.Gid == chk.id {
		return errutil.Success()
	}
	msg := "Group does not have expected ID"
	return errutil.GenericError(msg, chk.id, []int{group.Gid})
}

/*
#### UserExists
Description: Does this user exist?
Parameters:
- Username
Example parameters:
- siddharthist, root, user, 10
*/

type UserExists struct{ username string }

func (chk UserExists) New(params []string) (chkutil.Check, error) {
	// TODO validate username
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	chk.username = params[0]
	return chk, nil
}

func (chk UserExists) Status() (int, string, error) {
	if _, err := libcontaineruser.LookupUser(chk.username); err == nil {
		return errutil.Success()
	}
	return 1, "User does not exist: " + chk.username, nil
}

/*
#### UserHasUID
Description: Does this user have this UID?
Parameters:
- Username
- Expected UID (UID)
Example parameters:
- siddharthist, root, user, 10
- 11, 13, 17
*/

type UserHasUID struct {
	username    string
	expectedUID int
}

func (chk UserHasUID) ID() string { return "UserHasUID" }

func (chk UserHasUID) New(params []string) (chkutil.Check, error) {
	if len(params) != 2 {
		return chk, errutil.ParameterLengthError{2, params}
	}
	// TODO validate username
	chk.username = params[0]
	uidInt, err := strconv.ParseInt(params[1], 10, 32)
	if err != nil {
		return chk, errutil.ParameterTypeError{params[1], "int32"}
	}
	chk.expectedUID = int(uidInt)
	return chk, nil
}

func (chk UserHasUID) Status() (int, string, error) {
	usr, err := libcontaineruser.LookupUser(chk.username)
	if err != nil {
		return 1, "", err
	} else if usr.Uid == chk.expectedUID {
		return errutil.Success()
	}
	msg := "User " + chk.username + "didn't have UID" + string(chk.expectedUID)
	return 1, msg, nil
}

/*
#### UserHasGID
Description: Does this user have this GID?
Parameters:
- Username
- Expected GID (GID)
Example parameters:
- siddharthist, root, user, 10
- 11, 13, 17
*/

type UserHasGID struct {
	username    string
	expectedGID int
}

func (chk UserHasGID) New(params []string) (chkutil.Check, error) {
	if len(params) != 2 {
		return chk, errutil.ParameterLengthError{2, params}
	}
	gidInt, err := strconv.ParseInt(params[1], 10, 32)
	if err != nil {
		return chk, errutil.ParameterTypeError{params[1], "int32"}
	}
	chk.expectedGID = int(gidInt)
	// TODO validate username
	chk.username = params[0]
	return chk, nil
}

func (chk UserHasGID) Status() (int, string, error) {
	usr, err := libcontaineruser.LookupUser(chk.username)
	if err != nil {
		return 1, "", err
	} else if usr.Gid == chk.expectedGID {
		return errutil.Success()
	}
	msg := "User " + chk.username + "didn't have GID" + string(chk.expectedGID)
	return 1, msg, nil
}

/*
#### UserHasHomeDir
Description: Does this user have this home directory?
Parameters:
- Username
- Expected home directory (path)
Example parameters:
- siddharthist, root, 0
- /home/siddharthist, /root, /mnt/my/custom/dir
*/

type UserHasHomeDir struct{ username, expectedHomeDir string }

func (chk UserHasHomeDir) New(params []string) (chkutil.Check, error) {
	if len(params) != 2 {
		return chk, errutil.ParameterLengthError{2, params}
	}
	// TODO validate username
	chk.username = params[0]
	chk.expectedHomeDir = params[1]
	return chk, nil
}

func (chk UserHasHomeDir) Status() (int, string, error) {
	usr, err := libcontaineruser.LookupUser(chk.username)
	if err != nil {
		return 1, "", err
	} else if usr.Home == chk.expectedHomeDir {
		return errutil.Success()
	}
	msg := "User " + chk.username + "didn't have home dir " + chk.expectedHomeDir
	return 1, msg, nil
}
