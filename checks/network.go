package checks

import (
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	"github.com/CiscoCloud/distributive/chkutil"
	"github.com/CiscoCloud/distributive/errutil"
	"github.com/CiscoCloud/distributive/netstatus"
	"github.com/CiscoCloud/distributive/tabular"
	log "github.com/Sirupsen/logrus"
)

// parsePort determines whether or not this string represents a valid port
// number, and returns it if so, and an error if not.
func parsePort(portStr string) (uint16, error) {
	portInt, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil || portInt < 0 || portInt > 65535 {
		return 0, err
	}
	return uint16(portInt), nil
}

/*
#### Port
Description: Is this port open?
Parameters:
- Number (uint16): Port number (decimal)
Example parameters:
- 80, 8080, 8500, 5050
Dependencies:
- /proc/net/tcp
- /proc/net/udp
*/

type Port struct{ port uint16 }

func (chk Port) ID() string { return "Port" }
func init() {
    chkutil.Register("Port", func() chkutil.Check {
        return &Port{}
    })
    chkutil.Register("PortTCP", func() chkutil.Check {
        return &PortTCP{}
    })
    chkutil.Register("PortUDP", func() chkutil.Check {
        return &PortUDP{}
    })
    chkutil.Register("Up", func() chkutil.Check {
        return &Up{}
    })
    chkutil.Register("InterfaceExists", func() chkutil.Check {
        return &InterfaceExists{}
    })
    chkutil.Register("IP", func() chkutil.Check {
        return &IP4{}
    })
    chkutil.Register("IP6", func() chkutil.Check {
        return &IP6{}
    })
    chkutil.Register("RoutingTableGateway", func() chkutil.Check {
        return &RoutingTableGateway{}
    })
    chkutil.Register("RoutingTableDestination", func() chkutil.Check {
        return &RoutingTableDestination{}
    })
    chkutil.Register("RoutingTableInterface", func() chkutil.Check {
        return &RoutingTableInterface{}
    })
    chkutil.Register("Gateway", func() chkutil.Check {
        return &Gateway{}
    })
    chkutil.Register("GatewayInterface", func() chkutil.Check {
        return &GatewayInterface{}
    })
    chkutil.Register("ResponseMatches", func() chkutil.Check {
        return &ResponseMatches{}
    })
    chkutil.Register("ResponseMatchesInsecure", func() chkutil.Check {
        return &ResponseMatchesInsecure{}
    })
    chkutil.Register("TCP", func() chkutil.Check {
        return &TCP{}
    })
    chkutil.Register("TCPTimeout", func() chkutil.Check {
        return &TCPTimeout{}
    })
    chkutil.Register("UDPTimeout", func() chkutil.Check {
        return &UDPTimeout{}
    })
    chkutil.Register("UDP", func() chkutil.Check {
        return &UDP{}
    })
}

func (chk Port) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	} else if portInt, err := parsePort(params[0]); err == nil {
		chk.port = portInt
	} else {
		return chk, errutil.ParameterTypeError{params[0], "uint16"}
	}
	return chk, nil
}

func (chk Port) Status() (int, string, error) {
	if netstatus.PortOpen("tcp", chk.port) || netstatus.PortOpen("udp", chk.port) {
		return errutil.Success()
	}
	// convert ports to string to send to errutil.GenericError
	var strPorts []string
	openPorts := append(netstatus.OpenPorts("tcp"), netstatus.OpenPorts("udp")...)
	for _, port := range openPorts {
		strPorts = append(strPorts, fmt.Sprint(port))
	}
	return errutil.GenericError("Port not open", fmt.Sprint(chk.port), strPorts)
}

/*
#### PortTCP
Description: Is this port open on the TCP protocol?
Parameters:
- Number (uint16): Port number (decimal)
Example parameters:
- 80, 8080, 8500, 5050
Dependencies:
- /proc/net/tcp
*/

type PortTCP struct{ port uint16 }

func (chk PortTCP) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	} else if portInt, err := parsePort(params[0]); err == nil {
		chk.port = portInt
	} else {
		return chk, errutil.ParameterTypeError{params[0], "uint16"}
	}
	return chk, nil
}

func (chk PortTCP) Status() (int, string, error) {
	if netstatus.PortOpen("tcp", chk.port) {
		return errutil.Success()
	}
	// convert ports to string to send to errutil.GenericError
	var strPorts []string
	for _, port := range netstatus.OpenPorts("tcp") {
		strPorts = append(strPorts, fmt.Sprint(port))
	}
	return errutil.GenericError("Port not open", fmt.Sprint(chk.port), strPorts)
}

/*
#### PortUDP
Description: Is this port open on the UDP protocol?
Parameters:
- Number (uint16): Port number (decimal)
Example parameters:
- 80, 8080, 8500, 5050
Dependencies:
- /proc/net/udp
*/

