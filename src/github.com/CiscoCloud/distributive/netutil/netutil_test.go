package netutil

import (
	"testing"
)

// http://sqa.fyicenter.com/Online_Test_Tools/IP_Address_Format_Validator.php
// TODO make more
var validIPs = []string{
	"192.168.0.1",
}

var invalidIPs = []string{
	"256.256.256.256",
}

func TestGetHexPorts(t *testing.T) {
	t.Parallel()
	if len(GetHexPorts()) < 1 {
		t.Error("len(GetHexPorts) < 1")
	}
}

func TestOpenPorts(t *testing.T) {
	// TODO test valid range
	t.Parallel()
	if len(OpenPorts()) < 1 {
		t.Error("len(OpenPorts) < 1")
	}
}

func TestPortOpen(t *testing.T) {
	t.Parallel()
	for _, port := range OpenPorts() {
		if !PortOpen(port) {
			t.Errorf("PortOpen and OpenPorts reported differently for %s", port)
		}
	}
	// TODO test all other ports in valid range
}

func TestValidIP(t *testing.T) {
	t.Parallel()
	for _, valid := range validIPs {
		if !ValidIP(valid) {
			t.Errorf("ValidIP reported incorrectly for IP %s", valid)
		}
	}
	for _, invalid := range invalidIPs {
		if ValidIP(invalid) {
			t.Errorf("ValidIP reported incorrectly for IP %s", invalid)
		}
	}
}

func TestGetInterfaces(t *testing.T) {
	t.Parallel()
	if len(GetInterfaces()) < 1 {
		t.Error("len(GetInterfaces) < 1")
	}
}

func TestInterfaceIPs(t *testing.T) {
	t.Parallel()
	// TODO check for valid interface IP addresses with ValidIP
	for _, iface := range GetInterfaces() {
		if len(InterfaceIPs(iface.Name)) < 1 {
			t.Error("len(InterfaceIPs) < 1")
		}
	}
}

func TestResolvable(t *testing.T) {
	t.Parallel()
	// TODO
}

func TestCanConnect(t *testing.T) {
	t.Parallel()
	// TODO
}
