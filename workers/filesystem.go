package workers

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/CiscoCloud/distributive/chkutil"
	"github.com/CiscoCloud/distributive/errutil"
	"github.com/CiscoCloud/distributive/tabular"
	log "github.com/Sirupsen/logrus"
	"hash"
	"os"
	"regexp"
	"strings"
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

func (chk File) ID() string { return "File" }

func (chk File) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	chk.path = params[0]
	return chk, nil
}

func (chk File) Status() (int, string, error) {
	return isType("file", isFile, chk.path)
}

// returns true if there is a regular ol' file at path
func isFile(path string) (bool, error) {
	if is, _ := isSymlink(path); is {
		return false, nil
	}
	fileInfo, err := os.Stat(path)
	if fileInfo == nil || !fileInfo.Mode().IsRegular() {
		return false, err
	}
	return true, err
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

func (chk Directory) ID() string { return "Directory" }

func (chk Directory) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	chk.path = params[0]
	return chk, nil
}

func (chk Directory) Status() (int, string, error) {
	return isType("directory", isDirectory, chk.path)
}

// returns true if there is a regular ol' directory at path
func isDirectory(path string) (bool, error) {
	if is, _ := isSymlink(path); is {
		return false, nil
	}
	fileInfo, err := os.Stat(path)
	if fileInfo == nil || !fileInfo.Mode().IsDir() {
		return false, err
	}
	return true, err
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

func (chk Symlink) ID() string { return "Symlink" }

func (chk Symlink) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	chk.path = params[0]
	return chk, nil
}

func (chk Symlink) Status() (int, string, error) {
	return isType("symlink", isSymlink, chk.path)
}

func isSymlink(path string) (bool, error) {
	_, err := os.Readlink(path)
	if err == nil {
		return true, err
	}
	return false, err
}

/*
#### checksum
Description: Does this file match the expected checksum when using the specified
algorithm?
Parameters:
  - Algorithm (string): MD5 | SHA1 | SHA224 | SHA256 | SHA384 | SHA512
  - Expected checksum (checksum/string)
  - Path (filepath): Path to file to check the checksum of
Example parameters:
  - SHA1, SHA224, SHA256, SHA384, SHA512
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
	if _, err := os.Stat(path); err != nil {
		return chk, errutil.ParameterTypeError{path, "filepath"}
	}
	chk.path = path
	// TODO validate length of checksum string
	chk.expectedChksum = params[1]
	return chk, nil
}

func (chk Checksum) Status() (int, string, error) {
	// getChecksum returns the checksum of some data, using a specified
	// algorithm
	getChecksum := func(algorithm string, data []byte) (checksum string) {
		algorithm = strings.ToUpper(algorithm)
		// default
		var hasher hash.Hash
		switch algorithm {
		case "MD5":
			hasher = md5.New()
		case "SHA1":
			hasher = sha1.New()
		case "SHA224":
			hasher = sha256.New224()
		case "SHA256":
			hasher = sha256.New()
		case "SHA384":
			hasher = sha512.New384()
		case "SHA512":
			hasher = sha512.New()
		default:
			log.WithFields(log.Fields{
				"given":    algorithm,
				"expected": "MD5 | SHA1 | SHA224 | SHA256 | SHA384 | SHA512",
			}).Fatal("Invalid algorithm parameter passed to getChecksum")
		}
		hasher.Write(data)
		str := hex.EncodeToString(hasher.Sum(nil))
		return str

	}
	// getFileChecksum is self-explanatory
	getFileChecksum := func(algorithm string, path string) (checksum string) {
		if path == "" {
			log.Fatal("getFileChecksum got a blank path")
		} else if _, err := os.Stat(chk.path); err != nil {
			log.WithFields(log.Fields{
				"path": chk.path,
			}).Fatal("getFileChecksum got an invalid path")
		}
		return getChecksum(algorithm, chkutil.FileToBytes(path))
	}
	actualChksum := getFileChecksum(chk.algorithm, chk.path)
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

func (chk FileMatches) ID() string { return "FileMatches" }

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
	if _, err := os.Stat(path); err != nil {
		return chk, errutil.ParameterTypeError{path, "filepath"}
	}
	chk.path = path
	return chk, nil
}

func (chk FileMatches) Status() (int, string, error) {
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

func (chk Permissions) ID() string { return "Permissions" }

func (chk Permissions) New(params []string) (chkutil.Check, error) {
	if len(params) != 2 {
		return chk, errutil.ParameterLengthError{2, params}
	}
	if _, err := os.Stat(params[0]); err != nil {
		return chk, errutil.ParameterTypeError{params[0], "file"}
	}
	mode := params[1]
	modeRe := `-([r-][w-][x-]){3}`
	if len(mode) != 10 || !regexp.MustCompile(modeRe).MatchString(mode) {
		log.Debug("Did not match regexp " + modeRe) // TODO remove
		return chk, errutil.ParameterTypeError{mode, "filemode"}
	}
	chk.path = params[0]
	chk.expectedPerms = mode
	return chk, nil
}

func (chk Permissions) Status() (int, string, error) {
	finfo, err := os.Stat(chk.path)
	if err != nil {
		errutil.CouldntReadError(chk.path, err)
	}
	actualMode := fmt.Sprint(finfo.Mode().Perm()) // -rwxrw-r-- format
	if actualMode == chk.expectedPerms {
		return errutil.Success()
	}
	msg := "File modes did not match"
	return errutil.GenericError(msg, chk.expectedPerms, []string{actualMode})
}