type PortUDP struct{ port uint16 }

func (chk PortUDP) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	} else if portInt, err := parsePort(params[0]); err == nil {
		chk.port = portInt
	} else {
		return chk, errutil.ParameterTypeError{params[0], "uint16"}
	}
	return chk, nil
}

func (chk PortUDP) Status() (int, string, error) {
	if netstatus.PortOpen("udp", chk.port) {
		return errutil.Success()
	}
	// convert ports to string to send to errutil.GenericError
	var strPorts []string
	for _, port := range netstatus.OpenPorts("udp") {
		strPorts = append(strPorts, fmt.Sprint(port))
	}
	return errutil.GenericError("Port not open", fmt.Sprint(chk.port), strPorts)
}

/*
#### InterfaceExists
Description: Does this interface exist?
Parameters:
- Name (string): name of the interface
Example parameters:
- lo, wlp1s0, docker0
*/

type InterfaceExists struct{ name string }

func (chk InterfaceExists) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	chk.name = params[0]
	return chk, nil
}

func (chk InterfaceExists) Status() (int, string, error) {
	// getInterfaceNames returns the names of all network interfaces
	getInterfaceNames := func() (interfaces []string) {
		for _, iface := range netstatus.GetInterfaces() {
			interfaces = append(interfaces, iface.Name)
		}
		return
	}
	interfaces := getInterfaceNames()
	for _, iface := range interfaces {
		if iface == chk.name {
			return errutil.Success()
		}
	}
	return errutil.GenericError("Interface does not exist", chk.name, interfaces)
}

/*
#### Up
Description: Is this interface up?
Parameters:
- Name (string): name of the interface
Example parameters:
- lo, wlp1s0, docker0
*/

type Up struct{ name string }

func (chk Up) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	chk.name = params[0]
	return chk, nil
}

func (chk Up) Status() (int, string, error) {
	// getUpInterfaces returns all the names of the interfaces that are up
	getUpInterfaces := func() (interfaceNames []string) {
		for _, iface := range netstatus.GetInterfaces() {
			if iface.Flags&net.FlagUp != 0 {
				interfaceNames = append(interfaceNames, iface.Name)
			}
		}
		return interfaceNames

	}
	upInterfaces := getUpInterfaces()
	if tabular.StrIn(chk.name, upInterfaces) {
		return errutil.Success()
	}
	return errutil.GenericError("Interface is not up", chk.name, upInterfaces)
}

// ipCheck(int, string, error) is an abstraction of IP4 and
// IP6
func ipCheck(name string, address *net.IP, version int) (int, string, error) {
	ips := netstatus.InterfaceIPs(name)
	for _, ip := range ips {
		if ip.Equal(*address) {
			return errutil.Success()
		}
	}
	return errutil.GenericError("Interface does not have IP", address, ips)
}

/*
#### IP4
Description: Does this interface have this IPV4 address?
Parameters:
- Interface name (string)
- Address (IP address)
Example parameters:
- lo, wlp1s0, docker0
- 192.168.0.21, 222.111.0.22
*/

type IP4 struct {
	name string
	ip   net.IP
}

func (chk IP4) New(params []string) (chkutil.Check, error) {
	if len(params) != 2 {
		return chk, errutil.ParameterLengthError{2, params}
	} else if !netstatus.ValidIP(params[1]) {
		return chk, errutil.ParameterTypeError{params[1], "IP"}
	}
	chk.name = params[0]
	chk.ip = net.ParseIP(params[1])
	return chk, nil
}

func (chk IP4) Status() (int, string, error) {
	// TODO figure out IP pointer situation
	return ipCheck(chk.name, &chk.ip, 4)
}

/*
#### IP6
Description: Does this interface have this IPV6 address?
Parameters:
- Interface name (string)
- IP (IP address)
Example parameters:
- lo, wlp1s0, docker0
- FE80:0000:0000:0000:0202:B3FF:FE1E:8329, 2001:db8:0:1:1:1:1:1
*/

type IP6 struct {
	name string
	ip   net.IP
}

func (chk IP6) New(params []string) (chkutil.Check, error) {
	if len(params) != 2 {
		return chk, errutil.ParameterLengthError{2, params}
	} else if !netstatus.ValidIP(params[1]) {
		return chk, errutil.ParameterTypeError{params[1], "IP"}
	}
	chk.name = params[0]
	chk.ip = net.ParseIP(params[1])
	return chk, nil
}

func (chk IP6) Status() (int, string, error) {
	return ipCheck(chk.name, &chk.ip, 6)
}

/*
#### Gateway
Description: Does the default Gateway have this IP?
Parameters:
- IP (IP address)
Example parameters:
- 192.168.0.21, 222.111.0.22
*/

