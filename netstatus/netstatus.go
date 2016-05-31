// netstatus provides utility functions for querying several aspects of the
// network/host, especially as pertains to monitoring.
package netstatus

import (
	"fmt"
	"net"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

// PortOpen reports whether or not the given (decimal) port is open
// Its protocol argument can only be one of: "tcp" | "udp"
func PortOpen(protocol string, port uint16) bool {
	_, err := net.Dial(protocol, fmt.Sprintf("localhost:%v", port))
	return err == nil
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
func CanConnect(address, protocol string, timeout time.Duration) bool {
	var err error
	if timeout > 0 {
		_, err = net.Dial(strings.ToLower(protocol), address)
	} else {
		_, err = net.DialTimeout(strings.ToLower(protocol), address, timeout)
	}
	if err == nil {
		return true
	}
	return false
}
