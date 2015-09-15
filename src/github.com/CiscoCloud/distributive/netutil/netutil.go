/*
Package netutil provides some basic networking utilities, especially for health
checking.
*/
package netutil

import (
	"fmt"
	"github.com/CiscoCloud/distributive/chkutil"
	"github.com/CiscoCloud/distributive/tabular"
	log "github.com/Sirupsen/logrus"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// GetHexPorts gets all open ports as hex strings from /proc/net/tcp
func GetHexPorts() (ports []string) {
	path := "/proc/net/tcp"
	data := chkutil.FileToString(path)
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
	return ports
}

// OpenPorts gets a list of open/listening ports as integers
func OpenPorts() (ports []uint16) {
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
	for _, port := range GetHexPorts() {
		ports = append(ports, uint16(strHexToDecimal(port)))
	}
	return ports
}

// PortOpen reports whether or not the given (decimal) port is open
func PortOpen(port uint16) bool {
	uint16In := func(n uint16, slc []uint16) bool {
		for _, nPrime := range slc {
			if n == nPrime {
				return true
			}
		}
		return false
	}
	return uint16In(port, OpenPorts())
}

// ValidIP returns a boolean answering the question "is this a valid IPV4/6
// address?
func ValidIP(ipStr string) bool { return (net.ParseIP(ipStr) != nil) }

// GetInterfaces returns a list of network interfaces and handles any associated
// error. Just for DRY.
func GetInterfaces() []net.Interface {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatal("Could not read network interfaces")
	}
	return ifaces
}

// InterfaceIPs gets all the associated IP addresses of a given interface.
// If the interface can't be found, it will return nil.
func InterfaceIPs(name string) (ifaceAddresses []*net.IP) {
	for _, iface := range GetInterfaces() {
		if iface.Name == name {
			addresses, err := iface.Addrs()
			if err != nil {
				log.WithFields(log.Fields{
					"interface": iface.Name,
					"error":     err.Error(),
				}).Fatal("Could not get network addresses from interface.")
			}
			for _, addr := range addresses {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				ifaceAddresses = append(ifaceAddresses, &ip)
			}
			return ifaceAddresses
		}
	} // will only reach this line if the interface didn't exist
	return nil // will be empty
}

// Resolvable checks if the given host can be resolved on the TCP and UDP nets
func Resolvable(host string) bool {
	_, err := net.LookupHost(host)
	if err == nil {
		return true
	}
	return false
}

// CanConnect tests whether a connection can be made to a given host on its
// given port using protocol ("TCP"|"UDP")
func CanConnect(host string, protocol string, timeout time.Duration) bool {
	// TODO resolvable always fails
	if !Resolvable(host) {
		fmt.Println("Couldn't resolve host")
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
	switch strings.ToUpper(protocol) {
	case "TCP":
		fmt.Println("in TCP")
		tcpaddr, err := net.ResolveTCPAddr("tcp", host)
		resolveError(err)
		timeoutAddress = tcpaddr.String()
		if nanoseconds <= 0 {
			conn, err = net.DialTCP(timeoutNetwork, nil, tcpaddr)
		}
	case "UDP":
		fmt.Println("in UDP")
		timeoutNetwork = "udp"
		udpaddr, err := net.ResolveUDPAddr("udp", host)
		resolveError(err)
		timeoutAddress = udpaddr.String()
		if nanoseconds <= 0 {
			// TODO why the inconsistency here?
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
