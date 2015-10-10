package checklists

import (
	"github.com/CiscoCloud/distributive/checks"
	"github.com/CiscoCloud/distributive/chkutil"
	log "github.com/Sirupsen/logrus"
	"strings"
)

// constructCheck returns a new Check interface compliant object, translated
// from JSON and assigned its parameters
// TODO think about origin tracing - even by line in a checklist
func constructCheck(chkjs CheckJSON) chkutil.Check {
	switch strings.ToLower(chkjs.ID) {
	/***************** docker.go *****************/
	case "dockerimage":
		return checks.DockerImage{}
	case "dockerimageregexp":
		return checks.DockerImageRegexp{}
	case "dockerrunning":
		return checks.DockerRunning{}
	case "dockerrunningapi":
		return checks.DockerRunningAPI{}
	case "dockerrunningregexp":
		return checks.DockerRunningRegexp{}
		/***************** filesystem.go *****************/
	case "file":
		return checks.File{}
	case "directory":
		return checks.Directory{}
	case "symlink":
		return checks.Symlink{}
	case "checksum":
		return checks.Checksum{}
	case "filematches":
		return checks.FileMatches{}
	case "permissions":
		return checks.Permissions{}
		/***************** misc.go *****************/
	case "command":
		return checks.Command{}
	case "commandoutputmatches":
		return checks.CommandOutputMatches{}
	case "running":
		return checks.Running{}
	case "temp":
		return checks.Temp{}
	case "module":
		return checks.Module{}
	case "kernelparameter":
		return checks.KernelParameter{}
	case "phpconfig":
		return checks.PHPConfig{}
		/***************** network.go *****************/
	case "port":
		return checks.Port{}
	case "porttcp":
		return checks.PortTCP{}
	case "portudp":
		return checks.PortUDP{}
	case "interfaceexists":
		return checks.InterfaceExists{}
	case "up":
		return checks.Up{}
	case "ip4":
		return checks.IP4{}
	case "ip6":
		return checks.IP6{}
	case "gateway":
		return checks.Gateway{}
	case "gatewayinterface":
		return checks.GatewayInterface{}
	case "host":
		return checks.Host{}
	case "tcp":
		return checks.TCP{}
	case "udp":
		return checks.UDP{}
	case "tcptimeout":
		return checks.TCPTimeout{}
	case "udptimeout":
		return checks.UDPTimeout{}
	case "routingtabledestination":
		return checks.RoutingTableDestination{}
	case "routingtableinterface":
		return checks.RoutingTableInterface{}
	case "routingtablegateway":
		return checks.RoutingTableGateway{}
	case "responsematches":
		return checks.ResponseMatches{}
	case "responsematchesinsecure":
		return checks.ResponseMatchesInsecure{}
		/***************** packages.go *****************/
	case "repoexists":
		return checks.RepoExists{}
	case "repoexistsuri":
		return checks.RepoExistsURI{}
	case "pacmanignore":
		return checks.PacmanIgnore{}
	case "installed":
		return checks.Installed{}
		/***************** systemctl.go *****************/
	case "systemctlloaded":
		return checks.SystemctlLoaded{}
	case "systemctlactive":
		return checks.SystemctlActive{}
	case "systemctlsocklistening":
		return checks.SystemctlSockListening{}
	case "systemctltimer":
		return checks.SystemctlTimer{}
	case "systemctltimerloaded":
		return checks.SystemctlTimerLoaded{}
		/***************** usage.go *****************/
	case "memoryusage":
		return checks.MemoryUsage{}
	case "swapusage":
		return checks.SwapUsage{}
	case "freememory":
		return checks.FreeMemory{}
	case "freeswap":
		return checks.FreeSwap{}
	case "cpuusage":
		return checks.CPUUsage{}
	case "diskusage":
		return checks.DiskUsage{}
		/***************** users-and-groups.go *****************/
	case "groupexists":
		return checks.GroupExists{}
	case "useringroup":
		return checks.UserInGroup{}
	case "groupid":
		return checks.GroupID{}
	case "userexists":
		return checks.UserExists{}
	case "userhasuid":
		return checks.UserHasUID{}
	case "userhasgid":
		return checks.UserHasGID{}
	case "userhasusername":
		return checks.UserHasUsername{}
	case "userhashomedir":
		return checks.UserHasHomeDir{}
	/***************** default *****************/
	default:
		log.WithFields(log.Fields{
			"id": chkjs.ID,
		}).Fatalf("Invalid check ID")
	}
	return nil
}
