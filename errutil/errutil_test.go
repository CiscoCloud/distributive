package errutil

import (
	"fmt"
	"strings"
	"testing"
)

func TestGenericError(t *testing.T) {
	t.Parallel()
	msgs := []string{"msg1", "msg2", "msg3", "msg4", "msg5"}
	specs := []string{"spc1", "spc2", "spc3", "spc4", "spc5"}
	acts := [][]string{
		[]string{"act1"}, []string{"act2"}, []string{"act3"},
		[]string{"act4"}, []string{"act5"},
	}
	for i := range msgs {
		msg := msgs[i]
		spc := specs[i]
		act := acts[i]
		// TODO include error
		cd, ms, _ := GenericError(msg, spc, act)
		if !strings.Contains(ms, msg) {
			err := "GenericError's message didn't contain the message"
			err += "\n\tExpected: " + msg
			err += "\n\tActual: " + ms
			t.Error(err)
		}
		if !strings.Contains(ms, spc) {
			err := "GenericError's message didn't contain the specified"
			err += "\n\tExpected: " + spc
			err += "\n\tActual: " + ms
			t.Error(err)
		}
		if !strings.Contains(ms, act[0]) {
			err := "GenericError's message didn't contain the actual"
			err += "\n\tExpected: " + act[0]
			err += "\n\tActual: " + ms
			t.Error(err)
		}
		if cd <= 0 {
			err := "GenericError had a <= 0 exit code"
			err += "\n\tActual: " + fmt.Sprint(cd)
			t.Error(err)
		}
	}
}
