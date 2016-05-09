package chkutil

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

func TestSeparateByteUnits(t *testing.T) {
	t.Parallel()
	inputs := []string{
		"4400gb", "10b", "100mb", "3017TiB", "1234TBTB", "12Kb",
	}
	outputInts := []int{4400, 10, 100, 3017, 1234, 12}
	outputUnits := []string{"gb", "b", "mb", "tb", "", "kb"}
	for i := range inputs {
		input := inputs[i]
		expectedScalar := outputInts[i]
		expectedUnit := outputUnits[i]
		amount, unit, err := SeparateByteUnits(input)
		if err != nil {
			msg := "SeparateByteUnits reported unexpected error"
			msg += "\n\tString: " + input
			msg += "\n\tExpected scalar: " + fmt.Sprint(expectedScalar)
			msg += "\n\tExpected units: " + expectedUnit
			msg += "\n\tError: " + err.Error()

		} else if unit != expectedUnit || amount != expectedScalar {
			msg := "SeparateByteUnits didn't properly extract the units/amount"
			msg += "\n\tString: " + input
			msg += "\n\tExpected scalar: " + fmt.Sprint(expectedScalar)
			msg += "\n\tExpected units: " + expectedUnit
			msg += "\n\tActual scalar: " + fmt.Sprint(amount)
			msg += "\n\tActual units: " + unit
		}
	}
}

// TODO
func TestSubmatchMap(t *testing.T) {
	t.Parallel()
}

var paths = []string{
	"/proc/net/tcp",
	"/bin/sh",
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

func TestBytesToFile(t *testing.T) {
	t.Parallel()
	// TODO
}

func TestURLToBytes(t *testing.T) {
	t.Parallel()
	// TODO
}

func TestGetFileWithExtension(t *testing.T) {
	t.Parallel()
	// TODO
}
