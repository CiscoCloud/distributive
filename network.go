package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// getHexPorts gets all open ports as hex strings from /proc/net/tcp
func getHexPorts() (ports []string) {
	// TODO use stringToSlice here
	toReturn := []string{}
	tcp, err := ioutil.ReadFile("/proc/net/tcp")
	fatal(err)
	// matches only the beginnings of lines
	lines := bytes.Split(tcp, []byte("\n"))
	portRe := regexp.MustCompile(":([0-9A-F]{4})")
	for _, line := range lines {
		port := portRe.Find(line) // only get first port, which is local
		if port == nil {
			continue
		}
		portString := string(port[1:])
		toReturn = append(toReturn, portString)
	}
	return toReturn
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

// routeOutput returns the output of the command `route -n`, split into lines
// and on whitespace, without header rows. First coordinate is row, second is column.
func routeOutput() (toReturn [][]string) {
	// TODO use stringToSlice here
	out, err := exec.Command("route", "-n").Output()
	fatal(err)
	if out == nil {
		log.Fatal("Couldn't read output from `route -n`")
	}
	lines := strings.Split(string(out), "\n")
	// cut out headers
	if len(lines) == 0 {
		return
	}
	// split on whitespace
	re := regexp.MustCompile("\\s+")
	for _, line := range lines {
		if line != "" {
			toReturn = append(toReturn, re.Split(line, -1))
		}
	}
	return toReturn[2:]
}

// Gateway checks to see that the default gateway has a certain IP
func Gateway(address string) Thunk {
	// getGatewayAddress filters all gateway IPs for a non-zero value
	getGatewayAddress := func() (addr string) {
		// getAllGatewayAddresses returns all the IPs of all gateways
		getAllGatewayAddresses := func() (addresses []string) {
			out := routeOutput() // 2D array, out[row][col]
			for _, row := range out {
				gatewayIP := row[1]
				addresses = append(addresses, gatewayIP)
			}
			return addresses
		}
		for _, ip := range getAllGatewayAddresses() {
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
		for _, row := range routeOutput() {
			if row[1] != "0.0.0.0" {
				return row[7] // interface name
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
