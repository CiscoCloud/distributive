package workers

import (
	"testing"
)

var fileParameters = [][]string{
	[]string{"/proc/net/tcp"},
	[]string{"/bin/bash"},
	[]string{"/proc/filesystems"},
	[]string{"/proc/uptime"},
	[]string{"/proc/cpuinfo"},
}

var dirParameters = [][]string{
	[]string{"/dev"},
	[]string{"/var"},
	[]string{"/tmp"},
	[]string{"/opt"},
	[]string{"/usr"},
	[]string{"/usr/bin"},
}

var symlinkParameters = [][]string{
	[]string{"/bin/sh"},
}

var notPaths = append(notLengthOne,
	[]string{}, []string{`\{{[(`}, []string{"", "", ""}, []string{"fail"},
	[]string{""}, []string{"7"},
)

func TestFile(t *testing.T) {
	t.Parallel()
	validInputs := append(fileParameters, dirParameters...)
	validInputs = append(validInputs, symlinkParameters...)
	invalidInputs := append(notPaths, names...)
	goodEggs := fileParameters
	badEggs := append(dirParameters, symlinkParameters...)
	testParameters(validInputs, invalidInputs, File{}, t)
	testCheck(goodEggs, badEggs, File{}, t)
}

func TestDirectory(t *testing.T) {
	t.Parallel()
	validInputs := append(fileParameters, dirParameters...)
	validInputs = append(validInputs, symlinkParameters...)
	invalidInputs := append(notPaths, names...)
	goodEggs := dirParameters
	badEggs := append(fileParameters, symlinkParameters...)
	testParameters(validInputs, invalidInputs, Directory{}, t)
	testCheck(goodEggs, badEggs, Directory{}, t)
}

func TestSymlink(t *testing.T) {
	t.Parallel()
	validInputs := append(fileParameters, dirParameters...)
	validInputs = append(validInputs, symlinkParameters...)
	invalidInputs := append(notPaths, names...)
	goodEggs := dirParameters
	badEggs := append(dirParameters, fileParameters...)
	testParameters(validInputs, invalidInputs, Symlink{}, t)
	testCheck(goodEggs, badEggs, Symlink{}, t)
}

// $1 - algorithm, $2 - check against, $3 - path
func TestChecksum(t *testing.T) {
	t.Parallel()
	validInputs := [][]string{
		[]string{"md5", "d41d8cd98f00b204e9800998ecf8427e", "/dev/null"},
		[]string{"sha1", "da39a3ee5e6b4b0d3255bfef95601890afd80709", "/dev/null"},
		[]string{"sha256", "chksum", "/proc/cpuinfo"},
		[]string{"sha512", "chksum", "/proc/cpuinfo"},
	}
	// generate losers from all files - none of them have that checksum
	invalidInputs := [][]string{
		[]string{}, []string{"", "", ""}, []string{"", ""},
		[]string{"sha256", "chksum", "/invalid/path"},
	}
	invalidInputs = append(invalidInputs, names...)
	goodEggs := [][]string{validInputs[0], validInputs[1]}
	badEggs := [][]string{validInputs[2], validInputs[3]}
	testParameters(validInputs, invalidInputs, Checksum{}, t)
	testCheck(goodEggs, badEggs, Checksum{}, t)
}

func TestFileMatches(t *testing.T) {
	t.Parallel()
	validInputs := appendParameter(fileParameters, "")
	invalidInputs := [][]string{
		[]string{"", ""}, []string{}, []string{"/notfile", "notmatch"},
	}
	invalidInputs = append(invalidInputs, names...)
	goodEggs := validInputs
	badEggs := [][]string{
		[]string{"/dev/null", "something"},
		[]string{"/proc/cpuinfo", "siddharthist"},
	}
	testParameters(validInputs, invalidInputs, FileMatches{}, t)
	testCheck(goodEggs, badEggs, FileMatches{}, t)
}

// $1 - path, $2 - givenMode (-rwxrwxrwx)
func TestPermissions(t *testing.T) {
	t.Parallel()
	valid1 := appendParameter(fileParameters, "----------")
	valid2 := appendParameter(dirParameters, "-rwxrwxrwx")
	valid3 := appendParameter(symlinkParameters, "-r--r--r--")
	validInputs := append(append(valid1, valid2...), valid3...)
	invalid1 := appendParameter(fileParameters, "nonsense")
	invalid2 := appendParameter(dirParameters, "-rrrwwwxxx")
	invalid3 := appendParameter(symlinkParameters, "")
	invalidInputs := append(append(invalid1, invalid2...), invalid3...)
	invalidInputs = append(invalidInputs, names...)
	goodEggs := [][]string{
		[]string{"/dev/null", "-rw-rw-rw-"},
		[]string{"/proc/", "-r-xr-xr-x"},
		[]string{"/bin/", "-rwxr-xr-x"},
	}
	badEggs := appendParameter(fileParameters, "----------")
	testParameters(validInputs, invalidInputs, Permissions{}, t)
	testCheck(goodEggs, badEggs, Permissions{}, t)
}
