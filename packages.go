package main

import (
	"fmt"
	"github.com/CiscoCloud/distributive/tabular"
	"log"
	"os/exec"
	"regexp"
	"strings"
)

// register these functions as workers
func registerPackage() {
	registerCheck("installed", installed, 1)
	registerCheck("repoexists", repoExists, 2)
	registerCheck("repoexistsuri", repoExistsURI, 2)
	registerCheck("pacmanignore", pacmanIgnore, 1)
}

// getKeys returns the string keys from a string -> string map
func getKeys(m map[string]string) []string {
	keys := make([]string, len(m))
	i := 0
	for key := range managers {
		keys[i] = key
		i++
	}
	return keys
}

// package managers and their options for queries
var managers map[string]string = map[string]string{
	"dpkg":   "-s",
	"rpm":    "-q",
	"pacman": "-Qs",
}
var keys []string = getKeys(managers)

// getManager returns package manager as a string
func getManager() string {
	for _, program := range keys {
		cmd := exec.Command(program, "--version")
		err := cmd.Start()
		// as long as the command was found, return that manager
		message := ""
		if err != nil {
			message = err.Error()
		}
		if strings.Contains(message, "not found") == false {
			return program
		}
	}
	log.Fatal("No package manager found. Attempted: " + fmt.Sprint(managers))
	return "" // never reaches this return
}

type Repo struct {
	Id, Name, Url, Status string
}

// repoToString converts a Repo struct into a representable, printable string
func repoToString(r Repo) (str string) {
	str += "Name: " + r.Name
	str += " Id: " + r.Id
	str += " URL: " + r.Url
	str += " Status: " + r.Status
	return str
}

// getYumRepos constructs Repos from the yum.conf file at path. Gives non-zero
// Names, Fullnames, and Urls.
func getYumRepos() (repos []Repo) {
	// get output of `yum repolist`
	cmd := exec.Command("yum", "repolist")
	out, err := cmd.Output()
	outstr := string(out)
	if err != nil {
		execError(cmd, outstr, err)
	}
	// parse output
	slc := tabular.StringToSliceMultispace(outstr)
	ids := tabular.GetColumnNoHeader(0, slc)
	ids = ids[:len(ids)-2] // has extra line at end
	names := tabular.GetColumnNoHeader(1, slc)
	statuses := tabular.GetColumnNoHeader(2, slc)
	if len(ids) != len(names) || len(names) != len(statuses) {
		fmt.Println(ids)
		fmt.Println(names)
		fmt.Println(statuses)
		fmt.Println(len(ids))
		fmt.Println(len(names))
		fmt.Println(len(statuses))
		log.Fatal("Could not fetch metadata for every repo")
	}
	// Construct Repos
	for i, _ := range ids {
		repo := Repo{Name: names[i], Id: ids[i], Status: statuses[i]}
		repos = append(repos, repo)
	}
	return repos
}

// getAptRepos constructs Repos from the sources.list file at path. Gives
// non-zero Urls
func getAptRepos() (repos []Repo) {
	// getAptSources returns all the urls of all apt sources (including source
	// code repositories
	getAptSources := func() (urls []string) {
		otherLists := getFilesWithExtension("/etc/apt/sources.list.d", ".list")
		sourceLists := append([]string{"/etc/apt/sources.list"}, otherLists...)
		for _, f := range sourceLists {
			split := tabular.StringToSlice(fileToString(f))
			// filter out comments
			commentRegex := regexp.MustCompile("^\\s*#.*")
			for _, line := range split {
				if len(line) > 1 && !(commentRegex.MatchString(line[0])) {
					urls = append(urls, line[1])
				}
			}
		}
		return urls
	}
	for _, src := range getAptSources() {
		repos = append(repos, Repo{Url: src})
	}
	return repos
}

