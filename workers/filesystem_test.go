package workers

import (
	"testing"
)

var fileParameters = []parameters{
	[]string{"/proc/net/tcp"},
	[]string{"/bin/bash"},
	[]string{"/proc/filesystems"},
	[]string{"/proc/uptime"},
	[]string{"/proc/cpuinfo"},
}

var dirParameters = []parameters{
	[]string{"/dev"},
	[]string{"/var"},
	[]string{"/tmp"},
	[]string{"/opt"},
	[]string{"/usr"},
	[]string{"/usr/bin"},
}

var symlinkParameters = []parameters{
	[]string{"/bin/sh"},
}

func TestFile(t *testing.T) {
	losers := append(dirParameters, symlinkParameters...)
	testInputs(t, file, fileParameters, losers)
}

func TestDirectory(t *testing.T) {
	t.Parallel()
	losers := append(fileParameters, symlinkParameters...)
	testInputs(t, directory, dirParameters, losers)
}

func TestSymlink(t *testing.T) {
	t.Parallel()
	losers := append(fileParameters, dirParameters...)
	testInputs(t, symlink, symlinkParameters, losers)
}

// $1 - algorithm, $2 - check against, $3 - path
func TestChecksum(t *testing.T) {
	t.Parallel()
	winners := []parameters{
		[]string{"md5", "d41d8cd98f00b204e9800998ecf8427e", "/dev/null"},
	}
	// generate losers from all files - none of them have that checksum
	losers := []parameters{}
	for _, f := range fileParameters {
		loser := []string{"md5", "00000000000000000000000000000000", f[0]}
		losers = append(losers, loser)
	}
	testInputs(t, checksum, winners, losers)
}

func TestFileContains(t *testing.T) {
	t.Parallel()
	winners := []parameters{
		[]string{"/dev/null", ""},
	}
	losers := []parameters{
		[]string{"/dev/null", "fail"},
	}
	testInputs(t, fileContains, winners, losers)
}

// $1 - path, $2 - givenMode (-rwxrwxrwx)
func TestPermissions(t *testing.T) {
	t.Parallel()
	winners := []parameters{
		[]string{"/dev/null", "-rw-rw-rw-"},
		[]string{"/proc/", "-r-xr-xr-x"},
		[]string{"/bin/", "-rwxr-xr-x"},
	}
	losers := []parameters{
		[]string{"/dev/null", "----------"},
		[]string{"/proc/", "----------"},
		[]string{"/bin/", "----------"},
	}
	testInputs(t, permissions, winners, losers)
}

// $1 - path, $2 maxpercent
func TestDiskUsage(t *testing.T) {
	t.Parallel()
	winners := []parameters{[]string{"/", "99"}, []string{"/", "98"}}
	losers := []parameters{[]string{"/", "1"}, []string{"/", "2"}}
	testInputs(t, diskUsage, winners, losers)
}
