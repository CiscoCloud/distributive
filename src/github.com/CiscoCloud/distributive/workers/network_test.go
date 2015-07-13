package workers

import (
	"testing"
)

var validHosts = []parameters{
	[]string{"eff.org"},
	[]string{"mozilla.org"},
	[]string{"golang.org"},
}

var invalidHosts = []parameters{
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
	losers := []parameters{
		[]string{"49151"}, // reserved
		[]string{"5310"},  // Outlaws (1997 video game)
		[]string{"0"},     // reserved
		[]string{"2302"},  // Halo: Combat Evolved multiplayer
	}
	testInputs(t, port, []parameters{}, losers)
}

func TestInterfaceExists(t *testing.T) {
	t.Parallel()
	testInputs(t, interfaceExists, []parameters{[]string{"lo"}}, names)
}

func TestUp(t *testing.T) {
	t.Parallel()
	testInputs(t, up, []parameters{[]string{"lo"}}, names)
}

func TestIP4(t *testing.T) {
	t.Parallel()
	losers := appendParameter(names, "0.0.0.0")
	testInputs(t, ip4, []parameters{}, losers)
}

func TestIP6(t *testing.T) {
	t.Parallel()
	losers := appendParameter(names, "0000:000:0000:000:0000:0000:000:0000")
	testInputs(t, ip6, []parameters{}, losers)
}

func TestGatewayInterface(t *testing.T) {
	t.Parallel()
	testInputs(t, gatewayInterface, []parameters{}, names)
}

func TestHost(t *testing.T) {
	t.Parallel()
	testInputs(t, host, validHosts, invalidHosts)
}

func TestTCP(t *testing.T) {
	t.Parallel()
	testInputs(t, tcp, validHostsWithPort, invalidHostsWithPort)
}

func TestUDP(t *testing.T) {
	t.Parallel()
	testInputs(t, udp, validHostsWithPort, invalidHostsWithPort)
}

func TestTCPTimeout(t *testing.T) {
	t.Parallel()
	winners := appendParameter(validHostsWithPort, "5s")
	losers := appendParameter(validHostsWithPort, "1µs")
	testInputs(t, tcpTimeout, winners, losers)
}

func TestUDPTimeout(t *testing.T) {
	t.Parallel()
	winners := appendParameter(validHostsWithPort, "5s")
	losers := appendParameter(validHostsWithPort, "1µs")
	testInputs(t, udpTimeout, winners, losers)
}

func TestRoutingTableDestination(t *testing.T) {
	t.Parallel()
	losers := names
	testInputs(t, routingTableDestination, []parameters{}, losers)
}

func TestRoutingTableInterface(t *testing.T) {
	t.Parallel()
	losers := names
	testInputs(t, routingTableInterface, []parameters{}, losers)
}

func TestRoutingTableGateway(t *testing.T) {
	t.Parallel()
	losers := names
	testInputs(t, routingTableGateway, []parameters{}, losers)
}

func TestReponseMatches(t *testing.T) {
	t.Parallel()
	winners := appendParameter(validURLs, "html")
	losers := appendParameter(validURLs, "asfdjhow012u")
	testInputs(t, responseMatches, winners, losers)
}

func TestReponseMatchesInsecure(t *testing.T) {
	t.Parallel()
	winners := appendParameter(validURLs, "html")
	losers := appendParameter(validURLs, "asfdjhow012u")
	testInputs(t, responseMatches, winners, losers)
}
