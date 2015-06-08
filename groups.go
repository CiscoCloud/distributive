package main

import (
	"fmt"
	"io/ioutil"
	"log"
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
	data, err := ioutil.ReadFile("/etc/group")
	fatal(err)
	rowSep := regexp.MustCompile("\n")
	colSep := regexp.MustCompile(":")
	lines := separateString(rowSep, colSep, string(data))
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

// GroupExists determines whether a certain UNIX user group exists
func GroupExists(name string) Thunk {
	return func() (exitCode int, exitMessage string) {
		groups := getGroups()
		for _, group := range groups {
			if group.Name == name {
				return 0, ""
			}
		}
		msg := "Group not found: " + name
		msg += "\nExisting groups: " + fmt.Sprint(groups)
		return 1, msg
	}
}

// UserInGroup checks whether or not a given user is in a given group
func UserInGroup(user string, group string) Thunk {
	return func() (exitCode int, exitMessage string) {
		groups := getGroups()
		for _, g := range groups {
			if g.Name == group && strIn(user, g.Users) {
				return 0, ""
			}
		}
		return 1, "User not found in group: " + user + " " + group
	}
}

// GroupId checks to see if a group of a certain name has a given integer id
func GroupId(name string, id int) Thunk {
	return func() (exitCode int, exitMessage string) {
		groups := getGroups()
		for _, g := range groups {
			if g.Name == name && g.Id == id {
				return 0, ""
			}
		}
		return 1, "Group does not have ID: " + name + " " + fmt.Sprint(id)
	}
}
