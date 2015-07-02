package workers

import "github.com/CiscoCloud/distributive/wrkutils"

// RegisterAll registers all the checks in this package with wrkutils
func RegisterAll() {
	registerDocker()
	registerFilesystem()
	registerMisc()
	registerUsage()
	registerSystemctl()
	registerPackage()
	registerNetwork()
	registerUsersAndGroups()
}

func registerDocker() {
	wrkutils.RegisterCheck("dockerimage", dockerImage, 1)
	wrkutils.RegisterCheck("dockerrunning", dockerRunning, 1)
	wrkutils.RegisterCheck("dockerrunningapi", dockerRunningAPI, 2)
	wrkutils.RegisterCheck("dockerimageregexp", dockerImageRegexp, 1)
	wrkutils.RegisterCheck("dockerrunningregexp", dockerRunningRegexp, 1)
}

func registerFilesystem() {
	wrkutils.RegisterCheck("file", file, 1)
	wrkutils.RegisterCheck("directory", directory, 1)
	wrkutils.RegisterCheck("symlink", symlink, 1)
	wrkutils.RegisterCheck("checksum", checksum, 3)
	wrkutils.RegisterCheck("permissions", permissions, 2)
	wrkutils.RegisterCheck("filecontains", fileContains, 2)
}

func registerMisc() {
	wrkutils.RegisterCheck("command", command, 1)
	wrkutils.RegisterCheck("commandoutputmatches", commandOutputMatches, 2)
	wrkutils.RegisterCheck("running", running, 1)
	wrkutils.RegisterCheck("phpconfig", phpConfig, 2)
	wrkutils.RegisterCheck("temp", temp, 1)
	wrkutils.RegisterCheck("module", module, 1)
	wrkutils.RegisterCheck("kernelparameter", kernelParameter, 1)
}

func registerNetwork() {
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

func registerPackage() {
	wrkutils.RegisterCheck("installed", installed, 1)
	wrkutils.RegisterCheck("repoexists", repoExists, 2)
	wrkutils.RegisterCheck("repoexistsuri", repoExistsURI, 2)
	wrkutils.RegisterCheck("pacmanignore", pacmanIgnore, 1)
}

func registerSystemctl() {
	wrkutils.RegisterCheck("systemctlloaded", systemctlLoaded, 1)
	wrkutils.RegisterCheck("systemctlactive", systemctlActive, 1)
	wrkutils.RegisterCheck("systemctlsockpath", systemctlSockPath, 1)
	wrkutils.RegisterCheck("systemctlsockunit", systemctlSockUnit, 1)
	wrkutils.RegisterCheck("systemctltimer", systemctlTimer, 1)
	wrkutils.RegisterCheck("systemctltimerloaded", systemctlTimerLoaded, 1)
	wrkutils.RegisterCheck("systemctlunitfilestatus", systemctlUnitFileStatus, 2)
}

func registerUsage() {
	wrkutils.RegisterCheck("diskusage", diskUsage, 2)
	wrkutils.RegisterCheck("memoryusage", memoryUsage, 1)
	wrkutils.RegisterCheck("swapusage", swapUsage, 1)
	wrkutils.RegisterCheck("cpuusage", cpuUsage, 1)
	wrkutils.RegisterCheck("freememory", freeMemory, 1)
}

func registerUsersAndGroups() {
	wrkutils.RegisterCheck("groupexists", groupExists, 1)
	wrkutils.RegisterCheck("useringroup", userInGroup, 2)
	wrkutils.RegisterCheck("groupid", groupID, 2)
	wrkutils.RegisterCheck("userexists", userExists, 1)
	wrkutils.RegisterCheck("userhasuid", userHasUID, 2)
	wrkutils.RegisterCheck("userhasgid", userHasGID, 2)
	wrkutils.RegisterCheck("userhasusername", userHasUsername, 2)
	wrkutils.RegisterCheck("userhasname", userHasName, 2)
	wrkutils.RegisterCheck("userhashomedir", userHasHomeDir, 2)
}