// getPacmanRepos constructs Repos from the pacman.conf file at path. Gives
// non-zero Names and Urls
func getPacmanRepos(path string) (repos []Repo) {
	data := fileToLines(path)
	// match words and dashes in brackets without comments
	nameRegex := regexp.MustCompile("[^#]\\[(\\w|\\-)+\\]")
	// match lines that start with Include= or Server= and anything after that
	urlRegex := regexp.MustCompile("[^#](Include|Server)\\s*=\\s*.*")
	var names []string
	var urls []string
	for _, line := range data {
		if nameRegex.Match(line) {
			names = append(names, string(nameRegex.Find(line)))
		}
		if urlRegex.Match(line) {
			urls = append(urls, string(urlRegex.Find(line)))
		}
	}
	for i, name := range names {
		if len(urls) > i {
			repos = append(repos, Repo{Name: name, Url: urls[i]})
		}
	}
	return repos
}

// getRepos simply returns a list of Repos based on the manager chosen
func getRepos(manager string) (repos []Repo) {
	switch manager {
	case "yum":
		return getYumRepos()
	case "apt":
		return getAptRepos()
	case "pacman":
		return getPacmanRepos("/etc/pacman.conf")
	default:
		msg := "Cannot find repos of unsupported package manager: "
		_, message := genericError(msg, manager, []string{getManager()})
		log.Fatal(message)
	}
	return []Repo{} // will never reach here b/c of default case
}

// existsRepoWithProperty is an abstraction of YumRepoExists and YumRepoURL.
// It takes a struct field name to check, and an expected value. If the expected
// value is found in the field of a repo, it returns 0, "" else an error message.
// Valid choices for prop: "Url" | "Name" | "Name"
func existsRepoWithProperty(prop string, val *regexp.Regexp, manager string) (int, string) {
	var properties []string
	for _, repo := range getRepos(manager) {
		switch prop {
		case "Url":
			properties = append(properties, repo.Url)
		case "Name":
			properties = append(properties, repo.Name)
		case "Status":
			properties = append(properties, repo.Status)
		case "Id":
			properties = append(properties, repo.Id)
		default:
			log.Fatal("Repos don't have the requested property: " + prop)
		}
	}
	if tabular.ReIn(val, properties) {
		return 0, ""
	}
	msg := "Repo with given " + prop + " not found"
	return genericError(msg, val.String(), properties)
}

// repoExists checks to see that a given repo is listed in the appropriate
// configuration file
func repoExists(parameters []string) (exitCode int, exitMessage string) {
	re := parseUserRegex(parameters[1])
	return existsRepoWithProperty("Name", re, parameters[0])
}

// repoExistsURI checks to see if the repo with the given URI is listed in the
// appropriate configuration file
func repoExistsURI(parameters []string) (exitCode int, exitMessage string) {
	re := parseUserRegex(parameters[1])
	return existsRepoWithProperty("Url", re, parameters[0])
}

// pacmanIgnore checks to see whether a given package is in /etc/pacman.conf's
// IgnorePkg setting
func pacmanIgnore(parameters []string) (exitCode int, exitMessage string) {
	pkg := parameters[0]
	data := fileToString("/etc/pacman.conf")
	re := regexp.MustCompile("[^#]IgnorePkg\\s+=\\s+.+")
	find := re.FindString(data)
	var packages []string
	if find != "" {
		spl := strings.Split(find, " ")
		if len(spl) > 2 {
			packages = spl[2:] // first two are "IgnorePkg" and "="
			if tabular.StrIn(pkg, packages) {
				return 0, ""
			}
		}
	}
	msg := "Couldn't find package in IgnorePkg"
	return genericError(msg, pkg, packages)
}

// installed detects whether the OS is using dpkg, rpm, or pacman, queries
// a package accoringly, and returns an error if it is not installed.
func installed(parameters []string) (exitCode int, exitMessage string) {
	pkg := parameters[0]
	name := getManager()
	options := managers[name]
	out, _ := exec.Command(name, options, pkg).Output()
	if strings.Contains(string(out), pkg) {
		return 0, ""
	}
	msg := "Package was not found:"
	msg += "\n\tPackage name: " + pkg
	msg += "\n\tPackage manager: " + name
	return 1, msg
}
