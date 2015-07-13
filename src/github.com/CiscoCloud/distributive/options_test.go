package main

import (
	"testing"
)

func TestValidateFlags(t *testing.T) {
	validFiles := []string{
		"/proc/fb", "/proc/dma", "/proc/kcore", "/proc/iomem", "/proc/stat",
	}
	validURLs := []string{
		"http://goo.co", "https://twitter.ca", "http://eff.org",
		"http://mozilla.org", "https://stackoverflow.com",
	}
	validDirs := []string{"/proc", "/bin", "/sbin", "/opt", "/home"}
	for i := 0; i < 5; i++ {
		validateFlags(validFiles[i], validURLs[i], validDirs[i])
	}
}

func TestInitializeLogrus(t *testing.T) {
	for _, lvl := range []string{"debug", "info", "warn", "fatal", "panic"} {
		initializeLogrus(lvl)
	}
}