type Gateway struct{ ip net.IP }

func (chk Gateway) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	} else if !netstatus.ValidIP(params[0]) {
		return chk, errutil.ParameterTypeError{params[0], "IP"}
	}
	chk.ip = net.ParseIP(params[0])
	return chk, nil
}

func (chk Gateway) Status() (int, string, error) {
	// getGatewayAddress filters all Gateway IPs for a non-zero value
	getGatewayAddress := func() (addr string) {
		ips := RoutingTableColumn("Gateway")
		for _, ip := range ips {
			if ip != "0.0.0.0" {
				return ip
			}
		}
		return "0.0.0.0"
	}
	GatewayIP := getGatewayAddress()
	if chk.ip.String() == GatewayIP {
		return errutil.Success()
	}
	msg := "Gateway does not have address"
	return errutil.GenericError(msg, chk.ip.String(), []string{GatewayIP})
}

/*
#### GatewayInterface
Description: Is the default Gateway is using a specified interface?
Parameters:
- Name (string)
Example parameters:
- lo, wlp1s0, docker0
*/

type GatewayInterface struct{ name string }

func (chk GatewayInterface) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	chk.name = params[0]
	return chk, nil
}

func (chk GatewayInterface) Status() (int, string, error) {
	// getGatewayInterface returns the interface that the default Gateway is
	// operating on
	getGatewayInterface := func() (iface string) {
		ips := RoutingTableColumn("Gateway")
		names := RoutingTableColumn("Iface")
		for i, ip := range ips {
			if ip != "0.0.0.0" {
				msg := "Fewer names in kernel routing table than IPs"
				errutil.IndexError(msg, i, names)
				return names[i] // interface name
			}
		}
		return ""
	}
	iface := getGatewayInterface()
	if chk.name == iface {
		return errutil.Success()
	}
	msg := "Default Gateway does not operate on interface"
	return errutil.GenericError(msg, chk.name, []string{iface})
}

/*
#### Host
Description: Host checks if a given host can be resolved
Parameters:
Example parameters:
*/

type Host struct{ hostname string }

func (chk Host) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	chk.hostname = params[0]
	return chk, nil
}

func (chk Host) Status() (int, string, error) {
	if netstatus.Resolvable(chk.hostname) {
		return errutil.Success()
	}
	return 1, "Host cannot be resolved: " + chk.hostname, nil
}

/*
#### TCP
Description: Can a given IP/port can be reached with a TCP connection
Parameters:
Example parameters:
- 192.168.0.21, 222.111.0.22
*/

type TCP struct{ address string }

func (chk TCP) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	chk.address = params[0]
	return chk, nil
}

func (chk TCP) Status() (int, string, error) {
	if _, err := net.Dial("tcp", chk.address); err == nil {
		return errutil.Success()
	}
	return 1, fmt.Sprintf("Couldn't connect to %s", chk.address), nil
}

/*
#### UDP
Description: Like TCP but with UDP instead.
*/

type UDP struct{ address string }

func (chk UDP) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	chk.address = params[0]
	return chk, nil
}

func (chk UDP) Status() (int, string, error) {
	if _, err := net.Dial("udp", chk.address); err == nil {
		return errutil.Success()
	}
	return 1, fmt.Sprintf("Couldn't connect to %s", chk.address), nil
}

/*
#### TCPTimeout
Description: Like TCP, but with a second parameter of a timeout
Example parameters:
- 5s, 7μs, 12m, 5h, 3d
*/

type TCPTimeout struct {
	address string
	timeout time.Duration
}

func (chk TCPTimeout) New(params []string) (chkutil.Check, error) {
	if len(params) != 2 {
		return chk, errutil.ParameterLengthError{2, params}
	}
	chk.address = params[0]
	duration, err := time.ParseDuration(params[1])
	if err != nil {
		return chk, errutil.ParameterTypeError{params[1], "time.Duration"}
	}
	chk.timeout = duration
	return chk, nil
}

func (chk TCPTimeout) Status() (int, string, error) {
	if _, err := net.DialTimeout("udp", chk.address, chk.timeout); err == nil {
		return errutil.Success()
	}
	return 1, fmt.Sprintf("Couldn't connect to %s", chk.address), nil
}

/*
#### UDPTimeout
Description: Like TCPTimeout, but with UDP
*/

type UDPTimeout struct {
	address string
	timeout time.Duration
}

func (chk UDPTimeout) New(params []string) (chkutil.Check, error) {
	if len(params) != 2 {
		return chk, errutil.ParameterLengthError{2, params}
	}
	chk.address = params[0]
	duration, err := time.ParseDuration(params[1])
	if err != nil {
		return chk, errutil.ParameterTypeError{params[1], "time.Duration"}
	}
	chk.timeout = duration
	return chk, nil
}

