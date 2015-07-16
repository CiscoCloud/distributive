package workers

import "testing"

func TestCommand(t *testing.T) {
	t.Parallel()
	winners := []parameters{
		[]string{"sleep 0.00000001"},
		[]string{"echo this works"},
		[]string{"cd"},
		[]string{"mv --help"},
	}
	losers := []parameters{
		[]string{"sleep fail"},
		[]string{"cd /steppenwolf"},
		[]string{"mv /glass /bead-game"},
	}
	testInputs(t, command, winners, losers)
}

func TestCommandOutputMatches(t *testing.T) {
	t.Parallel()
	winners := []parameters{
		[]string{"echo siddhartha", "sid"},
		[]string{"cp --help", "cp"},
		[]string{"echo euler", "eu"},
	}
	losers := []parameters{
		[]string{"echo siddhartha", "fail"},
		[]string{"cp --help", "asdfalkjsdhldjfk"},
		[]string{"echo haskell", "curry"},
	}
	testInputs(t, commandOutputMatches, winners, losers)
}

func TestRunning(t *testing.T) {
	t.Parallel()
	winners := []parameters{
		[]string{"/sbin/init"},
		[]string{"[perf]"},
		[]string{"[kthreadd]"},
	}
	losers := []parameters{
		[]string{"bert"},
		[]string{"ernie"},
		[]string{"statler"},
	}
	testInputs(t, running, winners, losers)
}

func TestTemp(t *testing.T) {
	t.Parallel()
	winners := []parameters{
		[]string{"1414"}, // melting temp. of silicon
		[]string{"1510"}, // " " " steel
		[]string{"1600"}, // " " " glass
	}
	losers := []parameters{
		[]string{"0"}, // freezing temp. of water
		[]string{"1"},
		[]string{"2"},
	}
	testInputs(t, temp, winners, losers)
}

func TestModule(t *testing.T) {
	t.Parallel()
	winners := []parameters{
		[]string{"iptable_nat"},
		[]string{"snd"},
		[]string{"loop"},
	}
	losers := []parameters{
		[]string{"knecht"},
		[]string{"designori"},
		[]string{"tegularius"},
	}
	testInputs(t, module, winners, losers)
}

func TestKernelParameter(t *testing.T) {
	winners := []parameters{
		[]string{"net.ipv4.conf.all.accept_local"},
		[]string{"net.ipv4.conf.all.accept_redirects"},
		[]string{"net.ipv4.conf.all.arp_accept"},
	}
	losers := []parameters{
		[]string{"harry haller"},
		[]string{"loering"},
		[]string{"hermine"},
	}
	testInputs(t, kernelParameter, winners, losers)
}

func TestPHPConfig(t *testing.T) {
	t.Parallel()
	losers := []parameters{
		[]string{"vasudeva", "value"},
		[]string{"kamala", "value"},
		[]string{"gotama", "value"},
	}
	testInputs(t, phpConfig, []parameters{}, losers)
}
