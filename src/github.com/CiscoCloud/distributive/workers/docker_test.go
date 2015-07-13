package workers

import (
	"strings"
	"testing"
)

func TestDockerImageFail(t *testing.T) {
	parameters := []string{"failme"}
	msgTest := func(exitMessage string) (passing bool, msg string) {
		if strings.Contains(exitMessage, "not found") {
			return true, ""
		}
		return false, "Expected exit message to contain 'not found'"
	}
	generalTestFunction(dockerImage, parameters, gtZero, msgTest, t)
}
