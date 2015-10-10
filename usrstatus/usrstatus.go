// usrstatus provides utility functions for querying several aspects of Unix
// users and groups, especially as pertains to monitoring.
package usrstatus

import (
	"github.com/CiscoCloud/distributive/tabular"
	"io/ioutil"
	"regexp"
	"strconv"
)

// Group is a struct that contains all relevant information that can be parsed
// from an entry in /etc/group, namely the group's name, integer ID, and which
// users are a part of it.
type Group struct {
	Name  string
	ID    int
	Users []string
}

// Groups returns a slice of Group structs, parsed from /etc/group.
func Groups() (groups []Group, err error) {
	path := "/etc/group"
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return groups, err
	}
	rowSep := regexp.MustCompile(`\n`)
	colSep := regexp.MustCompile(`:`)
	lines := tabular.SeparateString(rowSep, colSep, string(data))
	commaRegexp := regexp.MustCompile(`,`)
	for _, line := range lines {
		if len(line) > 3 { // only lines that have all fields (non-empty)
			gid, err := strconv.ParseInt(line[2], 10, 64)
			if err != nil {
				return groups, err
			}
			userSlice := commaRegexp.Split(line[3], -1)
			group := Group{Name: line[0], ID: int(gid), Users: userSlice}
			groups = append(groups, group)
		}
	}
	return groups, nil
}

// UserInGroup asks whether or not the given user is a part of the given group.
func UserInGroup(username, groupname string) (bool, error) {
	groups, err := Groups()
	if err != nil {
		return false, err
	}
	for _, group := range groups {
		if group.Name == groupname && tabular.StrIn(username, group.Users) {
			return true, nil
		}
	}
	return false, nil
}
