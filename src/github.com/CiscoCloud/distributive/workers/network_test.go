package workers

import (
	"testing"
)

var validHosts = [][]string{
	[]string{"eff.org"},
	[]string{"mozilla.org"},
	[]string{"golang.org"},
}

var invalidHosts = [][]string{
	[]string{"asldkjahserbapsidpuflnaskjdcasd.com"},
	[]string{"aspoiqpweroiqewruqpwioepbpasdfb.net"},
	[]string{"lkjqhwelrjblrjbbrbbbnnzasdflbaj.org"},
}

var validURLs = prefixParameter(validHosts, "http://")
var invalidURLs = prefixParameter(invalidHosts, "http://")
var validHostsWithPort = suffixParameter(validHosts, ":80")
var invalidHostsWithPort = suffixParameter(invalidHosts, ":80")

func TestPort(t *testing.T) {
	t.Parallel()
	validInputs := ints
	invalidInputs := notInts
	goodEggs := [][]string{}
	badEggs := [][]string{
		[]string{"49151"}, // reserved
		[]string{"5310"},  // Outlaws (1997 video game)
		[]string{"0"},     // reserved
		[]string{"2302"},  // Halo: Combat Evolved multiplayer
	}
	testParameters(validInputs, invalidInputs, Port{}, t)
	testCheck(goodEggs, badEggs, Port{}, t)
}

func TestInterfaceExists(t *testing.T) {
	t.Parallel()
	validInputs := names
	invalidInputs := notLengthOne
	goodEggs := [][]string{}
	badEggs := [][]string{}
	testParameters(validInputs, invalidInputs, InterfaceExists{}, t)
	testCheck(goodEggs, badEggs, InterfaceExists{}, t)
}

func TestUp(t *testing.T) {
	t.Parallel()
	validInputs := names
	invalidInputs := notLengthOne
	goodEggs := [][]string{}
	badEggs := [][]string{}
	testParameters(validInputs, invalidInputs, Up{}, t)
	testCheck(goodEggs, badEggs, Up{}, t)
}

func TestIP4(t *testing.T) {
	t.Parallel()
	validInputs := appendParameter(names, "0000:000:0000:000:0000:0000:000:0000")
	invalidInputs := notLengthTwo
	goodEggs := [][]string{}
	badEggs := validInputs
	testParameters(validInputs, invalidInputs, IP4{}, t)
	testCheck(goodEggs, badEggs, IP4{}, t)
}

func TestIP6(t *testing.T) {
	t.Parallel()
	validInputs := appendParameter(names, "0000:000:0000:000:0000:0000:000:0000")
	invalidInputs := notLengthTwo
	goodEggs := [][]string{}
	badEggs := validInputs
	testParameters(validInputs, invalidInputs, IP6{}, t)
	testCheck(goodEggs, badEggs, IP6{}, t)
}

func TestGatewayInterface(t *testing.T) {
	t.Parallel()
	testParameters(names, notLengthOne, IP6{}, t)
	testCheck([][]string{}, names, IP6{}, t)
}

func TestHost(t *testing.T) {
	t.Parallel()
	validInputs := names
	invalidInputs := notLengthOne
	goodEggs := validHosts
	badEggs := invalidHosts
	testParameters(validInputs, invalidInputs, Host{}, t)
	testCheck(goodEggs, badEggs, Host{}, t)
}

func TestTCP(t *testing.T) {
	t.Parallel()
	testParameters(names, notLengthOne, TCP{}, t)
	testCheck(validHostsWithPort, invalidHostsWithPort, TCP{}, t)
}

func TestUDP(t *testing.T) {
	t.Parallel()
	testParameters(names, notLengthOne, UDP{}, t)
	testCheck(validHostsWithPort, invalidHostsWithPort, UDP{}, t)
}

func TestTCPTimeout(t *testing.T) {
	t.Parallel()
	goodEggs := appendParameter(validHostsWithPort, "5s")
	badEggs := appendParameter(validHostsWithPort, "1µs")
	testParameters(names, notLengthOne, TCPTimeout{}, t)
	testCheck(goodEggs, badEggs, TCPTimeout{}, t)
}

func TestUDPTimeout(t *testing.T) {
	t.Parallel()
	goodEggs := appendParameter(validHostsWithPort, "5s")
	badEggs := appendParameter(validHostsWithPort, "1µs")
	testParameters(names, notLengthOne, UDPTimeout{}, t)
	testCheck(goodEggs, badEggs, UDPTimeout{}, t)
}

func TestRoutingTableDestination(t *testing.T) {
	t.Parallel()
	testParameters(names, notLengthOne, RoutingTableDestination{}, t)
	testCheck([][]string{}, names, RoutingTableDestination{}, t)
}

func TestRoutingTableInterface(t *testing.T) {
	t.Parallel()
	testParameters(names, notLengthOne, RoutingTableInterface{}, t)
	testCheck([][]string{}, names, RoutingTableInterface{}, t)
}

func TestRoutingTableGateway(t *testing.T) {
	t.Parallel()
	testParameters(names, notLengthOne, RoutingTableGateway{}, t)
	testCheck([][]string{}, names, RoutingTableGateway{}, t)
}

func TestReponseMatches(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("Skipping tests that query remote servers in short mode")
	} else {
		validInputs := appendParameter(names, "match")
		invalidInputs := notLengthTwo
		goodEggs := appendParameter(validURLs, "html")
		badEggs := appendParameter(validURLs, "asfdjhow012u")
		testParameters(validInputs, invalidInputs, ResponseMatches{}, t)
		testCheck(goodEggs, badEggs, ResponseMatches{}, t)
	}
}

func TestReponseMatchesInsecure(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("Skipping tests that query remote servers in short mode")
	} else {
		validInputs := appendParameter(names, "match")
		invalidInputs := notLengthTwo
		goodEggs := appendParameter(validURLs, "html")
		badEggs := appendParameter(validURLs, "asfdjhow012u")
		testParameters(validInputs, invalidInputs, ResponseMatchesInsecure{}, t)
		testCheck(goodEggs, badEggs, ResponseMatchesInsecure{}, t)
	}
}
