package main

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"
)

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
	Name, Fullname, Url string
}

// repoToString converts a Repo struct into a representable, printable string
func repoToString(r Repo) (str string) {
	str += "Name: " + r.Name
	str += " Full name: " + r.Fullname
	str += " URL: " + r.Url
	return str
}

// getYumRepos constructs Repos from the yum.conf file at path. Gives non-zero
// Names, Fullnames, and Urls.
func getYumRepos(path string) (repos []Repo) {
	var fullNames []string
	var urls []string
	commentRegex := regexp.MustCompile("^\\s*#.*")
	for _, line := range fileToLines(path) {
		// filter comments and convert to string
		strLine := string(line)
		if !(commentRegex.Match(line)) {
			// first, attempt to replace the prefix
			replaceName := strings.TrimPrefix(strLine, "name=")
			replaceURL := strings.TrimPrefix(strLine, "baseurl=")
			// if they are different, we know a prefix was replaced
			if replaceName != strLine {
				fullNames = append(fullNames, replaceName)
			} else if replaceURL != strLine {
				urls = append(urls, replaceURL)
			}
		}
	}
	// Get shortest list to zip with, so we don't get an index error
	shortList := fullNames
	if len(fullNames) > len(urls) {
		shortList = urls
	}
	// Construct Repos
	whitespaceRegex := regexp.MustCompile("\\s+")
	for i, _ := range shortList {
		nameSplit := whitespaceRegex.Split(fullNames[i], -1)
		shortName := nameSplit[len(nameSplit)-1]
		repo := Repo{Name: shortName, Fullname: fullNames[i], Url: urls[i]}
		repos = append(repos, repo)
	}
	return repos
}

// getAptRepos constructs Repos from the sources.list file at path. Gives
// non-zero Urls
func getAptRepos(path string) (repos []Repo) {
	// getAptSources returns all the urls of all apt sources (including source
	// code repositories
	getAptSources := func(path string) (urls []string) {
		split := stringToSlice(fileToString(path))
		// filter out comments
		commentRegex := regexp.MustCompile("^\\s*#.*")
		for _, line := range split {
			if len(line) > 1 && !(commentRegex.MatchString(line[0])) {
				urls = append(urls, line[1])
			}
		}
		return urls
	}
	for _, src := range getAptSources(path) {
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
		return getYumRepos("/etc/yum.conf")
	case "apt":
		return getAptRepos("/etc/apt/sources.list")
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
// Valid choices for prop: "Url" | "Name" | "Fullname"
func existsRepoWithProperty(prop string, val string, manager string) (int, string) {
	var properties []string
	for _, repo := range getRepos(manager) {
		switch prop {
		case "Url":
			properties = append(properties, repo.Url)
		case "Name":
			properties = append(properties, repo.Name)
		case "Fullname":
			properties = append(properties, repo.Fullname)
		default:
			log.Fatal("Repos don't have the requested property: " + prop)
		}
	}
	if strIn(val, properties) {
		return 0, ""
	}
	msg := "Repo with given " + prop + " not found"
	return genericError(msg, val, properties)
}

// repoExists checks to see that a given repo is listed in the appropriate
// configuration file
func repoExists(parameters []string) (exitCode int, exitMessage string) {
	return existsRepoWithProperty("Name", parameters[1], parameters[0])
}

// repoExistsURI checks to see if the repo with the given URI is listed in the
// appropriate configuration file
func repoExistsURI(parameters []string) (exitCode int, exitMessage string) {
	return existsRepoWithProperty("Url", parameters[1], parameters[0])
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
			if strIn(pkg, packages) {
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
