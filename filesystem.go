// filesystem.go provides filesystem related thunks.
package main

import (
	"os"
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
