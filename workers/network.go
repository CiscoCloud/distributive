package workers

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/CiscoCloud/distributive/tabular"
	"github.com/CiscoCloud/distributive/wrkutils"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

// RegisterNetwork registers these checks so they can be used.
func RegisterNetwork() {
	wrkutils.RegisterCheck("port", port, 1)
	wrkutils.RegisterCheck("interface", interfaceExists, 1)
	wrkutils.RegisterCheck("up", up, 1)
	wrkutils.RegisterCheck("ip4", ip4, 2)
	wrkutils.RegisterCheck("ip6", ip6, 2)
	wrkutils.RegisterCheck("gateway", gateway, 1)
	wrkutils.RegisterCheck("gatewayinterface", gatewayInterface, 1)
	wrkutils.RegisterCheck("host", host, 1)
	wrkutils.RegisterCheck("tcp", tcp, 1)
	wrkutils.RegisterCheck("udp", udp, 1)
	wrkutils.RegisterCheck("tcptimeout", tcpTimeout, 2)
	wrkutils.RegisterCheck("udptimeout", udpTimeout, 2)
	wrkutils.RegisterCheck("routingtabledestination", routingTableDestination, 1)
	wrkutils.RegisterCheck("routingtableinterface", routingTableInterface, 1)
	wrkutils.RegisterCheck("routingtablegateway", routingTableGateway, 1)
	wrkutils.RegisterCheck("responsematches", responseMatches, 2)
	wrkutils.RegisterCheck("responsematchesinsecure", responseMatchesInsecure, 2)
}

// port parses /proc/net/tcp to determine if a given port is in an open state
// and returns an error if it is not.
func port(parameters []string) (exitCode int, exitMessage string) {
	// getHexPorts gets all open ports as hex strings from /proc/net/tcp
	getHexPorts := func() (ports []string) {
		data := wrkutils.FileToString("/proc/net/tcp")
		table := tabular.ProbabalisticSplit(data)
		// TODO by header isn't working
		//localAddresses := tabular.GetColumnByHeader("local_address", table)
		localAddresses := tabular.GetColumnNoHeader(1, table)
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
	strHexToDecimal := func(hex string) int {
		portInt, err := strconv.ParseInt(hex, 16, 64)
		if err != nil {
			log.Fatal("Couldn't parse hex number " + hex + ":\n\t" + err.Error())
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
		log.Fatal("Could not read network interfaces:\n\t" + err.Error())
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
		ips := routingTableColumn(1)
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
		ips := routingTableColumn(1)
		names := routingTableColumn(1)
		for i, ip := range ips {
			if ip != "0.0.0.0" {
				if len(names) < i {
					msg := "Fewer names in kernel routing table than IPs:"
					msg += "\n\tNames: " + fmt.Sprint(names)
					msg += "\n\tIPs: " + fmt.Sprint(ips)
					log.Fatal(msg)
				}
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

// canConnect tests whether a connection can be made to a given host on its
// given port using protocol ("TCP"|"UDP")
func canConnect(host string, protocol string, timeout time.Duration) bool {
	parseerr := func(err error) {
		if err != nil {
			log.Fatal("Could not parse " + protocol + " address: " + host)
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
		parseerr(err)
		timeoutAddress = tcpaddr.String()
		if nanoseconds <= 0 {
			conn, err = net.DialTCP(timeoutNetwork, nil, tcpaddr)
		}
	case "UDP":
		timeoutNetwork = "udp"
		udpaddr, err := net.ResolveUDPAddr("udp", host)
		parseerr(err)
		timeoutAddress = udpaddr.String()
		if nanoseconds <= 0 {
			conn, err = net.DialUDP("udp", nil, udpaddr)
		}
	default:
		log.Fatal("Unsupported protocol: " + protocol)
	}
	// if a duration was specified, use it
	if nanoseconds > 0 {
		conn, err = net.DialTimeout(timeoutNetwork, timeoutAddress, timeout)
		if err != nil {
			fmt.Println(err)
		}
	}
	if conn != nil {
		defer conn.Close()
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
		msg := "Configuration error: Could not parse duration: "
		log.Fatal(msg + timeoutstr)
	}
	if canConnect(host, protocol, dur) {
		return 0, ""
	}
	return 1, "Could not connect over " + protocol + " to host: " + host
}

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
func routingTableColumn(column int) []string {
	cmd := exec.Command("route", "-n")
	return wrkutils.CommandColumnNoHeader(column, cmd)[1:]
}

// routingTableMatch(exitCode int, exitMessage string) constructs a Worker that returns whether or not the
// given string was found in the given column of the routing table. It is an
// astraction of routingTableDestination, routingTableInterface, and
// routingTableGateway
func routingTableMatch(col int, str string) (exitCode int, exitMessage string) {
	column := routingTableColumn(col)
	if tabular.StrIn(str, column) {
		return 0, ""
	}
	return wrkutils.GenericError("Not found in routing table", str, column)
}

// routingTableDestination checks if an IP address is a destination in the
// kernel's IP routing table, as accessed by `route -n`.
func routingTableDestination(parameters []string) (exitCode int, exitMessage string) {
	return routingTableMatch(0, parameters[0])
}

// routingTableInterface checks if a given name is an interface in the
// kernel's IP routing table, as accessed by `route -n`.
func routingTableInterface(parameters []string) (exitCode int, exitMessage string) {
	return routingTableMatch(7, parameters[0])
}

// routeTableGateway checks if an IP address is a gateway's IP in the
// kernel's IP routing table, as accessed by `route -n`.
func routingTableGateway(parameters []string) (exitCode int, exitMessage string) {
	return routingTableMatch(1, parameters[0])
}

// URLToBytes gets the response from urlstr and returns it as a byte string
// TODO: allow insecure requests
// http://stackoverflow.com/questions/12122159/golang-how-to-do-a-https-request-with-bad-certificate
func URLToBytes(urlstr string, secure bool) []byte {
	// create http client
	transport := &http.Transport{}
	if !secure {
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	client := &http.Client{Transport: transport}
	// get response from URL
	resp, err := client.Get(urlstr)
	if err != nil {
		wrkutils.CouldntReadError(urlstr, err)
	}
	defer resp.Body.Close()

	// read response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		msg := "Bad response, couldn't read body:"
		msg += "\n\tURL: " + urlstr
		msg += "\n\tError: " + err.Error()
		log.Fatal(msg)
	} else if body == nil || bytes.Equal(body, []byte{}) {
		msg := "Body of response was empty:"
		msg += "\n\tURL: " + urlstr
		log.Fatal(msg)
	}
	return body
}

// responseMatchesGeneral is an abstraction of responseMatches and
// responseMatchesInsecure that simply varies in the security of the connection
func responseMatchesGeneral(parameters []string, secure bool) (exitCode int, exitMessage string) {
	urlstr := parameters[0]
	re := wrkutils.ParseUserRegex(parameters[1])
	body := URLToBytes(urlstr, secure)
	if re.Match(body) {
		return 0, ""
	}
	msg := "Response didn't match regexp:"
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
