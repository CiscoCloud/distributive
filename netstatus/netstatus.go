// netstatus provides utility functions for querying several aspects of the
// network/host, especially as pertains to monitoring.
package netstatus

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

// GetHexPorts gets all open ports as hex strings from /proc/net/{tcp,udp}
// Its protocol argument can only be one of: "tcp" | "udp"
func GetHexPorts(protocol string) (ports []string) {
	var path string
	switch strings.ToLower(protocol) {
	case "tcp":
		path = "/proc/net/tcp"
	case "udp":
		path = "/proc/net/udp"
	default:
		log.WithFields(log.Fields{
			"protocol":        protocol,
			"valid protocols": "tcp|udp",
		}).Fatal("Invalid protocol passed to GetHexPorts!")
	}
	data := chkutil.FileToString(path)
	rowSep := regexp.MustCompile(`\n+`)
	colSep := regexp.MustCompile(`\s+`)
	table := tabular.SeparateString(rowSep, colSep, data)
	localAddresses := tabular.GetColumnByHeader("local_address", table)
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

// OpenPorts gets a list of open/listening TCP or UDP ports as integers from
// the information at /proc/net/tcp and /proc/net/udp, which may not reflect
// all of the ports that can be accessed externally.
// Its protocol argument can only be one of: "tcp" | "udp"
func OpenPorts(protocol string) (ports []uint16) {
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
	for _, port := range GetHexPorts(protocol) {
		ports = append(ports, uint16(strHexToDecimal(port)))
	}
	return ports
}

// PortOpen reports whether or not the given (decimal) port is open
// Its protocol argument can only be one of: "tcp" | "udp"
func PortOpen(protocol string, port uint16) bool {
	dur, err := time.ParseDuration("5s")
	if err != nil {
		log.Fatal(err)
	}
	return CanConnect(fmt.Sprintf("localhost:%v", port), protocol, dur)
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
	var conn net.Conn
	var err error
	var timeoutNetwork = "tcp"
	var timeoutAddress string
	nanoseconds := timeout.Nanoseconds()
	switch strings.ToUpper(protocol) {
	case "TCP":
		tcpaddr, err := net.ResolveTCPAddr("tcp", host)
		if err != nil {
			return false
		}
		timeoutAddress = tcpaddr.String()
		if nanoseconds <= 0 {
			conn, err = net.DialTCP(timeoutNetwork, nil, tcpaddr)
		}
	case "UDP":
		timeoutNetwork = "udp"
		udpaddr, err := net.ResolveUDPAddr("udp", host)
		if err != nil {
			return false
		}
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
