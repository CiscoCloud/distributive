package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os/exec"
	"regexp"
	"strconv"
)

// getHexPorts gets all open ports as hex strings from /proc/net/tcp
func getHexPorts() (ports []string) {
	out, err := ioutil.ReadFile("/proc/net/tcp")
	fatal(err)
	localAddresses := getColumnNoHeader(1, stringToSlice(string(out)))
	portRe := regexp.MustCompile(":([0-9A-F]{4})")
	for _, address := range localAddresses {
		port := portRe.FindString(address)
		if port != "" {
			portString := string(port[1:])
			ports = append(ports, portString)
		}
	}
	return ports
}

// strHexToDecimal converts from string containing hex number to int
func strHexToDecimal(hex string) int {
	portInt, err := strconv.ParseInt(hex, 16, 64)
	fatal(err)
	return int(portInt)
}

// getOpenPorts gets a list of open/listening ports as integers
func getOpenPorts() (ports []int) {
	for _, port := range getHexPorts() {
		ports = append(ports, strHexToDecimal(port))
	}
	return ports

}

// Port parses /proc/net/tcp to determine if a given port is in an open state
// and returns an error if it is not.
func Port(port int) Thunk {
	return func() (exitCode int, exitMessage string) {
		for _, p := range getOpenPorts() {
			if p == port {
				return 0, ""
			}
		}
		return 1, "Port " + fmt.Sprint(port) + " did not respond."
	}
}

// Interface detects if a network interface exists
func Interface(name string) Thunk {
	// getInterfaceNames returns the names of all network interfaces
	getInterfaceNames := func() (interfaces []string) {
		ifaces, err := net.Interfaces()
		fatal(err)
		for _, iface := range ifaces {
			interfaces = append(interfaces, iface.Name)
		}
		return
	}
	return func() (exitCode int, exitMessage string) {
		for _, iface := range getInterfaceNames() {
			if iface == name {
				return 0, ""
			}
		}
		return 1, "Interface does not exist: " + name
	}
}

// Up determines if a network interface is up and running or not
func Up(name string) Thunk {
	return func() (exitCode int, exitMessage string) {
		interfaces, err := net.Interfaces()
		fatal(err)
		for _, iface := range interfaces {
			if iface.Name == name && iface.Flags&net.FlagUp != 0 {
				return 0, ""
			}
		}
		return 1, "Interface is down: " + name
	}
}

// hasIP checks to see if the interface that goes by name has the right address,
// given an IP version (4 or 6)
func hasIP(name string, address string, version int) bool {
	// ensure valid IP version
	if version != 4 && version != 6 {
		msg := "Misconfigured JSON: Unsupported IP version: "
		log.Fatal(msg + fmt.Sprint(version))
	}
	interfaces, err := net.Interfaces()
	fatal(err)
	for _, iface := range interfaces {
		addresses, err := iface.Addrs()
		fatal(err)
		// only check addresses if it is the correct interface
		if iface.Name == name {
			for _, addr := range addresses {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				if version == 4 && ip.To4().String() == address {
					return true
				} else if version == 6 && ip.To16().String() == address {
					return true
				}
			}
		}
	}
	return false
}

// Ip4 checks to see if this network interface has this ipv4 address
func Ip4(name string, address string) Thunk {
	return func() (exitCode int, exitMessage string) {
		if hasIP(name, address, 4) {
			return 0, ""
		}
		return 1, "Interface does not have IP: " + name + " " + address
	}
}

// Ip6 checks to see if this network interface has this ipv6 address
func Ip6(name string, address string) Thunk {
	return func() (exitCode int, exitMessage string) {
		if hasIP(name, address, 6) {
			return 0, ""
		}
		return 1, "Interface does not have IP: " + name + " " + address
	}
}

// Gateway checks to see that the default gateway has a certain IP
func Gateway(address string) Thunk {
	// getGatewayAddress filters all gateway IPs for a non-zero value
	getGatewayAddress := func() (addr string) {
		// first column in the output of route -n is the address of the gateway
		cmd := exec.Command("route", "-n")
		ips := commandColumnNoHeader(1, cmd)[1:] // has additional header row
		for _, ip := range ips {
			if ip != "0.0.0.0" {
				return ip
			}
		}
		return "0.0.0.0"
	}
	return func() (exitCode int, exitMessage string) {
		if address == getGatewayAddress() {
			return 0, ""
		}
		return 1, "Gateway does not have address: " + address
	}
}

// GatewayInterface checks that the default gateway is using a specified interface
func GatewayInterface(name string) Thunk {
	// getGatewayInterface returns the interface that the default gateway is
	// operating on
	getGatewayInterface := func() (iface string) {
		cmd := exec.Command("route", "-n")
		ips := commandColumnNoHeader(1, cmd)[1:] // has additional header row
		// calling the same command twice doesn't work
		cmd = exec.Command("route", "-n")
		names := commandColumnNoHeader(7, cmd)[1:] // has additional header row
		for i, ip := range ips {
			if ip != "0.0.0.0" {
				// TODO catch indexerror
				return names[i] // interface name
			}
		}
		return ""
	}
	return func() (exitCode int, exitMessage string) {
		if name == getGatewayInterface() {
			return 0, ""
		}
		return 1, "Default gateway does not operate on interface: " + name
	}
}

// Host checks if a given host can be resolved.
func Host(host string) Thunk {
	// resolvable  determines whether a given host can be reached
	resolvable := func(name string) bool {
		_, err := net.LookupHost(host)
		if err == nil {
			return true
		}
		return false
	}
	return func() (exitCode int, exitMessage string) {
		if resolvable(host) {
			return 0, ""
		}
		return 1, "Host cannot be resolved: " + host
	}
}

// canConnect tests whether a connection can be made to a given host on its
// given port using protocol ("TCP"|"UDP")
func canConnect(host string, protocol string) bool {
	parseerr := func(err error) {
		if err != nil {
			log.Fatal("Could not parse " + protocol + " address: " + host)
		}
	}
	switch protocol {
	case "TCP":
		tcpaddr, err := net.ResolveTCPAddr("tcp", host)
		parseerr(err)
		conn, err := net.DialTCP("tcp", nil, tcpaddr)
		if conn != nil {
			defer conn.Close()
		}
		if err == nil {
			return true
		}
		return false
	case "UDP":
		udpaddr, err := net.ResolveUDPAddr("udp", host)
		parseerr(err)
		conn, err := net.DialUDP("udp", nil, udpaddr)
		if conn != nil {
			defer conn.Close()
		}
		if err == nil {
			return true
		}
		return false
	default:
		log.Fatal("Unsupported protocol: " + protocol)
	}
	return false
}

// TCP sees ig a given IP/port can be reached with a TCP connection
func TCP(host string) Thunk {
	return func() (exitCode int, exitMessage string) {
		if canConnect(host, "TCP") {
			return 0, ""
		}
		return 1, "Could not connect over TCP to host: " + host
	}
}

// UDP is like TCP but with UDP instead.
func UDP(host string) Thunk {
	return func() (exitCode int, exitMessage string) {
		if canConnect(host, "UDP") {
			return 0, ""
		}
		return 1, "Could not connect over UDP to host: " + host
	}
}
