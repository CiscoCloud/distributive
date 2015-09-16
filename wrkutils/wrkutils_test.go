package wrkutils

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

func TestCommandOutput(t *testing.T) {
	t.Parallel()
	cmds := []*exec.Cmd{
		exec.Command("echo", "test"),
		exec.Command("ls", "-a"),
		exec.Command("cat", "/dev/null"),
	}
	outputs := []string{"test", ".", ""}
	for i := range cmds {
		cmd := cmds[i]
		expected := outputs[i]
		actual := CommandOutput(cmd)
		if !strings.Contains(actual, expected) {
			msg := "Command output did not contain expected string"
			msg += "\n\tCommand: " + fmt.Sprint(cmd.Args)
			msg += "\n\tExpected output: " + expected
			msg += "\n\tActual output: " + actual
		}
	}
}

func TestGetByteUnits(t *testing.T) {
	t.Parallel()
	inputs := []string{
		"4400gb", "10b", "100mb", "nonsense", "3017TiB", "1234TBTB", "12Kb",
	}
	outputs := []string{"gb", "b", "mb", "", "tb", "", "kb"}
	for i := range inputs {
		input := inputs[i]
		expected := outputs[i]
		actual := GetByteUnits(input)
		if actual != expected {
			msg := "GetByteUnits didn't properly extract the units"
			msg += "\n\tString: " + input
			msg += "\n\tExpected output: " + expected
			msg += "\n\tActual output: " + actual
		}
	}
}

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
		cd, ms := GenericError(msg, spc, act)
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

func TestParseUserRegex(t *testing.T) {
	t.Parallel()
	regexps := []string{
		`aslfd`, `\w`, `\s\w\d`,
		`regexp.MustCompile('regexp.MustCompile(...)')`,
		`string(string(string(string(string(string(string))))))`,
	}
	for _, restr := range regexps {
		ParseUserRegex(restr)
	}
}

var paths = []string{
	"/proc/net/tcp",
	"/bin/bash",
	"/proc/filesystems",
	"/proc/uptime",
	"/proc/cpuinfo",
}

func TestFileToBytes(t *testing.T) {
	t.Parallel()
	for _, path := range paths {
		result := FileToBytes(path)
		if result == nil {
			msg := "FileToBytes returned a nil result"
			msg += "Path: " + fmt.Sprint(path)
			msg += "Result: " + fmt.Sprint(result)
			t.Error(msg)
		} else if len(result) < 1 {
			msg := "FileToBytes returned a zero length result"
			msg += "Path: " + fmt.Sprint(path)
			msg += "Result: " + fmt.Sprint(result)
			t.Error(msg)
		}
	}
}

func TestFileToString(t *testing.T) {
	t.Parallel()
	for _, path := range paths {
		result := FileToString(path)
		if result == "" {
			msg := "FileToString returned an empty result"
			msg += "Path: " + fmt.Sprint(path)
			msg += "Result: " + fmt.Sprint(result)
			t.Error(msg)
		}
	}
}

func TestFileToLines(t *testing.T) {
	t.Parallel()
	for _, path := range paths {
		result := FileToLines(path)
		if result == nil {
			msg := "FileToLines returned a nil result"
			msg += "Path: " + fmt.Sprint(path)
			msg += "Result: " + fmt.Sprint(result)
			t.Error(msg)
		} else if len(result) < 1 {
			msg := "FileToLines returned a zero length result"
			msg += "Path: " + fmt.Sprint(path)
			msg += "Result: " + fmt.Sprint(result)
			t.Error(msg)
		}
	}
}

func TestParseMyInt(t *testing.T) {
	t.Parallel()
	inputs := []string{
		"9223372036854775807", "-9223372036854775808",
		"0", "-1", "100", "102348", "-3104981209384",
	}
	outputs := []int{
		9223372036854775807, -9223372036854775808,
		0, -1, 100, 102348, -3104981209384,
	}
	for i := range inputs {
		input := inputs[i]
		expected := outputs[i]
		actual := ParseMyInt(input)
		if actual != expected {
			msg := "ParseMyInt didn't output the expected result"
			msg += "\n\tExpected: " + fmt.Sprint(expected)
			msg += "\n\tActual: " + fmt.Sprint(actual)
			t.Error(msg)
		}
	}
}
