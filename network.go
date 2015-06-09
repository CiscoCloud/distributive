package main

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"regexp"
	"strconv"
)

// getHexPorts gets all open ports as hex strings from /proc/net/tcp
func getHexPorts() (ports []string) {
	data := fileToString("/proc/net/tcp")
	localAddresses := getColumnNoHeader(1, stringToSlice(data))
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
	if err != nil {
		log.Fatal("Couldn't parse hex number " + hex + ":\n\t" + err.Error())
	}
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
		open := getOpenPorts()
		for _, p := range open {
			if p == port {
				return 0, ""
			}
		}
		// Convert ports to string to send to notInError
		var strPorts []string
		for _, port := range open {
			strPorts = append(strPorts, fmt.Sprint(port))
		}
		return notInError("Port not open", fmt.Sprint(port), strPorts)
	}
}

// getInterfaces returns a list of network interfaces and handles any associated
// error. Just for DRY.
func getInterfaces() []net.Interface {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Fatal("Could not read network interfaces:\n\t" + err.Error())
	}
	return ifaces
}

// Interface detects if a network interface exists
func Interface(name string) Thunk {
	// getInterfaceNames returns the names of all network interfaces
	getInterfaceNames := func() (interfaces []string) {
		for _, iface := range getInterfaces() {
			interfaces = append(interfaces, iface.Name)
		}
		return
	}
	return func() (exitCode int, exitMessage string) {
		interfaces := getInterfaceNames()
		for _, iface := range interfaces {
			if iface == name {
				return 0, ""
			}
		}
		return notInError("Interface does not exist", name, interfaces)
	}
}

// Up determines if a network interface is up and running or not
func Up(name string) Thunk {
	// getUpInterfaces returns all the names of the interfaces that are up
	getUpInterfaces := func() (interfaceNames []string) {
		for _, iface := range getInterfaces() {
			if iface.Flags&net.FlagUp != 0 {
				interfaceNames = append(interfaceNames, iface.Name)
			}
		}
		return interfaceNames

	}
	return func() (exitCode int, exitMessage string) {
		upInterfaces := getUpInterfaces()
		if strIn(name, upInterfaces) {
			return 0, ""
		}
		return notInError("Interface is not up", name, upInterfaces)
	}
}

// getIPs gets all the associated IP addresses of a given interface as a slice
// of strings, with a given IP protocol version (4|6)
func getInterfaceIPs(name string, version int) (ifaceAddresses []string) {
	// ensure valid IP version
	if version != 4 && version != 6 {
		msg := "Misconfigured JSON: Unsupported IP version: "
		log.Fatal(msg + fmt.Sprint(version))
	}
	for _, iface := range getInterfaces() {
		if iface.Name == name {
			addresses, err := iface.Addrs()
			if err != nil {
				msg := "Could not get network addressed from interface: "
				msg += "\n\tInterface name: " + iface.Name
				msg += "\n\tError: " + err.Error()
				log.Fatal(msg)
			}
			for _, addr := range addresses {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				switch version {
				case 4:
					ifaceAddresses = append(ifaceAddresses, ip.To4().String())
				case 6:
					ifaceAddresses = append(ifaceAddresses, ip.To16().String())
				}
			}
			return ifaceAddresses

		}
	} // will only reach this line if the interface didn't exist
	return ifaceAddresses // will be empty
}

// getIPThunk is an abstraction of Ip4 and Ip6
func getIPThunk(name string, address string, version int) Thunk {
	return func() (exitCode int, exitMessage string) {
		ips := getInterfaceIPs(name, version)
		if strIn(address, ips) {
			return 0, ""
		}
		return notInError("Interface does not have IP", address, ips)
	}
}

// Ip4 checks to see if this network interface has this ipv4 address
func Ip4(name string, address string) Thunk {
	return getIPThunk(name, address, 4)
}

// Ip6 checks to see if this network interface has this ipv6 address
func Ip6(name string, address string) Thunk {
	return getIPThunk(name, address, 6)
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
		gatewayIP := getGatewayAddress()
		if address == gatewayIP {
			return 0, ""
		}
		msg := "Gateway does not have address"
		return notInError(msg, address, []string{gatewayIP})
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
		iface := getGatewayInterface()
		if name == iface {
			return 0, ""
		}
		msg := "Default gateway does not operate on interface"
		return notInError(msg, name, []string{iface})
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

// getConnectionThunk is an abstraction of TCP and UDP
func getConnectionThunk(host string, protocol string) Thunk {
	return func() (exitCode int, exitMessage string) {
		if canConnect(host, protocol) {
			return 0, ""
		}
		return 1, "Could not connect over " + protocol + " to host: " + host
	}
}

// TCP sees ig a given IP/port can be reached with a TCP connection
func TCP(host string) Thunk {
	return getConnectionThunk(host, "TCP")
}

// UDP is like TCP but with UDP instead.
func UDP(host string) Thunk {
	return getConnectionThunk(host, "UDP")
}
