package usrstatus

import (
	"testing"
)

func TestGroups(t *testing.T) {
	t.Parallel()
	groups, err := Groups()
	if err != nil {
		t.Error("Groups failed")
	}
	if len(groups) < 1 {
		t.Error("Couldn't find any groups in /etc/group")
	}
}
