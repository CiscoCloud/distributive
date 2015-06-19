package main

import (
	"log"
	"os/exec"
	"strings"
)

// getDockerImages returns a list of all downloaded Docker images
func getDockerImages() (images []string) {
	cmd := exec.Command("docker", "images")
	return commandColumnNoHeader(0, cmd)
}

// dockerImage checks to see that the specified Docker image (e.g. "user/image",
// "ubuntu", etc.) is downloaded (pulled) on the host
func dockerImage(parameters []string) (exitCode int, exitMessage string) {
	name := parameters[0]
	images := getDockerImages()
	if strIn(name, images) {
		return 0, ""
	}
	return genericError("Docker image was not found", name, images)
}

// dockerImageRegexp is like dockerImage, but with a regexp match instead
func dockerImageRegexp(parameters []string) (exitCode int, exitMessage string) {
	re := parseUserRegex(parameters[0])
	images := getDockerImages()
	if reIn(re, images) {
		return 0, ""
	}
	return genericError("Docker image was not found", re.String(), images)
}

// getRunningContainers returns a list of names of running docker containers
func getRunningContainers() (images []string) {
	out, err := exec.Command("docker", "ps", "-a").CombinedOutput()
	outstr := string(out)
	// `docker images` requires root permissions
	if err != nil && strings.Contains(outstr, "permission denied") {
		log.Fatal("Permission denied when running: docker ps -a")
	}
	if err != nil {
		log.Fatal("Error while running `docker ps -a`" + "\n\t" + err.Error())
	}
	// the output of `docker ps -a` has spaces in columns, but each column
	// is separated by 2 or more spaces
	lines := stringToSliceMultispace(outstr)
	if len(lines) < 1 {
		return []string{}
	}
	names := getColumnNoHeader(1, lines)    // all docker container names
	statuses := getColumnNoHeader(4, lines) // all docker container statuses
	for i, status := range statuses {
		if strings.Contains(status, "Up") && len(names) > i {
			images = append(images, names[i])
		}
	}
	return images
}

// dockerRunning checks to see if a specified docker container is running
// (e.g. "user/container")
func dockerRunning(parameters []string) (exitCode int, exitMessage string) {
	name := parameters[0]
	running := getRunningContainers()
	if strContainedIn(name, running) {
		return 0, ""
	}
	return genericError("Docker container not runnning", name, running)
}

// dockerRunningRegexp is like dockerRunning, but with a regexp match instead
func dockerRunningRegexp(parameters []string) (exitCode int, exitMessage string) {
	re := parseUserRegex(parameters[0])
	running := getRunningContainers()
	if reIn(re, running) {
		return 0, ""
	}
	return genericError("Docker container not runnning", re.String(), running)
}
