package checks

import (
	"github.com/CiscoCloud/distributive/chkutil"
	"github.com/CiscoCloud/distributive/dockerstatus"
	"github.com/CiscoCloud/distributive/errutil"
	"github.com/CiscoCloud/distributive/tabular"
	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"os"
	"regexp"
	"strings"
)

/*
#### DockerImage
Description: Is this Docker image present?
Parameters:
  - Name (string):  Name of the image
Example parameters:
  - "user/image", "ubuntu"
*/

type DockerImage struct{ name string }

func init() { 
    chkutil.Register("DockerImage", func() chkutil.Check {
        return &DockerImage{}
    })
    chkutil.Register("DockerRunning", func() chkutil.Check {
        return &DockerRunning{}
    })
    chkutil.Register("DockerRunningRegext", func() chkutil.Check {
        return &DockerRunningRegexp{}
    })
}

func (chk DockerImage) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	chk.name = params[0]
	return chk, nil
}

func (chk DockerImage) Status() (int, string, error) {
	images, err := dockerstatus.DockerImageRepositories()
	if err != nil {
		return 1, "", err
	}
	if tabular.StrIn(chk.name, images) {
		return errutil.Success()
	}
	return errutil.GenericError("Docker image was not found", chk.name, images)
}

/*
#### DockerImageRegexp
Description: Works like DockerImage, but matches via a regexp, rather than a
string.
*/

type DockerImageRegexp struct{ re *regexp.Regexp }

func (chk DockerImageRegexp) ID() string { return "DockerImageRegexp" }

func (chk DockerImageRegexp) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	re, err := regexp.Compile(params[0])
	if err != nil {
		return chk, errutil.ParameterTypeError{params[0], "regexp"}
	}
	chk.re = re
	return chk, nil
}

func (chk DockerImageRegexp) Status() (int, string, error) {
	images, err := dockerstatus.DockerImageRepositories()
	if err != nil {
		return 1, "", err
	}
	if tabular.ReIn(chk.re, images) {
		return errutil.Success()
	}
	msg := "Docker image was not found."
	return errutil.GenericError(msg, chk.re.String(), images)
}

/*
#### DockerRunning
Description: Is this Docker container running?
Parameters:
  - Name (string):  Name of the container
Example parameters:
  - "user/container", "user/container:latest"
*/

type DockerRunning struct{ name string }

func (chk DockerRunning) ID() string { return "DockerRunning" }

func (chk DockerRunning) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	chk.name = params[0]
	return chk, nil
}

func (chk DockerRunning) Status() (int, string, error) {
	running, err := dockerstatus.RunningContainers()
	if err != nil {
		return 1, "", err
	}
	if tabular.StrContainedIn(chk.name, running) {
		return errutil.Success()
	}
	msg := "Docker container not runnning"
	return errutil.GenericError(msg, chk.name, running)
}

// getRunningContainersAPI is like getRunningContainers, but uses an external
// library in order to access the Docker API
func getRunningContainersAPI(endpoint string) (containers []string) {
	client, err := docker.NewClient(endpoint)
	if err != nil {
		log.WithFields(log.Fields{
			"endpoint": endpoint,
			"error":    err.Error(),
		}).Fatal("Couldn't create Docker API client")
	}
	ctrs, err := client.ListContainers(docker.ListContainersOptions{All: false})
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatal("Couldn't list Docker containers")
	}
	for _, ctr := range ctrs {
		if strings.Contains(ctr.Status, "Up") {
			containers = append(containers, ctr.Image)
		}
	}
	return containers
}

/*
#### DockerRunningAPI
Description: Works like DockerRunning, but fetches information from the Docker
API endpoint instead.
Parameters:
  - Path (filepath): Path to Docker socket
  - Name (string): Name of the container
Example parameters:
  - "/var/run/docker.sock", "/path/to/docker.sock"
  - "user/container", "user/container:latest"
*/

type DockerRunningAPI struct{ path, name string }

func (chk DockerRunningAPI) New(params []string) (chkutil.Check, error) {
	if len(params) != 2 {
		return chk, errutil.ParameterLengthError{2, params}
	}
	path := params[0]
	if _, err := os.Stat(path); err != nil {
		return chk, errutil.ParameterTypeError{path, "filepath"}
	}
	chk.path = path
	chk.name = params[1]
	return chk, nil
}

func (chk DockerRunningAPI) Status() (int, string, error) {
	running := getRunningContainersAPI(chk.path)
	if tabular.StrContainedIn(chk.name, running) {
		return errutil.Success()
	}
	msg := "Docker container not runnning"
	return errutil.GenericError(msg, chk.name, running)
}

/*
#### DockerRunningRegexp
Description: Works like DockerRunning, but matches with a regexp instead of a
string.
Parameters:
  - Regexp (regexp): Regexp to match names with
Example parameters:
  - "user/.+", "user/[cC](o){2,3}[nta]tai\w{2}r"
*/
type DockerRunningRegexp struct{ re *regexp.Regexp }

func (chk DockerRunningRegexp) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	re, err := regexp.Compile(params[0])
	if err != nil {
		return chk, errutil.ParameterTypeError{params[0], "regexp"}
	}
	chk.re = re
	return chk, nil
}

func (chk DockerRunningRegexp) Status() (int, string, error) {
	running, err := dockerstatus.RunningContainers()
	if err != nil {
		return 1, "", err
	}
	if tabular.ReIn(chk.re, running) {
		return errutil.Success()
	}
	msg := "Docker container not runnning"
	return errutil.GenericError(msg, chk.re.String(), running)
}
