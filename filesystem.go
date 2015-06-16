package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
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

// File checks to see if the given path represents a normal file
func file(parameters []string) (exitCode int, exitMessage string) {
	// returns true if there is a regular ol' file at path
	isFile := func(path string) (bool, error) {
		fileInfo, err := os.Stat(path)
		if fileInfo.Mode().IsRegular() {
			return true, err
		}
		return false, err
	}
	return isType("file", isFile, parameters[0])
}

// Directory checks to see if a directory exists at the specified path
func directory(parameters []string) (exitCode int, exitMessage string) {
	isDirectory := func(path string) (bool, error) {
		fileInfo, err := os.Stat(path)
		if fileInfo.Mode().IsDir() {
			return true, err
		}
		return false, err
	}
	return isType("directory", isDirectory, parameters[0])
}

// Symlink checks to see if a symlink exists at a given path
func symlink(parameters []string) (exitCode int, exitMessage string) {
	// isSymlink checks to see if a symlink exists at this path.
	isSymlink := func(path string) (bool, error) {
		_, err := os.Readlink(path)
		if err == nil {
			return true, err
		}
		return false, err
	}
	return isType("symlink", isSymlink, parameters[0])
}

// Checksum checks the hash of a given file using the given algorithm
func Checksum(parameters []string) (exitCode int, exitMessage string) {
	// getChecksum returns the checksum of some data, using a specified
	// algorithm
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
	// getFileChecksum is self-explanatory
	getFileChecksum := func(algorithm string, path string) (checksum string) {
		return getChecksum(algorithm, fileToBytes(path))
	}

	algorithm := parameters[0]
	checkAgainst := parameters[1]
	path := parameters[2]
	chksum := getFileChecksum(algorithm, path)
	if chksum == checkAgainst {
		return 0, ""
	}
	msg := "Checksums do not match for file: " + path
	return genericError(msg, checkAgainst, []string{chksum})
}
