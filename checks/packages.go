package checks

import (
	"github.com/CiscoCloud/distributive/chkutil"
	"github.com/CiscoCloud/distributive/errutil"
	"github.com/CiscoCloud/distributive/tabular"
	log "github.com/Sirupsen/logrus"
	"os/exec"
	"regexp"
	"strings"
)

// TODO implement package managers as interfaces
// with fields getRepos, getInstalledPackages, queryInstalled

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
		// TODO replace with golang cmd in path
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

// repo is a unified interface for pacman, dpkg, and rpm repos
type repo struct {
	ID     string
	Name   string // rpm
	URL    string // dpkg, pacman
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
				errutil.IndexError(msg, index, slc)
			}
		}() // invoke inside defer
		if len(slc) > index {
			return slc[index]
		}
		return ""
	}

	// get and parse output of `yum repolist`
	cmd := exec.Command("yum", "repolist")
	outstr := chkutil.CommandOutput(cmd)
	slc := tabular.ProbabalisticSplit(outstr)

	ids := tabular.GetColumnNoHeader(0, slc) // TODO use columnbyheader here
	errutil.IndexError("getYumRepos", 2, ids)
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
		otherLists := chkutil.GetFilesWithExtension("/etc/apt/sources.list.d", ".list")
		sourceLists := append([]string{"/etc/apt/sources.list"}, otherLists...)
		for _, f := range sourceLists {
			split := tabular.ProbabalisticSplit(chkutil.FileToString(f))
			// filter out comments
			commentRegex := regexp.MustCompile(`^\s*#`)
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
	data := chkutil.FileToLines(path)
	// match words and dashes in brackets without comments
	nameRegex := regexp.MustCompile(`[^#]\[(\w|\-)+\]`)
	// match lines that start with Include= or Server= and anything after that
	urlRegex := regexp.MustCompile(`[^#](Include|Server)\s*=\s*.*`)
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
	case "rpm":
		return getYumRepos()
	case "dpkg":
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
func existsRepoWithProperty(prop string, val *regexp.Regexp, manager string) (int, string, error) {
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
		return errutil.Success()
	}
	msg := "Repo with given " + prop + " not found"
	return errutil.GenericError(msg, val.String(), properties)
}

/*
#### RepoExists
Description: Is this repo present?
Parameters:
  - Pakage manager: rpm | dpkg | pacman
  - Regexp (regexp): Regexp to match the name of the repo
Example parameters:
  - "base", "firefox-[nN]ightly"
*/

type RepoExists struct {
	manager string
	re      *regexp.Regexp
}

func (chk RepoExists) ID() string { return "RepoExists" }

func (chk RepoExists) New(params []string) (chkutil.Check, error) {
	if len(params) != 2 {
		return chk, errutil.ParameterLengthError{2, params}
	}
	re, err := regexp.Compile(params[1])
	if err != nil {
		return chk, errutil.ParameterTypeError{params[1], "regexp"}
	}
	chk.re = re
	if !tabular.StrIn(params[0], keys) {
		return chk, errutil.ParameterTypeError{params[0], "package manager"}
	}
	chk.manager = params[0]
	return chk, nil
}

func (chk RepoExists) Status() (int, string, error) {
	return existsRepoWithProperty("Name", chk.re, chk.manager)
}

/*
#### RepoExistsURI
Description: Is a repo with this URI present?
Parameters:
  - Pakage manager: rpm | dpkg | pacman
  - Regexp (regexp): Regexp to match the URI of the repo
Example parameters:
  - "http://my-repo.example.com", "/path/to/repo"
Depedencies:
  - dpkg | pacman
*/

type RepoExistsURI struct {
	manager string
	re      *regexp.Regexp
}

func (chk RepoExistsURI) ID() string { return "RepoExistsURI" }

func (chk RepoExistsURI) New(params []string) (chkutil.Check, error) {
	if len(params) != 2 {
		return chk, errutil.ParameterLengthError{2, params}
	}
	re, err := regexp.Compile(params[1])
	if err != nil {
		return chk, errutil.ParameterTypeError{params[1], "regexp"}
	}
	chk.re = re
	if !tabular.StrIn(params[0], keys) {
		return chk, errutil.ParameterTypeError{params[0], "package manager"}
	}
	chk.manager = params[0]
	return chk, nil
}

func (chk RepoExistsURI) Status() (int, string, error) {
	return existsRepoWithProperty("URL", chk.re, chk.manager)
}

/*
#### PacmanIgnore
Description: Are upgrades to this package ignored by pacman?
Parameters:
  - Package (string): Name of the package
Example parameters:
  - node, python, etcd
Depedencies:
  - pacman, specifically /etc/pacman.conf
*/

type PacmanIgnore struct{ pkg string }

func (chk PacmanIgnore) ID() string { return "PacmanIgnore" }

func (chk PacmanIgnore) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	chk.pkg = params[0]
	return chk, nil
}

func (chk PacmanIgnore) Status() (int, string, error) {
	path := "/etc/pacman.conf"
	data := chkutil.FileToString(path)
	re := regexp.MustCompile(`[^#]IgnorePkg\s+=\s+.+`)
	find := re.FindString(data)
	var packages []string
	if find != "" {
		spl := strings.Split(find, " ")
		errutil.IndexError("Not enough lines in "+path, 2, spl)
		packages = spl[2:] // first two are "IgnorePkg" and "="
		if tabular.StrIn(chk.pkg, packages) {
			return errutil.Success()
		}
	}
	msg := "Couldn't find package in IgnorePkg"
	return errutil.GenericError(msg, chk.pkg, packages)
}

/*
#### Installed
Description: Is this package Installed?
Parameters:
  - Package (string): Name of the package
Example parameters:
  - node, python, etcd
Depedencies:
  - pacman | dpkg | rpm
*/

type Installed struct{ pkg string }

func (chk Installed) ID() string { return "Installed" }

func (chk Installed) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	chk.pkg = params[0]
	return chk, nil
}

func (chk Installed) Status() (int, string, error) {
	name := getManager()
	options := managers[name]
	cmd := exec.Command(name, options, chk.pkg)
	out, err := cmd.CombinedOutput()
	outstr := string(out)
	if strings.Contains(outstr, chk.pkg) {
		return errutil.Success()
	} else if outstr != "" && err != nil {
		// pacman - outstr == "", and exit code of 1 if it can't find the pkg
		errutil.ExecError(cmd, outstr, err)
	}
	msg := "Package was not found:"
	msg += "\n\tPackage name: " + chk.pkg
	msg += "\n\tPackage manager: " + name
	msg += "\n\tCommand output: " + outstr
	return 1, msg, nil

}
