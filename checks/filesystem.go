package checks

import (
	"errors"
	"os"
	"regexp"
	"strings"

	"github.com/CiscoCloud/distributive/chkutil"
	"github.com/CiscoCloud/distributive/errutil"
	"github.com/CiscoCloud/distributive/fsstatus"
	"github.com/CiscoCloud/distributive/tabular"
	log "github.com/Sirupsen/logrus"
)

type fileCondition func(path string) (bool, error)

// isType checks if the resource at path is of the type specified by name by
// passing path to checker. Mostly used to abstract Directory, File, Symlink.
func isType(name string, checker fileCondition, path string) (int, string, error) {
	boo, err := checker(path)
	if os.IsNotExist(err) {
		return 1, "No such file or directory: " + path, nil
	} else if os.IsPermission(err) {
		return 1, "", errors.New("Insufficient Permissions to read: " + path)
	} else if boo {
		return errutil.Success()
	}
	return 1, "Is not a " + name + ": " + path, nil
}

/*
#### file
Description: Does this regular file exist?
Parameters:
- Path (filepath): Path to file
Example parameters:
- "/var/mysoftware/config.file", "/foo/bar/baz"
*/

type File struct{ path string }

func init() {
    chkutil.Register("File", func() chkutil.Check {
        return &File{}
    })
    chkutil.Register("Directory", func() chkutil.Check {
        return &Directory{}
    })
    chkutil.Register("Symlink", func() chkutil.Check {
        return &Symlink{}
    })
    chkutil.Register("Permissions", func() chkutil.Check {
        return &Permissions{}
    })
    chkutil.Register("Checksum", func() chkutil.Check {
        return &Checksum{}
    })
    chkutil.Register("FileMatches", func() chkutil.Check {
        return &FileMatches{}
    })
}

func (chk File) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	chk.path = params[0]
	return chk, nil
}

func (chk File) Status() (int, string, error) {
	return isType("file", fsstatus.IsFile, chk.path)
}

/*
#### directory
Description: Does this regular directory exist?
Parameters:
- Path (filepath): Path to directory
Example parameters:
- "/var/run/mysoftware.d/", "/foo/bar/baz/"
*/

type Directory struct{ path string }

func (chk Directory) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	chk.path = params[0]
	return chk, nil
}

func (chk Directory) Status() (int, string, error) {
	return isType("directory", fsstatus.IsDirectory, chk.path)
}

/*
#### symlink
Description: Does this symlink exist?
Parameters:
- Path (filepath): Path to symlink
Example parameters:
- "/var/run/mysoftware.d/", "/foo/bar/baz", "/bin/sh"
*/

type Symlink struct{ path string }

func (chk Symlink) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	chk.path = params[0]
	return chk, nil
}

func (chk Symlink) Status() (int, string, error) {
	return isType("symlink", fsstatus.IsSymlink, chk.path)
}

/*
#### checksum
Description: Does this file match the expected checksum when using the specified
algorithm?
Parameters:
- Algorithm (string): MD5 | SHA1 | SHA224 | SHA256 | SHA384 | SHA512 |
SHA3224 | SHA3256 | SHA3384 | SHA3512
- Expected checksum (checksum/string)
- Path (filepath): Path to file to check the checksum of
Example parameters:
- MD5, SHA1, SHA224, SHA256, SHA384, SHA512, SHA3224, SHA3256, SHA3384,
- d41d8cd98f00b204e9800998ecf8427e, c6cf669dbd4cf2fbd59d03cc8039420a48a037fe
- /dev/null, /etc/config/important-file.conf
*/

type Checksum struct{ algorithm, expectedChksum, path string }

func (chk Checksum) ID() string { return "checksum" }

