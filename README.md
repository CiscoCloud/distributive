# Overview

Distributive is a tool for running distributed health checks in server clusters.
It was designed with Consul in mind, but is platform agnostic.
The architecture is such that some external server will ask the host to
execute this program, reading in from some JSON file, and will record this
program's exit code and standard out.

The exit code meanings are defined as [Consul recognizes them] [1]
 * Exit code 0 - Check is passing
 * Exit code 1 - Check is warning
 * Any other code - Check is failing
As of right now, only exit codes 0 and 1 are used, even if a check fails.

## Usage

Run the binary with the command line flag `f` and an argument pointing to the
json file containing the check you wish to run.
```
distributive -f ./health-checks/sleep.json
```

## JSON format
General field names:
 * "Name" : Descriptive name for this check.
 * "Notes" : Human-readable description of this service (not used by Distributive).

Specific checks:
 * "Command" : Run a shell command.
 * "Running" : Is this service running on the server?
 * "Exists" : Does this file exist on the server?
 * "Installed" : Is this program installed on the server?
 * "Temp" : Does the CPU temp exceed this integer (Celcius)?

Dependencies for certain checks:
 * "Running" depends on the ability to execute `ps aux`.
 * "Temp" depends on the package lm_sensors.
 * "Installed" depends on any of the three following package managers: dpkg, rpm, or pacman.
 * "Port" reads from `/proc/net/tcp`, and depends on its proper population for accuracy.

## Comparison to Other Software

Distributive was created with the idea of pushing responsibiliy to the nodes,
which grants the program a certain flexibility in what kind of checks it can run.
It has access to local data that cannot or should not be accessed over a network,
by another server. It was also designed around the idea of constantly changing
infrastructure, with servers being added and destroyed constantly, and changing
IP addresses.

### Serverspec

Serverspec runs on a single control server, and requires each check to be in a
directory matching the hostname of the machine to run it on. Distributive was
designed for dynamic systems with changing IPs, which can report into Consul or
another framework as soon as they are ready, and require little or no centralized
configuration.

### Nagios

Nagios is an end-to-end monitoring, security, and notification framework. It is
designed around the central control server approach to monitoring. Nagios provides
many other services not included in Distributive, but may not be suitable for
some projects due to its size, complexity, and centralized architecture.
Distributive is simple, lightweight, and easy to configure.


[1]: https://www.consul.io/docs/agent/checks.html "Consul"
