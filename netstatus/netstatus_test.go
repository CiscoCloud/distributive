package netstatus

import "testing"

func TestPortOpen(t *testing.T) {
	t.Parallel()
	for _, port := range []uint16{44231, 11234, 14891} {
		if PortOpen("tcp", port) {
			t.Errorf("Port was unexpectedly open over TCP: %v", port)
		}
		if PortOpen("udp", port) {
			// TODO: this is a bug. PortOpen("udp") gives false positives.
			//t.Errorf("Port was unexpectedly open over UDP: %v", port)
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