func (chk Checksum) New(params []string) (chkutil.Check, error) {
	if len(params) != 3 {
		return chk, errutil.ParameterLengthError{3, params}
	}
	valid := []string{"MD5", "SHA1", "SHA224", "SHA256", "SHA384", "SHA512"}
	if !tabular.StrIn(strings.ToUpper(params[0]), valid) {
		return chk, errutil.ParameterTypeError{params[0], "algorithm"}
	}
	chk.algorithm = params[0]
	path := params[2]
	chk.path = path
	// TODO validate length of checksum string
	chk.expectedChksum = params[1]
	return chk, nil
}

func (chk Checksum) Status() (int, string, error) {
	if _, err := os.Stat(chk.path); err != nil {
		return 2, "", err
	}

	// getFileChecksum is self-explanatory
	fileChecksum := func(algorithm string, path string) string {
		if path == "" {
			log.Fatal("getFileChecksum got a blank path")
		} else if _, err := os.Stat(chk.path); err != nil {
			log.WithFields(log.Fields{
				"path": chk.path,
			}).Fatal("fileChecksum got an invalid path")
		}
		// we already validated the aglorithm
		chksum, _ := fsstatus.Checksum(algorithm, chkutil.FileToBytes(path))
		return chksum
	}
	actualChksum := fileChecksum(chk.algorithm, chk.path)
	if actualChksum == chk.expectedChksum {
		return errutil.Success()
	}
	msg := "Checksums do not match for file: " + chk.path
	return errutil.GenericError(msg, chk.expectedChksum, []string{actualChksum})
}

/*
#### FileMatches
Description: Does this file match this regexp?
Parameters:
- Path (filepath): Path to file to check the contents of
- Regexp (regexp): Regexp to query file with
Example parameters:
- /dev/null, /etc/config/important-file.conf
- "str", "myvalue=expected", "IP=\d{1,3}.\d{1,3}.\d{1,3}.\d{1,3}"
*/

type FileMatches struct {
	path string
	re   *regexp.Regexp
}

func (chk FileMatches) New(params []string) (chkutil.Check, error) {
	if len(params) != 2 {
		return chk, errutil.ParameterLengthError{2, params}
	}
	re, err := regexp.Compile(params[1])
	if err != nil {
		return chk, errutil.ParameterTypeError{params[1], "regexp"}
	}
	chk.re = re
	path := params[0]
	chk.path = path
	return chk, nil
}

func (chk FileMatches) Status() (int, string, error) {
	if _, err := os.Stat(chk.path); err != nil {
		return 2, "", err
	}
	if chk.re.Match(chkutil.FileToBytes(chk.path)) {
		return errutil.Success()
	}
	msg := "File does not match regexp:"
	msg += "\n\tFile: " + chk.path
	msg += "\n\tRegexp: " + chk.re.String()
	return 1, msg, nil
}

/*
#### Permissions
Description: Does this file have the given Permissions?
Parameters:
- Path (filepath): Path to file to check the Permissions of
- Mode (filemode): Filemode to expect
Example parameters:
- /dev/null, /etc/config/important-file.conf
- -rwxrwxrwx, -rw-rw---- -rw-------, -rwx-r-x-r-x
*/

type Permissions struct{ path, expectedPerms string }

func (chk Permissions) New(params []string) (chkutil.Check, error) {
	if len(params) != 2 {
		return chk, errutil.ParameterLengthError{2, params}
	}
	mode := params[1]
	modeRe := `-([r-][w-][x-]){3}`
	if len(mode) != 10 || !regexp.MustCompile(modeRe).MatchString(mode) {
		return chk, errutil.ParameterTypeError{mode, "filemode"}
	}
	chk.path = params[0]
	chk.expectedPerms = mode
	return chk, nil
}

func (chk Permissions) Status() (int, string, error) {
	if _, err := os.Stat(chk.path); err != nil {
		return 1, "", err
	}
	passed, err := fsstatus.FileHasPermissions(chk.expectedPerms, chk.path)
	if err != nil {
		return 1, "", err
	}
	if passed {
		return errutil.Success()
	}
	return 1, "File did not have permissions: " + chk.expectedPerms, nil
}
