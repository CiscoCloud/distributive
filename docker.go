package main

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"
)

// separateString is an abstraction of stringToSlice that takes two kinds of
// separators, and splits a string into a 2D slice based on those separators
func separateString(rowSep *regexp.Regexp, colSep *regexp.Regexp, str string) (output [][]string) {
	lines := rowSep.Split(str, -1)
	for _, line := range lines {
		output = append(output, colSep.Split(line, -1))
	}
	return output
}

// stringToSlice takes in a string and returns a 2D slice of its output,
// separated on whitespace and newlines
func stringToSlice(str string) (output [][]string) {
	rowSep := regexp.MustCompile("\n+")
	colSep := regexp.MustCompile("\\s+")
	return separateString(rowSep, colSep, str)
}

// getColumn isolates the entries of a single column from a 2D slice
func getColumn(col int, slc [][]string) (column []string) {
	for _, line := range slc {
		if len(line) > col {
			column = append(column, line[col])
		}
	}
	return column
}

// strIn checks to see if a given string is in a slice of strings
func strIn(str string, slice []string) bool {
	for _, sliceString := range slice {
		if str == sliceString {
			return true
		}
	}
	return false
}

// DockerImage ensures that a specified docker image (e.g. "user/image",
// "ubuntu", etc.) is downloaded (pulled) on the host
func DockerImage(name string) Thunk {
	getDockerImages := func() (images []string) {
		out, err := exec.Command("docker", "images").CombinedOutput()
		outstr := string(out)
		// `docker images` requires root permissions
		if strings.Contains(outstr, "permission denied") {
			log.Fatal("Permission denied when running: docker images")
		}
		fatal(err)
		lines := stringToSlice(outstr)
		// there are no images
		if len(lines) < 1 {
			return []string{}
		}
		// skip header row, get first (image) column
		return getColumn(0, lines[1:])
	}
	return func() (exitCode int, exitMessage string) {
		if strIn(name, getDockerImages()) {
			return 0, ""
		}
		return 1, "Docker image was not found: " + name
	}
}

// DockerRunning checks to see if a specified docker container is running
// (e.g. "user/container")
func DockerRunning(name string) Thunk {
	getRunningContainers := func() (images []string) {
		out, err := exec.Command("docker", "ps", "-a").CombinedOutput()
		outstr := string(out)
		// `docker images` requires root permissions
		if strings.Contains(outstr, "permission denied") {
			log.Fatal("Permission denied when running: docker ps -a")
		}
		fatal(err)
		// get first column of output, as this is the image column
		rowSep := regexp.MustCompile("\n+")
		colSep := regexp.MustCompile("\\s{2,}")
		lines := separateString(rowSep, colSep, outstr)
		if len(lines) < 1 {
			return []string{}
		}
		lines = lines[1:] // ignore headers
		names := getColumn(1, lines)
		statuses := getColumn(4, lines)
		for i, status := range statuses {
			if strings.Contains(status, "Up") && len(names) > i {
				images = append(images, names[i])
			}
		}
		fmt.Println(images)
		return images
	}
	return func() (exitCode int, exitMessage string) {
		fmt.Println(getRunningContainers())
		if strIn(name, getRunningContainers()) {
			return 0, ""
		}
		return 1, "Docker container not runnning: " + name
	}
}
