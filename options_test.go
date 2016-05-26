package main

import (
	"testing"
)

func TestValidateFlags(t *testing.T) {
	validFiles := []string{
		"/proc/net/tcp", "/bin/sh", "/proc/filesystems",
		"/proc/uptime", "/proc/cpuinfo",
	}
	validURLs := []string{
		"http://goo.co", "https://twitter.ca", "http://eff.org",
		"http://mozilla.org", "https://stackoverflow.com",
	}
	validDirs := []string{"/dev", "/var", "/tmp", "/usr", "/usr/bin"}
	for i := 0; i < 5; i++ {
		validateFlags(validFiles[i], validURLs[i], validDirs[i])
	}
}

func TestInitializeLogrus(t *testing.T) {
	for _, lvl := range []string{"debug", "info", "warn", "fatal", "panic"} {
		initializeLogrus(lvl)
	}
}
