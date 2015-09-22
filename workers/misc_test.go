package workers

import "testing"

func TestCommand(t *testing.T) {
	//t.Parallel()
	validInputs := [][]string{
		{"sleep 0.00000001"}, {"echo this works"}, {"cd"}, {"mv --help"},
	}
	invalidInputs := notLengthOne
	goodEggs := validInputs
	badEggs := [][]string{
		{"sleep fail"}, {"cd /steppenwolf"}, {"mv /glass /bead-game"},
	}
	badEggs = append(badEggs, names...)
	testParameters(validInputs, invalidInputs, Command{}, t)
	testCheck(goodEggs, badEggs, Command{}, t)
}

func TestCommandOutputMatches(t *testing.T) {
	//t.Parallel()
	validInputs := [][]string{
		{"echo siddhartha", "sid"}, {"cp --help", "cp"}, {"echo euler", "eu"},
	}
	invalidInputs := notLengthTwo
	goodEggs := validInputs
	badEggs := [][]string{
		{"echo siddhartha", "fail"},
		{"cp --help", "asdfalkjsdhldjfk"},
		{"echo haskell", "curry"},
	}
	testParameters(validInputs, invalidInputs, CommandOutputMatches{}, t)
	testCheck(goodEggs, badEggs, CommandOutputMatches{}, t)
}

func TestRunning(t *testing.T) {
	//t.Parallel()
	validInputs := append(names, [][]string{
		{"proc"}, {"nginx"}, {"anything"}, {"worker"}, {"distributive"},
	}...)
	invalidInputs := notLengthOne
	goodEggs := [][]string{}
	badEggs := dirParameters
	testParameters(validInputs, invalidInputs, Running{}, t)
	testCheck(goodEggs, badEggs, Running{}, t)
}

func TestTemp(t *testing.T) {
	//t.Parallel()
	validInputs := positiveInts[:len(positiveInts)-2] // only small ints
	invalidInputs := append(append(names, notInts...), notLengthOne...)
	goodEggs := [][]string{
		{"1414"}, // melting temp. of silicon
		{"1510"}, // " " " steel
		{"1600"}, // " " " glass
	}
	badEggs := [][]string{{"0"}, {"1"}, {"2"}}
	testParameters(validInputs, invalidInputs, Temp{}, t)
	testCheck(goodEggs, badEggs, Temp{}, t)
}

func TestModule(t *testing.T) {
	//t.Parallel()
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
		{"net.ipv4.conf.all.accept_local"},
		{"net.ipv4.conf.all.accept_redirects"},
		{"net.ipv4.conf.all.arp_accept"},
	}
	badEggs := names
	testParameters(validInputs, invalidInputs, KernelParameter{}, t)
	testCheck(goodEggs, badEggs, KernelParameter{}, t)
}

func TestPHPConfig(t *testing.T) {
	//t.Parallel()
	validInputs := appendParameter(names, "dummy-value")
	invalidInputs := notLengthTwo
	goodEggs := [][]string{}
	badEggs := validInputs
	testParameters(validInputs, invalidInputs, PHPConfig{}, t)
	testCheck(goodEggs, badEggs, PHPConfig{}, t)
}
