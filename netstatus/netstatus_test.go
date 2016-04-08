package netstatus

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestGetHexPorts(t *testing.T) {
	t.Parallel()
	if len(GetHexPorts("tcp")) < 1 {
		t.Errorf("len(GetHexPorts('tcp')) = %s", len(GetHexPorts("tcp")))
	} else if len(GetHexPorts("udp")) < 1 {
		t.Logf("len(GetHexPorts('udp') = %s", len(GetHexPorts("udp")))
	}
}

func TestOpenPorts(t *testing.T) {
	t.Parallel()
	for _, protocol := range [2]string{"tcp", "udp"} {
		ports := OpenPorts(protocol)
		if len(ports) < 1 {
			t.Logf("OpenPorts reported zero open ports for %s", protocol)
		}
		// test how many we can actually connect to as well (just FYI)
		couldConnect := 0
		for _, port := range ports {
			if 0 > port || 65535 < port {
				msg := "OpenPorts reported invalid port number on " + protocol
				t.Errorf(msg+": %d", port)
			} else {
				dur, _ := time.ParseDuration("10s")
				addr := fmt.Sprintf("localhost:%v", port)
				if _, err := net.DialTimeout(protocol, addr, dur); err != nil {
					couldConnect++
				}
			}
		}
		t.Logf("Could connect to %v/%v ports", couldConnect, len(ports))
	}
}

func TestPortOpen(t *testing.T) {
	t.Parallel()
	for _, protocol := range [2]string{"tcp", "udp"} {
		for _, port := range OpenPorts(protocol) {
			if !PortOpen(protocol, port) {
				msg := "PortOpen and OpenPorts reported differently for "
				t.Logf("%s %v with protocol %s", msg, port, protocol)
			}
		}
	}
}

func TestValidIP(t *testing.T) {
	t.Parallel()
	validIPs := []string{
		"192.168.0.1",
		"10.0.5.23",
		"192.168.1.123",
		"172.16.0.41",
		"fe80::5e51:4fff:fe98:4da1",
		"2001:cdba:0000:0000:0000:0000:3257:9652",
	}
	invalidIPs := []string{
		"256.256.256.256",
		":::::::::1",
	}
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
	for _, iface := range GetInterfaces() {
		ips := InterfaceIPs(iface.Name)
		if len(ips) < 1 {
			msg := ". Are all of your interfaces connected?"
			t.Error("len(InterfaceIPs) < 1" + msg)
		} else {
			for _, ip := range ips {
				if !ValidIP(ip.String()) {
					msg := "IntefaceIPs reported an invalid address: %s"
					t.Errorf(msg, ip.String())
				}
			}
		}
	}
}

func TestResolvable(t *testing.T) {
	t.Parallel()
	goodHosts := [1]string{"localhost"}
	for _, host := range goodHosts {
		if !Resolvable(host) {
			t.Error("Resolvable reported incorrectly for %s", host)
		}
	}
	badHosts := [3]string{"aspdofhas", "piqwehrpb", "qoiwufbsal"}
	for _, host := range badHosts {
		if Resolvable(host) {
			t.Error("Resolvable reported incorrectly for %s", host)
		}
	}
}
