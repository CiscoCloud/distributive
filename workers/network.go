package workers

import (
	"fmt"
	"github.com/CiscoCloud/distributive/tabular"
	"github.com/CiscoCloud/distributive/wrkutils"
	log "github.com/Sirupsen/logrus"
	"net"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// port parses /proc/net/tcp to determine if a given port is in an open state
// and returns an error if it is not.
func port(parameters []string) (exitCode int, exitMessage string) {
	// getHexPorts gets all open ports as hex strings from /proc/net/tcp
	getHexPorts := func() (ports []string) {
		paths := [2]string{"/proc/net/tcp","/proc/net/udp"}
		for _,path := range paths {
			data := wrkutils.FileToString(path)
			table := tabular.ProbabalisticSplit(data)
			// TODO by header isn't working
			//localAddresses := tabular.GetColumnByHeader("local_address", table)
			localAddresses := tabular.GetColumnNoHeader(1, table)
			portRe := regexp.MustCompile(`([0-9A-F]{8}):([0-9A-F]{4})`)
			for _, address := range localAddresses {
				port := portRe.FindString(address)
				if port != "" {
					if len(port) < 10 {
						log.WithFields(log.Fields{
							"port":   port,
							"length": len(port),
						}).Fatal("Couldn't parse port number in " + path)
					}
					portString := string(port[9:])
					ports = append(ports, portString)
				}
			}
		}
		return ports
	}

	// strHexToDecimal converts from string containing hex number to int
	strHexToDecimal := func(hex string) int {
		portInt, err := strconv.ParseInt(hex, 16, 64)
		if err != nil {
			log.WithFields(log.Fields{
				"number": hex,
				"error":  err.Error(),
			}).Fatal("Couldn't parse hex number")
		}
		return int(portInt)
	}

	// getOpenPorts gets a list of open/listening ports as integers
	getOpenPorts := func() (ports []int) {
		for _, port := range getHexPorts() {
			ports = append(ports, strHexToDecimal(port))
		}
		return ports
	}

	// TODO check if it is in a valid range
	port := wrkutils.ParseMyInt(parameters[0])
	open := getOpenPorts()
	for _, p := range open {
		if p == port {
			return 0, ""
		}
	}
	// convert ports to string to send to wrkutils.GenericError
	var strPorts []string
	for _, port := range open {
		strPorts = append(strPorts, fmt.Sprint(port))
	}
	return wrkutils.GenericError("Port not open", fmt.Sprint(port), strPorts)
}

// getInterfaces returns a list of network interfaces and handles any associated
// error. Just for DRY.
func getInterfaces() []net.Interface {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatal("Could not read network interfaces")
	}
	return ifaces
}

// interfaceExists detects if a network interface exists,
func interfaceExists(parameters []string) (exitCode int, exitMessage string) {
	// getInterfaceNames returns the names of all network interfaces
	getInterfaceNames := func() (interfaces []string) {
		for _, iface := range getInterfaces() {
			interfaces = append(interfaces, iface.Name)
		}
		return
	}
	name := parameters[0]
	interfaces := getInterfaceNames()
	for _, iface := range interfaces {
		if iface == name {
			return 0, ""
		}
	}
	return wrkutils.GenericError("Interface does not exist", name, interfaces)
}

// up determines if a network interface is up and running or not
func up(parameters []string) (exitCode int, exitMessage string) {
	// getUpInterfaces returns all the names of the interfaces that are up
	getUpInterfaces := func() (interfaceNames []string) {
		for _, iface := range getInterfaces() {
			if iface.Flags&net.FlagUp != 0 {
				interfaceNames = append(interfaceNames, iface.Name)
			}
		}
		return interfaceNames

	}
	name := parameters[0]
	upInterfaces := getUpInterfaces()
	if tabular.StrIn(name, upInterfaces) {
		return 0, ""
	}
	return wrkutils.GenericError("Interface is not up", name, upInterfaces)
}

