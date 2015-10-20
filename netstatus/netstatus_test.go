package netstatus

import (
	"fmt"
	"testing"
	"time"
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
	if len(GetHexPorts("tcp")) < 1 {
		t.Error("len(GetHexPorts(tcp)) < 1")
	} else if len(GetHexPorts("udp")) < 1 {
		t.Error("len(GetHexPorts(udp)) < 1")
	}
}

func TestOpenPorts(t *testing.T) {
	// TODO test valid range
	t.Parallel()
	if len(OpenPorts("tcp")) < 1 {
		t.Error("len(OpenPorts(tcp)) < 1")
	} else if len(OpenPorts("udp")) < 1 {
		t.Error("len(OpenPorts(udp)) < 1")
	}
}

func TestPortOpen(t *testing.T) {
	t.Parallel()
	for _, protocol := range [2]string{"tcp", "udp"} {
		for _, port := range OpenPorts(protocol) {
			if !PortOpen(protocol, port) {
				msg := "PortOpen and OpenPorts reported differently for "
				t.Errorf(msg + fmt.Sprint(port) + " with protocol " + protocol)
			}
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
			msg := ". Are all of your interfaces connected?"
			t.Error("len(InterfaceIPs) < 1" + msg)
		}
	}
}

func TestResolvable(t *testing.T) {
	t.Parallel()
	// TODO
}

func TestCanConnect(t *testing.T) {
	t.Parallel()
	goodHosts := []string{"eff.org:80", "google.com:80", "bing.com:80"}
	for _, host := range goodHosts {
		duration, err := time.ParseDuration("20s")
		if err != nil {
			t.Error(err.Error())
		}
		if !CanConnect(host, "TCP", duration) {
			t.Error("Couldn't connect to host " + host)
		}
	}
	badHosts := []string{"asdklfhabssdla.com:80", "lkjashldfb.com:80"}
	for _, host := range badHosts {
		duration, err := time.ParseDuration("20s")
		if err != nil {
			t.Error(err.Error())
		}
		if CanConnect(host, "TCP", duration) {
			t.Error("Could connect to host " + host)
		}
	}
}
