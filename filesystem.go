package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"io/ioutil"
	"os"
	"strings"
)

type fileTypeCheck func(path string) (bool, error)

// isType checks if the resource at path is of the type specified by name by
// passing path to checker. Mostly used to abstract Directory, File, Symlink.
func isType(name string, checker fileTypeCheck, path string) (exitCode int, message string) {
	boo, err := checker(path)
	if os.IsNotExist(err) {
		return 1, "No such file or directory: " + path
	}
	if os.IsPermission(err) {
		return 1, "Insufficient permissions to read: " + path
	}
	if boo {
		return 0, ""
	}
	return 1, "Is not a " + name + ": " + path
}

// Exists checks to see if a file exists at given path
func File(path string) Thunk {
	// returns true if there is a regular ol' file at path
	isFile := func(path string) (bool, error) {
		fileInfo, err := os.Stat(path)
		if fileInfo.Mode().IsRegular() {
			return true, err
		}
		return false, err
	}

	return func() (exitCode int, exitMessage string) {
		return isType("file", isFile, path)
	}
}

// Directory checks to see if a directory exists at the specified path
func Directory(path string) Thunk {
	isDirectory := func(path string) (bool, error) {
		fileInfo, err := os.Stat(path)
		if fileInfo.Mode().IsDir() {
			return true, err
		}
		return false, err
	}
	return func() (exitCode int, exitMessage string) {
		return isType("directory", isDirectory, path)
	}
}

// Symlink checks to see if a symlink exists at a given path
func Symlink(path string) Thunk {
	// isSymlink checks to see if a symlink exists at this path.
	isSymlink := func(path string) (bool, error) {
		_, err := os.Readlink(path)
		if err == nil {
			return true, err
		}
		return false, err
	}
	return func() (exitCode int, exitMessage string) {
		return isType("symlink", isSymlink, path)
	}
}

// Checksum checks the hash of a given file using the given algorithm
func Checksum(algorithm string, checkAgainst string, path string) Thunk {
	getChecksum := func(algorithm string, data []byte) (checksum string) {
		algorithm = strings.ToUpper(algorithm)
		// default
		hasher := md5.New()
		switch algorithm {
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
		}
		hasher.Write(data)
		str := hex.EncodeToString(hasher.Sum(nil))
		return str

	}
	getFileChecksum := func(algorithm string, path string) (checksum string) {
		data, err := ioutil.ReadFile(path)
		fatal(err)
		return getChecksum(algorithm, data)
	}
	return func() (exitCode int, exitMessage string) {
		chksum := getFileChecksum(algorithm, path)
		if chksum == checkAgainst {
			return 0, ""
		}
		msg := "Checksums do not match for file: " + path + "\n"
		msg += "Given: " + checkAgainst + "\n"
		msg += "Calculated: " + chksum
		return 1, msg
	}
}