func (chk UDPTimeout) Status() (int, string, error) {
	if _, err := net.DialTimeout("udp", chk.address, chk.timeout); err == nil {
		return errutil.Success()
	}
	return 1, fmt.Sprintf("Couldn't connect to %s", chk.address), nil
}

// returns a column of the routing table as a slice of strings
// TODO read from /proc/net/route instead
func RoutingTableColumn(name string) []string {
	cmd := exec.Command("route", "-n")
	out := chkutil.CommandOutput(cmd)
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

// RoutingTableMatch asks: Is this value in this column of the routing table?
func RoutingTableMatch(col string, str string) (int, string, error) {
	column := RoutingTableColumn(col)
	if tabular.StrIn(str, column) {
		return errutil.Success()
	}
	return errutil.GenericError("Not found in routing table", str, column)
}

/*
#### RoutingTableDestination
Description: Is this IP address in the kernel's IP routing table?
Parameters:
- IP (IP address)
Example parameters:
- 192.168.0.21, 222.111.0.22
Dependencies:
- `route -n`
*/

type RoutingTableDestination struct{ ip net.IP }

func (chk RoutingTableDestination) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	if !netstatus.ValidIP(params[0]) {
		return chk, errutil.ParameterTypeError{params[0], "IP address"}
	}
	chk.ip = net.ParseIP(params[0])
	return chk, nil
}

func (chk RoutingTableDestination) Status() (int, string, error) {
	return RoutingTableMatch("Destination", chk.ip.To4().String())
}

/*
#### RoutingTableInterface
Description: Is this interface in the kernel's IP routing table?
Parameters:
- Name (string)
Example parameters:
- lo, wlp1s0, docker0
Dependencies:
- `route -n`
*/

type RoutingTableInterface struct{ name string }

func (chk RoutingTableInterface) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	chk.name = params[0]
	return chk, nil
}

func (chk RoutingTableInterface) Status() (int, string, error) {
	return RoutingTableMatch("Iface", chk.name)
}

/*
#### RoutingTableGateway
Description: Is this the Gateway's IP address, as listed in the routing table?
Parameters:
- IP (IP address)
Example parameters:
- 192.168.0.21, 222.111.0.22
*/

// routeTableGateway checks if an IP address is a Gateway's IP in the
// kernel's IP routing table, as accessed by `route -n`.
type RoutingTableGateway struct{ name string }

func (chk RoutingTableGateway) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	chk.name = params[0]
	return chk, nil
}

func (chk RoutingTableGateway) Status() (int, string, error) {
	return RoutingTableMatch("Gateway", chk.name)
}

// ResponseMatchesGeneral is an abstraction of ResponseMatches and
// ResponseMatchesInsecure that simply varies in the security of the connection
func ResponseMatchesGeneral(urlstr string, re *regexp.Regexp, secure bool) (int, string, error) {
	body := chkutil.URLToBytes(urlstr, secure)
	if re.Match(body) {
		return errutil.Success()
	}
	msg := "Response didn't match regexp"
	return errutil.GenericError(msg, re.String(), []string{string(body)})
}

/*
#### ResponseMatches
Description: Does the response from this URL match this regexp?
Parameters:
- URL (URL string)
- Regexp (regexp)
Example parameters:
- http://my-server.example.com, http://eff.org
- "40[0-9]", "my welome message!", "key:value"
*/

type ResponseMatches struct {
	urlstr string
	re     *regexp.Regexp
}

func (chk ResponseMatches) New(params []string) (chkutil.Check, error) {
	if len(params) != 2 {
		return chk, errutil.ParameterLengthError{2, params}
	}
	// TODO validate URL
	chk.urlstr = params[0]
	re, err := regexp.Compile(params[1])
	if err != nil {
		return chk, errutil.ParameterTypeError{params[1], "regexp"}
	}
	chk.re = re
	return chk, nil
}

func (chk ResponseMatches) Status() (int, string, error) {
	return ResponseMatchesGeneral(chk.urlstr, chk.re, true)
}

/*
#### ResponseMatchesInsecure
Description: Like ResponseMatches, but without SSL certificate validation
*/

type ResponseMatchesInsecure struct {
	urlstr string
	re     *regexp.Regexp
}

func (chk ResponseMatchesInsecure) New(params []string) (chkutil.Check, error) {
	if len(params) != 2 {
		return chk, errutil.ParameterLengthError{2, params}
	}
	// TODO validate URL
	chk.urlstr = params[0]
	re, err := regexp.Compile(params[1])
	if err != nil {
		return chk, errutil.ParameterTypeError{params[1], "regexp"}
	}
	chk.re = re
	return chk, nil
}

func (chk ResponseMatchesInsecure) Status() (int, string, error) {
	return ResponseMatchesGeneral(chk.urlstr, chk.re, false)
}