// getInterface IPs gets all the associated IP addresses of a given interface
// as a slice of strings, with a given IP protocol version (4|6)
func getInterfaceIPs(name string, version int) (ifaceAddresses []string) {
	// ensure valid IP version
	if version != 4 && version != 6 {
		log.WithFields(log.Fields{
			"version":  version,
			"expected": "4 | 6",
		}).Fatal("Probable configuration error: Unsupported IP version")
	}
	for _, iface := range getInterfaces() {
		if iface.Name == name {
			addresses, err := iface.Addrs()
			if err != nil {
				log.WithFields(log.Fields{
					"interface": iface.Name,
					"error":     err.Error(),
				}).Fatal("Could not get network addressed from interface")
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

// getIPWorker(exitCode int, exitMessage string) is an abstraction of Ip4 and Ip6
func getIPWorker(name string, address string, version int) (exitCode int, exitMessage string) {
	ips := getInterfaceIPs(name, version)
	if tabular.StrIn(address, ips) {
		return 0, ""
	}
	return wrkutils.GenericError("Interface does not have IP", address, ips)
}

// ip4 checks to see if this network interface has this ipv4 address
func ip4(parameters []string) (exitCode int, exitMessage string) {
	return getIPWorker(parameters[0], parameters[1], 4)
}

// ip6 checks to see if this network interface has this ipv6 address
func ip6(parameters []string) (exitCode int, exitMessage string) {
	return getIPWorker(parameters[0], parameters[1], 6)
}

// gateway checks to see that the default gateway has a certain IP
func gateway(parameters []string) (exitCode int, exitMessage string) {
	// getGatewayAddress filters all gateway IPs for a non-zero value
	getGatewayAddress := func() (addr string) {
		ips := routingTableColumn("Gateway")
		for _, ip := range ips {
			if ip != "0.0.0.0" {
				return ip
			}
		}
		return "0.0.0.0"
	}
	address := parameters[0]
	gatewayIP := getGatewayAddress()
	if address == gatewayIP {
		return 0, ""
	}
	msg := "Gateway does not have address"
	return wrkutils.GenericError(msg, address, []string{gatewayIP})
}

// gatewayInterface checks that the default gateway is using a specified interface
func gatewayInterface(parameters []string) (exitCode int, exitMessage string) {
	// getGatewayInterface returns the interface that the default gateway is
	// operating on
	getGatewayInterface := func() (iface string) {
		ips := routingTableColumn("Gateway")
		names := routingTableColumn("Iface")
		for i, ip := range ips {
			if ip != "0.0.0.0" {
				msg := "Fewer names in kernel routing table than IPs"
				wrkutils.IndexError(msg, i, names)
				return names[i] // interface name
			}
		}
		return ""
	}
	name := parameters[0]
	iface := getGatewayInterface()
	if name == iface {
		return 0, ""
	}
	msg := "Default gateway does not operate on interface"
	return wrkutils.GenericError(msg, name, []string{iface})
}

// host checks if a given host can be resolved.
func host(parameters []string) (exitCode int, exitMessage string) {
	// resolvable  determines whether a given host can be reached
	resolvable := func(name string) bool {
		_, err := net.LookupHost(name)
		if err == nil {
			return true
		}
		return false
	}
	host := parameters[0]
	if resolvable(host) {
		return 0, ""
	}
	return 1, "Host cannot be resolved: " + host
}

func validHost(host string) bool {
	_, err := net.ResolveTCPAddr("tcp", host)
	if err != nil && strings.Contains(err.Error(), "no such host") {
		return false
	}
	_, err = net.ResolveUDPAddr("udp", host)
	if err != nil && strings.Contains(err.Error(), "no such host") {
		return false
	}
	return true
}

// canConnect tests whether a connection can be made to a given host on its
// given port using protocol ("TCP"|"UDP")
func canConnect(host string, protocol string, timeout time.Duration) bool {
	if !validHost(host) {
		return false
	}
	resolveError := func(err error) {
		if err != nil {
			log.WithFields(log.Fields{
				"protocol": protocol,
				"address":  host,
				"error":    err.Error(),
			}).Fatal("Couldn't parse network address")
		}
	}
	var conn net.Conn
	var err error
	var timeoutNetwork = "tcp"
	var timeoutAddress string
	nanoseconds := timeout.Nanoseconds()
	switch protocol {
	case "TCP":
		tcpaddr, err := net.ResolveTCPAddr("tcp", host)
		resolveError(err)
		timeoutAddress = tcpaddr.String()
		if nanoseconds <= 0 {
			conn, err = net.DialTCP(timeoutNetwork, nil, tcpaddr)
		}
	case "UDP":
		timeoutNetwork = "udp"
		udpaddr, err := net.ResolveUDPAddr("udp", host)
		resolveError(err)
		timeoutAddress = udpaddr.String()
		if nanoseconds <= 0 {
			conn, err = net.DialUDP("udp", nil, udpaddr)
		}
	default:
		msg := "Probable configuration error: Unsupported protocol"
		log.WithField("protocol", protocol).Fatal(msg)
	}
	// if a duration was specified, use it
	if nanoseconds > 0 {
		conn, err = net.DialTimeout(timeoutNetwork, timeoutAddress, timeout)
		if conn != nil {
			defer conn.Close()
		}
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Warn("Error while connecting to host")
		}
	}
	if err == nil {
		return true
	}
	return false
}

// getConnection(exitCode int, exitMessage string) is an abstraction of TCP and UDP
func getConnectionWorker(host string, protocol string, timeoutstr string) (exitCode int, exitMessage string) {
	dur, err := time.ParseDuration(timeoutstr)
	if err != nil {
		log.WithFields(log.Fields{
			"duration": timeoutstr,
			"error":    err.Error(),
		}).Fatal("Probable configuration error: Could not parse duration")
	}
	if canConnect(host, protocol, dur) {
		return 0, ""
	}
	return 1, "Could not connect over " + protocol + " to host: " + host
}

// TODO add default port of :80 if none is provided
// tcp sees if a given IP/port can be reached with a TCP connection
func tcp(parameters []string) (exitCode int, exitMessage string) {
	return getConnectionWorker(parameters[0], "TCP", "0ns")
}

// udp is like TCP but with UDP instead.
func udp(parameters []string) (exitCode int, exitMessage string) {
	return getConnectionWorker(parameters[0], "UDP", "0ns")
}

// tcpTimeout is like TCP, but with a timeout parameter
func tcpTimeout(parameters []string) (exitCode int, exitMessage string) {
	return getConnectionWorker(parameters[0], "TCP", parameters[1])
}

// udpTimeout is like tcpTimeout but with UDP instead.
func udpTimeout(parameters []string) (exitCode int, exitMessage string) {
	return getConnectionWorker(parameters[0], "UDP", parameters[1])
}

// returns a column of the routing table as a slice of strings
func routingTableColumn(name string) []string {
	cmd := exec.Command("route", "-n")
	out := wrkutils.CommandOutput(cmd)
	table := tabular.ProbabalisticSplit(out)
	if len(table) < 1 {
		log.WithFields(log.Fields{
			"column": name,
			"table":  "\n" + tabular.ToString(table),
		}).Fatal("Routing table was not available or not properly parsed")
	}
	finalTable := table[1:] // has extra line before headers
	return tabular.GetColumnByHeader(name, finalTable)
}

// routingTableMatch(exitCode int, exitMessage string) constructs a Worker that returns whether or not the
// given string was found in the given column of the routing table. It is an
// astraction of routingTableDestination, routingTableInterface, and
// routingTableGateway
func routingTableMatch(col string, str string) (exitCode int, exitMessage string) {
	column := routingTableColumn(col)
	if tabular.StrIn(str, column) {
		return 0, ""
	}
	return wrkutils.GenericError("Not found in routing table", str, column)
}

// routingTableDestination checks if an IP address is a destination in the
// kernel's IP routing table, as accessed by `route -n`.
func routingTableDestination(parameters []string) (exitCode int, exitMessage string) {
	return routingTableMatch("Destination", parameters[0])
}

// routingTableInterface checks if a given name is an interface in the
// kernel's IP routing table, as accessed by `route -n`.
func routingTableInterface(parameters []string) (exitCode int, exitMessage string) {
	return routingTableMatch("Iface", parameters[0])
}

// routeTableGateway checks if an IP address is a gateway's IP in the
// kernel's IP routing table, as accessed by `route -n`.
func routingTableGateway(parameters []string) (exitCode int, exitMessage string) {
	return routingTableMatch("Gateway", parameters[0])
}

// responseMatchesGeneral is an abstraction of responseMatches and
// responseMatchesInsecure that simply varies in the security of the connection
func responseMatchesGeneral(parameters []string, secure bool) (exitCode int, exitMessage string) {
	urlstr := parameters[0]
	re := wrkutils.ParseUserRegex(parameters[1])
	body := wrkutils.URLToBytes(urlstr, secure)
	if re.Match(body) {
		return 0, ""
	}
	msg := "Response didn't match regexp"
	return wrkutils.GenericError(msg, re.String(), []string{string(body)})
}

// responseMatches asks: does the response from this URL match this regexp?
func responseMatches(parameters []string) (exitCode int, exitMessage string) {
	return responseMatchesGeneral(parameters, true)
}

// responseMatchesInsecure is just like responseMatches, but it doesn't verify
// the SSL cert on the other end.
func responseMatchesInsecure(parameters []string) (exitCode int, exitMessage string) {
	return responseMatchesGeneral(parameters, false)
}
