// Package dockerstatus provides a few functions for getting very simple data
// out of Docker, mostly for use in simple status checks.
package dockerstatus

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/CiscoCloud/distributive/chkutil"
	"github.com/CiscoCloud/distributive/tabular"
)

var dockerCmdTimeout, _ = time.ParseDuration("10s")

// DockerImageRepositories returns a slice of the names of the Docker images
// present on the host (what's under the REPOSITORIES column of `docker images`)
func DockerImageRepositories() (images []string, err error) {
	cmd := exec.Command("docker", "images")
	out, err := chkutil.CommandTimeout(cmd, dockerCmdTimeout)
	if err != nil {
		// try escalating to sudo, the error might have been one of permissions
		cmd = exec.Command("sudo", "docker", "images")
		out, err = chkutil.CommandTimeout(cmd, dockerCmdTimeout)
		if err != nil {
			return images, err
		}
	}
	table := tabular.ProbabalisticSplit(string(out))
	return tabular.GetColumnByHeader("REPOSITORY", table), nil
}

// parse the output of `docker ps -a` into a list of running container names
func parseRunningContainers(output string) (containers []string) {
	output = "IMAGE\tSTATUS\tNAMES\n" + output
	rowRegexp := regexp.MustCompile(`\n`)
	columnRegexp := regexp.MustCompile(`\t`)
	lines := tabular.SeparateString(rowRegexp, columnRegexp, output)
	names := tabular.GetColumnByHeader("IMAGE", lines)
	statuses := tabular.GetColumnByHeader("STATUS", lines)
	for i, status := range statuses {
		// index error caught by second condition in if clause
		if strings.Contains(status, "Up") && len(names) > i {
			containers = append(containers, names[i])
		}
	}
	return containers
}

// RunningContainers returns a list of names of running docker containers
// (what's under the IMAGE column of `docker ps -a` if it has status "Up").
func RunningContainers() (containers []string, err error) {
	outputFormat := `{{.Image}}\t{{.Status}}\t{{.Names}}`
	cmd := exec.Command("docker", "ps", "-a", "--format", outputFormat)
	out, err := chkutil.CommandTimeout(cmd, dockerCmdTimeout)
	if err != nil {
		cmd = exec.Command("sudo", "docker", "ps", "-a", "--format", outputFormat)
		out, err = chkutil.CommandTimeout(cmd, dockerCmdTimeout)
		if err != nil {
			return []string{}, err
		}
	} else if out == "" {
		return []string{}, fmt.Errorf("Command produced no output: %v", cmd.Args)
	}
	return parseRunningContainers(string(out)), nil
}

// DaemonResponding checks to see if the Docker daemon responds to commands
// within the given timeout. If everything goes well, it returns nil.
func DaemonResponding(timeout time.Duration) error {
	_, err := chkutil.CommandTimeout(exec.Command("docker", "ps"), timeout)
	return err
}
