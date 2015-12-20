// fsstatus provides utility functions for querying several aspects of the
// filesystem, especially as pertains to monitoring.
package fsstatus

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"github.com/CiscoCloud/distributive/tabular"
	"golang.org/x/crypto/sha3"
	"hash"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// IsFile checks to see if there's a regular ol' file at path.
func IsFile(path string) (bool, error) {
	if is, _ := IsSymlink(path); is {
		return false, nil
	}
	fileInfo, err := os.Stat(path)
	if fileInfo == nil || !fileInfo.Mode().IsRegular() {
		return false, err
	}
	return true, err
}

// IsDirectory checks to see if there's a regular ol' directory at path.
func IsDirectory(path string) (bool, error) {
	if is, _ := IsSymlink(path); is {
		return false, nil
	}
	fileInfo, err := os.Stat(path)
	if fileInfo == nil || !fileInfo.Mode().IsDir() {
		return false, err
	}
	return true, err
}

// IsSymlink checks to see if there's a symlink at path.
func IsSymlink(path string) (bool, error) {
	_, err := os.Readlink(path)
	if err == nil {
		return true, err
	}
	return false, err
}

// Checksum returns the checksum of some data, using a specified algorithm.
// It only returns an error when an invalid algorithm is used. The valid ones
// are MD5, SHA1, SHA224, SHA256, SHA384, SHA512, SHA3224, SHA3256, SHA3384,
// and SHA3512.
func Checksum(algorithm string, data []byte) (checksum string, err error) {
	// default
	var hasher hash.Hash
	switch strings.ToUpper(algorithm) {
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
	case "SHA3224":
		hasher = sha3.New224()
	case "SHA3256":
		hasher = sha3.New256()
	case "SHA3384":
		hasher = sha3.New384()
	case "SHA3512":
		hasher = sha3.New512()
	default:
		msg := "Invalid algorithm parameter passed go Checksum: %s"
		return checksum, fmt.Errorf(msg, algorithm)
	}
	hasher.Write(data)
	str := hex.EncodeToString(hasher.Sum(nil))
	return str, nil
}

// FileHasPermissions checks to see whether the file/directory/etc. at the given
// path has the given permissions (of the format -rwxrwxrwx)
func FileHasPermissions(expectedPerms string, path string) (bool, error) {
	finfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	actualMode := fmt.Sprint(finfo.Mode().Perm()) // -rwxrw-r-- format
	return (actualMode == expectedPerms), nil
}

// how many inodes are in this given state? state can be one of:
// total, used, free, percent
func inodesInState(filesystem, state string) (total uint64, err error) {
	cmd := exec.Command("df", "-i")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return total, err
	}
	table := tabular.ProbabalisticSplit(string(out))
	filesystems := tabular.GetColumnByHeader("Filesystem", table)
	var totals []string
	switch state {
	case "total":
		totals = tabular.GetColumnByHeader("Inodes", table)
	case "free":
		totals = tabular.GetColumnByHeader("IFree", table)
	case "used":
		totals = tabular.GetColumnByHeader("IUsed", table)
	case "percent":
		totals = tabular.GetColumnByHeader("IUse%", table)
	default:
		formatStr := "Internal error: unexpected state in inodesInState: %v"
		return total, fmt.Errorf(formatStr, state)
	}
	if len(filesystems) != len(totals) {
		formatStr := "The number of filesystems (%d) didn't match the number of"
		formatStr += " inode totals (%d)!"
		return total, fmt.Errorf(formatStr, len(filesystems), len(totals))
	}
	for i := range totals {
		if filesystems[i] == filesystem {
			// trim the % in case the state was "percent"
			return strconv.ParseUint(strings.TrimSuffix(totals[i], "%"), 10, 64)
		}
	}
	return total, fmt.Errorf("Couldn't find that filesystem: " + filesystem)
}

// FreeInodes reports the number of free inodes in a given filesystem, e.g.
// /dev/sda1, as given by `df -i`
func FreeInodes(filesystem string) (free uint64, err error) {
	return inodesInState(filesystem, "free")
}

// UsedInodes is like FreeInodes
func UsedInodes(filesystem string) (used uint64, err error) {
	return inodesInState(filesystem, "used")
}

// TotalInodes is like FreeInodes
func TotalInodes(filesystem string) (total uint64, err error) {
	return inodesInState(filesystem, "total")
}

// PercentInodesUsed reports the percentage given by `df -i`, which is ceilinged
// to the next integer value, so a percent like .001% would round to 1%.
func PercentInodesUsed(filesystem string) (percent uint8, err error) {
	percent64, err := inodesInState(filesystem, "percent")
	if err != nil {
		return percent, err
	}
	return uint8(percent64), err
}
