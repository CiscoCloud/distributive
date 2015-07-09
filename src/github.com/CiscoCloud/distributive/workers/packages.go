package workers

import (
	"github.com/CiscoCloud/distributive/tabular"
	"github.com/CiscoCloud/distributive/wrkutils"
	log "github.com/Sirupsen/logrus"
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
var managers = map[string]string{
	"dpkg":   "-s",
	"rpm":    "-q",
	"pacman": "-Qs",
}
var keys = getKeys(managers)

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
	log.WithFields(log.Fields{
		"attempted": keys,
	}).Fatal("No supported package manager found.")
	return "" // never reaches this return
}

// repo is a unified interface for pacman, apt, and yum repos
type repo struct {
	ID     string
	Name   string // yum
	URL    string // apt, pacman
	Status string
}

// repoToString converts a repo struct into a representable, printable string
func repoToString(r repo) (str string) {
	str += "Name: " + r.Name
	str += " ID: " + r.ID
	str += " URL: " + r.URL
	str += " Status: " + r.Status
	return str
}

// getYumRepos constructs Repos from the yum.conf file at path. Gives non-zero
// Names, Fullnames, and URLs.
func getYumRepos() (repos []repo) {
	// safeAccess allows access w/o fear of a panic into a slice of strings
	safeAccess := func(slc []string, index int) string {
		// catch runtime panic
		defer func() {
			if err := recover(); err != nil {
				msg := "safeAccess: Please report this error"
				wrkutils.IndexError(msg, index, slc)
			}
		}() // invoke inside defer
		if len(slc) > index {
			return slc[index]
		}
		return ""
	}

	// get and parse output of `yum repolist`
	cmd := exec.Command("yum", "repolist")
	outstr := wrkutils.CommandOutput(cmd)
	slc := tabular.ProbabalisticSplit(outstr)

	ids := tabular.GetColumnNoHeader(0, slc) // TODO use columnbyheader here
	wrkutils.IndexError("getYumRepos", 2, ids)
	ids = ids[:len(ids)-2]

	names := tabular.GetColumnNoHeader(1, slc)    // TODO and here
	statuses := tabular.GetColumnNoHeader(2, slc) // TODO and here
	if len(ids) != len(names) || len(names) != len(statuses) {
		log.WithFields(log.Fields{
			"names":    len(names),
			"ids":      len(ids),
			"statuses": len(statuses),
		}).Warn("Could not fetch complete metadata for every repo.")
	}
	// Construct repos
	for i := range ids {
		name := safeAccess(names, i)
		id := safeAccess(ids, i)
		status := safeAccess(statuses, i)
		repo := repo{Name: name, ID: id, Status: status}
		repos = append(repos, repo)
	}
	return repos
}

// getAptrepos constructs repos from the sources.list file at path. Gives
// non-zero URLs
func getAptRepos() (repos []repo) {
	// getAptSources returns all the urls of all apt sources (including source
	// code repositories
	getAptSources := func() (urls []string) {
		otherLists := wrkutils.GetFilesWithExtension("/etc/apt/sources.list.d", ".list")
		sourceLists := append([]string{"/etc/apt/sources.list"}, otherLists...)
		for _, f := range sourceLists {
			split := tabular.ProbabalisticSplit(wrkutils.FileToString(f))
			// filter out comments
			commentRegex := regexp.MustCompile("^\\s*#")
			for _, line := range split {
				if len(line) > 1 && !(commentRegex.MatchString(line[0])) {
					urls = append(urls, line[1])
				}
			}
		}
		return urls
	}
	for _, src := range getAptSources() {
		repos = append(repos, repo{URL: src})
	}
	return repos
}

// getPacmanRepos constructs repos from the pacman.conf file at path. Gives
// non-zero Names and URLs
func getPacmanRepos(path string) (repos []repo) {
	data := wrkutils.FileToLines(path)
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
	if len(names) != len(urls) {
		log.WithFields(log.Fields{
			"names": len(names),
			"urls":  len(urls),
		}).Warn("Could not fetch complete metadata for every repo.")
	}
	for i, name := range names {
		if len(urls) > i {
			repos = append(repos, repo{Name: name, URL: urls[i]})
		} else {
			repos = append(repos, repo{Name: name})
		}
	}
	return repos
}

// getRepos simply returns a list of repos based on the manager chosen
func getRepos(manager string) (repos []repo) {
	switch manager {
	case "yum":
		return getYumRepos()
	case "apt":
		return getAptRepos()
	case "pacman":
		return getPacmanRepos("/etc/pacman.conf")
	default:
		log.WithFields(log.Fields{
			"attempted": manager,
			"supported": keys,
		}).Fatal("Cannot find repos of unsupported package manager")
	}
	return []repo{} // will never reach here b/c of default case
}

// existsRepoWithProperty is an abstraction of YumRepoExists and YumRepoURL.
// It takes a struct field name to check, and an expected value. If the expected
// value is found in the field of a repo, it returns 0, "" else an error message.
// Valid choices for prop: "URL" | "Name" | "Name"
func existsRepoWithProperty(prop string, val *regexp.Regexp, manager string) (int, string) {
	var properties []string
	for _, repo := range getRepos(manager) {
		switch prop {
		case "URL":
			properties = append(properties, repo.URL)
		case "Name":
			properties = append(properties, repo.Name)
		case "Status":
			properties = append(properties, repo.Status)
		case "ID":
			properties = append(properties, repo.ID)
		default:
			log.Fatal("Repos don't have the requested property: " + prop)
		}
	}
	if tabular.ReIn(val, properties) {
		return 0, ""
	}
	msg := "Repo with given " + prop + " not found"
	return wrkutils.GenericError(msg, val.String(), properties)
}

// repoExists checks to see that a given repo is listed in the appropriate
// configuration file
func repoExists(parameters []string) (exitCode int, exitMessage string) {
	re := wrkutils.ParseUserRegex(parameters[1])
	return existsRepoWithProperty("Name", re, parameters[0])
}

// repoExistsURI checks to see if the repo with the given URI is listed in the
// appropriate configuration file
func repoExistsURI(parameters []string) (exitCode int, exitMessage string) {
	re := wrkutils.ParseUserRegex(parameters[1])
	return existsRepoWithProperty("URL", re, parameters[0])
}

// pacmanIgnore checks to see whether a given package is in /etc/pacman.conf's
// IgnorePkg setting
func pacmanIgnore(parameters []string) (exitCode int, exitMessage string) {
	pkg := parameters[0]
	path := "/etc/pacman.conf"
	data := wrkutils.FileToString(path)
	re := regexp.MustCompile("[^#]IgnorePkg\\s+=\\s+.+")
	find := re.FindString(data)
	var packages []string
	if find != "" {
		spl := strings.Split(find, " ")
		wrkutils.IndexError("Not enough lines in "+path, 2, spl)
		packages = spl[2:] // first two are "IgnorePkg" and "="
		if tabular.StrIn(pkg, packages) {
			return 0, ""
		}
	}
	msg := "Couldn't find package in IgnorePkg"
	return wrkutils.GenericError(msg, pkg, packages)
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
