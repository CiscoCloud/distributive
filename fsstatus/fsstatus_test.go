package fsstatus

import (
	"testing"
)

var fileParameters = []string{
	"/proc/net/tcp",
	"/bin/bash",
	"/proc/filesystems",
	"/proc/uptime",
	"/proc/cpuinfo",
}

var dirParameters = []string{
	"/dev", "/var", "/tmp", "/opt", "/usr", "/usr/bin",
}

var symlinkParameters = []string{"/bin/sh"}

type fileTest func(path string) (bool, error)

// abstraction of TestIsFile, TestIsDir, TestIsSymlink
func testIsType(f fileTest, goodEggs, badEggs []string, t *testing.T) {
	for _, egg := range goodEggs {
		if is, _ := f(egg); !is {
			t.Errorf("Test reported incorrectly for %s", egg)
		}
	}
	for _, egg := range badEggs {
		if is, _ := f(egg); is {
			t.Errorf("Test reported incorrectly for %s", egg)
		}
	}
}

func TestIsFile(t *testing.T) {
	t.Parallel()
	badEggs := append(dirParameters, symlinkParameters...)
	testIsType(IsFile, fileParameters, badEggs, t)
}

func TestIsDirectory(t *testing.T) {
	t.Parallel()
	badEggs := append(fileParameters, symlinkParameters...)
	testIsType(IsDirectory, dirParameters, badEggs, t)
}

func TestIsSymlink(t *testing.T) {
	t.Parallel()
	badEggs := append(fileParameters, dirParameters...)
	testIsType(IsSymlink, symlinkParameters, badEggs, t)
}

func TestChecksum(t *testing.T) {
	t.Parallel()
	lst := [][]string{
		{"", "md5", "d41d8cd98f00b204e9800998ecf8427e"},
		{"", "sha1", "da39a3ee5e6b4b0d3255bfef95601890afd80709"},
		{"", "sha224", "d14a028c2a3a2bc9476102bb288234c415a2b01f828ea62ac5b3e42f"},
		{"", "sha256", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{"", "sha384", "38b060a751ac96384cd9327eb1b1e36a21fdb71114be07434c0cc7bf63f6e1da274edebfe76f65fbd51ad2f14898b95b"},
		{"", "sha512", "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e"},
	}
	for _, trio := range lst {
		actual, _ := Checksum(trio[1], []byte(trio[0]))
		if actual != trio[2] {
			t.Error("Data did not have expected checksum: %s", trio[0])
		}
	}
}
