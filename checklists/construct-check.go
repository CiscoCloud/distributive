package checklists

import (
	"github.com/CiscoCloud/distributive/chkutil"
	"github.com/CiscoCloud/distributive/workers"
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
		return workers.DockerImage{}
	case "dockerimageregexp":
		return workers.DockerImageRegexp{}
	case "dockerrunning":
		return workers.DockerRunning{}
	case "dockerrunningapi":
		return workers.DockerRunningAPI{}
	case "dockerrunningregexp":
		return workers.DockerRunningRegexp{}
		/***************** filesystem.go *****************/
	case "file":
		return workers.File{}
	case "directory":
		return workers.Directory{}
	case "symlink":
		return workers.Symlink{}
	case "checksum":
		return workers.Checksum{}
	case "filematches":
		return workers.FileMatches{}
	case "permissions":
		return workers.Permissions{}
		/***************** misc.go *****************/
	case "command":
		return workers.Command{}
	case "commandoutputmatches":
		return workers.CommandOutputMatches{}
	case "running":
		return workers.Running{}
	case "temp":
		return workers.Temp{}
	case "module":
		return workers.Module{}
	case "kernelparameter":
		return workers.KernelParameter{}
	case "phpconfig":
		return workers.PHPConfig{}
		/***************** network.go *****************/
	case "port":
		return workers.Port{}
	case "interfaceexists":
		return workers.InterfaceExists{}
	case "up":
		return workers.Up{}
	case "ip4":
		return workers.IP4{}
	case "ip6":
		return workers.IP6{}
	case "gateway":
		return workers.Gateway{}
	case "gatewayinterface":
		return workers.GatewayInterface{}
	case "host":
		return workers.Host{}
	case "tcp":
		return workers.TCP{}
	case "udp":
		return workers.UDP{}
	case "tcptimeout":
		return workers.TCPTimeout{}
	case "udptimeout":
		return workers.UDPTimeout{}
	case "routingtabledestination":
		return workers.RoutingTableDestination{}
	case "routingtableinterface":
		return workers.RoutingTableInterface{}
	case "routingtablegateway":
		return workers.RoutingTableGateway{}
	case "responsematches":
		return workers.ResponseMatches{}
	case "responsematchesinsecure":
		return workers.ResponseMatchesInsecure{}
		/***************** packages.go *****************/
	case "repoexists":
		return workers.RepoExists{}
	case "repoexistsuri":
		return workers.RepoExistsURI{}
	case "pacmanignore":
		return workers.PacmanIgnore{}
	case "installed":
		return workers.Installed{}
		/***************** systemctl.go *****************/
	case "systemctlloaded":
		return workers.SystemctlLoaded{}
	case "systemctlactive":
		return workers.SystemctlActive{}
	case "systemctlsocklistening":
		return workers.SystemctlSockListening{}
	case "systemctlsockunit":
		return workers.SystemctlSockUnit{}
	case "systemctltimer":
		return workers.SystemctlTimer{}
	case "systemctltimerloaded":
		return workers.SystemctlTimerLoaded{}
		/***************** usage.go *****************/
	case "memoryusage":
		return workers.MemoryUsage{}
	case "swapusage":
		return workers.SwapUsage{}
	case "freememory":
		return workers.FreeMemory{}
	case "freeswap":
		return workers.FreeSwap{}
	case "cpuusage":
		return workers.CPUUsage{}
	case "diskusage":
		return workers.DiskUsage{}
		/***************** users-and-groups.go *****************/
	case "groupexists":
		return workers.GroupExists{}
	case "useringroup":
		return workers.UserInGroup{}
	case "groupid":
		return workers.GroupID{}
	case "userexists":
		return workers.UserExists{}
	case "userhasuid":
		return workers.UserHasUID{}
	case "userhasgid":
		return workers.UserHasGID{}
	case "userhasusername":
		return workers.UserHasUsername{}
	case "userhashomedir":
		return workers.UserHasHomeDir{}
	/***************** default *****************/
	default:
		log.WithFields(log.Fields{
			"id": chkjs.ID,
		}).Fatalf("Invalid check ID")
	}
	return nil
}
