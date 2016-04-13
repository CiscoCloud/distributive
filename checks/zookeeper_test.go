package checks

import (
	"testing"
)

var validInputs = [][]string{
	[]string{"2ms", "wikipedia.org:9814"},
	[]string{"1ms", "mozilla.org:9814"},
}

var invalidInputs = [][]string{
	{"", "mozilla.net"},
	{"nottime", "wikipedia.org"},
}

func TestZooKeeperRUOK(t *testing.T) {
	t.Parallel()
	testParameters(validInputs, invalidInputs, ZooKeeperRUOK{}, t)
	testCheck([][]string{}, validInputs, ZooKeeperRUOK{}, t)
}

func TestZooKeeperLatency(t *testing.T) {
	t.Parallel()
	testParameters(validInputs, invalidInputs, ZooKeeperLatency{}, t)
	testCheck([][]string{}, validInputs, ZooKeeperLatency{}, t)
}
