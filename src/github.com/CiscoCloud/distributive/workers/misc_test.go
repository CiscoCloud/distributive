package workers

import "testing"

func TestCommand(t *testing.T) {
	t.Parallel()
	validInputs := [][]string{
		[]string{"sleep 0.00000001"},
		[]string{"echo this works"},
		[]string{"cd"},
		[]string{"mv --help"},
	}
	invalidInputs := notLengthOne
	goodEggs := validInputs
	badEggs := [][]string{
		[]string{"sleep fail"},
		[]string{"cd /steppenwolf"},
		[]string{"mv /glass /bead-game"},
	}
	badEggs = append(badEggs, names...)
	testParameters(validInputs, invalidInputs, Command{}, t)
	testCheck(goodEggs, badEggs, Command{}, t)
}

func TestCommandOutputMatches(t *testing.T) {
	t.Parallel()
	validInputs := [][]string{
		[]string{"echo siddhartha", "sid"},
		[]string{"cp --help", "cp"},
		[]string{"echo euler", "eu"},
	}
	invalidInputs := notLengthTwo
	goodEggs := validInputs
	badEggs := [][]string{
		[]string{"echo siddhartha", "fail"},
		[]string{"cp --help", "asdfalkjsdhldjfk"},
		[]string{"echo haskell", "curry"},
	}
	testParameters(validInputs, invalidInputs, CommandOutputMatches{}, t)
	testCheck(goodEggs, badEggs, CommandOutputMatches{}, t)
}

func TestRunning(t *testing.T) {
	t.Parallel()
	validInputs := [][]string{
		[]string{"proc"}, []string{"nginx"}, []string{"anything"},
		[]string{"worker"}, []string{"distributive"},
	}
	invalidInputs := notLengthOne
	invalidInputs = append(invalidInputs, names...)
	goodEggs := [][]string{}
	badEggs := dirParameters
	testParameters(validInputs, invalidInputs, Running{}, t)
	testCheck(goodEggs, badEggs, Running{}, t)
}

func TestTemp(t *testing.T) {
	t.Parallel()
	validInputs := ints
	invalidInputs := append(append(names, notInts...), notLengthOne...)
	goodEggs := [][]string{
		[]string{"1414"}, // melting temp. of silicon
		[]string{"1510"}, // " " " steel
		[]string{"1600"}, // " " " glass
	}
	badEggs := [][]string{
		[]string{"0"}, // freezing temp. of water
		[]string{"1"},
		[]string{"2"},
	}
	testParameters(validInputs, invalidInputs, Temp{}, t)
	testCheck(goodEggs, badEggs, Temp{}, t)
}

func TestModule(t *testing.T) {
	t.Parallel()
	validInputs := names
	invalidInputs := notLengthOne
	goodEggs := [][]string{}
	badEggs := names
	testParameters(validInputs, invalidInputs, Module{}, t)
	testCheck(goodEggs, badEggs, Module{}, t)
}

func TestKernelParameter(t *testing.T) {
	validInputs := names
	invalidInputs := notLengthOne
	goodEggs := [][]string{
		[]string{"net.ipv4.conf.all.accept_local"},
		[]string{"net.ipv4.conf.all.accept_redirects"},
		[]string{"net.ipv4.conf.all.arp_accept"},
	}
	badEggs := names
	testParameters(validInputs, invalidInputs, KernelParameter{}, t)
	testCheck(goodEggs, badEggs, KernelParameter{}, t)
}

func TestPHPConfig(t *testing.T) {
	t.Parallel()
	validInputs := appendParameter(names, "dummy-value")
	invalidInputs := notLengthTwo
	goodEggs := [][]string{}
	badEggs := validInputs
	testParameters(validInputs, invalidInputs, PHPConfig{}, t)
	testCheck(goodEggs, badEggs, PHPConfig{}, t)
}
