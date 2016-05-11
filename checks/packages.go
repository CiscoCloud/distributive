package checks

import (
	"regexp"
	"strings"

	"github.com/CiscoCloud/distributive/chkutil"
	"github.com/CiscoCloud/distributive/errutil"
	"github.com/CiscoCloud/distributive/tabular"
	gossresource "github.com/aelsabbahy/goss/resource"
	gosssystem "github.com/aelsabbahy/goss/system"
	gossutil "github.com/aelsabbahy/goss/util"
)

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

func init() {
    chkutil.Register("PacmanIgnore", func() chkutil.Check {
        return &PacmanIgnore{}
    })
    chkutil.Register("Installed", func() chkutil.Check {
        return &Installed{}
    })
}

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
- pacman | dpkg | rpm | apk
*/

type Installed struct{ pkg string }

func (chk Installed) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	chk.pkg = params[0]
	return chk, nil
}

func (chk Installed) Status() (int, string, error) {
	var pkg gosssystem.Package
	switch gosssystem.DetectPackageManager() {
	case "deb":
		pkg = gosssystem.NewDebPackage(chk.pkg, nil, gossutil.Config{})
	case "apk":
		pkg = gosssystem.NewAlpinePackage(chk.pkg, nil, gossutil.Config{})
	case "pacman":
		pkg = gosssystem.NewPacmanPackage(chk.pkg, nil, gossutil.Config{})
	default:
		pkg = gosssystem.NewRpmPackage(chk.pkg, nil, gossutil.Config{})
	}
	// initialize the package
	pkg2, err := gossresource.NewPackage(pkg, gossutil.Config{})
	if err != nil {
		return 1, "", err
	} else if pkg2.Installed {
		return errutil.Success()
	}
	return 1, "Package not found", nil
}
