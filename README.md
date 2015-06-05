# Overview

<!-- markdown-toc start - Don't edit this section. Run M-x markdown-toc/generate-toc again -->
**Table of Contents**

- [Overview](#overview)
    - [Usage](#usage)
    - [JSON format](#json-format)
        - [Check names](#check-names)
        - [Dependencies for certain checks](#dependencies-for-certain-checks)
    - [Comparison to Other Software](#comparison-to-other-software)
        - [Serverspec](#serverspec)
        - [Nagios](#nagios)
    - [Supported Frameworks](#supported-frameworks)

<!-- markdown-toc end -->


Distributive is a tool for running distributed health checks in server clusters.
It was designed with Consul in mind, but is platform agnostic.  The architecture
is such that some external server will ask the host to execute this program,
reading in from some JSON file, and will record this
program's exit code and standard out.

The exit code meanings are defined as [Consul] [1] and [Sensu] [2] recognize
them.

 * Exit code 0 - Checklist is passing
 * Exit code 1 - Checklist is warning
 * Any other code - Checklist is failing

As of right now, only exit codes 0 and 1 are used, even if a checklist fails.

## Usage

Build a binary with `go build .` and run it with the command line flag `f`
followed by an argument pointing to the json file containing the checklist you
wish to run.
```
distributive -f ./samples/file-checks.json
```

## JSON format

General field names
=======

 * `"Name"` : Descriptive name for a check/list (string)
 * `"Notes"` : Human-readable description of this check/list (not used by Distributive).
 * `"Check"` : Type of check to be run (string)
 * `"Parameters"` : Parameters to pass to the check (array of string)

Check names
=======

 * `"command"` : Run a shell command.
 * `"running"` : Is this service running on the server?
 * `"file"` : Is there a file at this path?
 * `"directory"` : Is there a directory at this path?
 * `"symlink"` : Is there a symlink at this path?
 * `"checksum"`: Using this algorithm and given this sum, is this file valid (three parameters)?
 * `"installed"` : Is this program installed on the server?
 * `"temp"` : Does the CPU temp exceed this integer (Celcius)?
 * `"port"` : Is this port in an open state?
 * `"interface"` : Does this network interface exist?
 * `"up"` : Is this network interface up?
 * `"ip4"` : Does this interface have the specified IP address (two parameters)?
 * `"ip6"` : Does this interface have the specified IP address (two parameters)?
 * `"gateway"` : Does the default gateway have the specified IP address?
 * `"gatewayInterface"` : Is the default gateway operating on this interface?
 * `"module"` : Is this kernel module activated?
 * `"kernelParameter"` : Is this kernel parameter specified?
 * `"dockerImage" : Does this Docker image exist on the host?
 * `"dockerRunning" : Is this Docker container running (must include version,
 e.g. user/container:latest)?

#### Dependencies for certain checks

All dependencies should be installed on Linux systems by default, except for
lm_sensors.
 * `"running"` depends on the ability to execute `ps aux`.
 * `"temp"` depends on the package lm_sensors.
 * `"installed"` depends on any of the three following package managers: dpkg, rpm, or pacman.
 * `"port"` reads from `/proc/net/tcp`, and depends on its proper population for accuracy.
 * `"module"` reads from the output of `/sbin/lsmod`.
 * `"kernelParameter"` reads from the output of `/sbin/sysctl -q -n`.

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
designed for dynamic systems with changing IPs, which can report into Consul,
Sensu, or another framework as soon as they are ready, and require little or no
centralized configuration. Additionally, Distributive attempts to rely as little
as possible on external tools/commands, using mostly just the Go standard library.

### Nagios

Nagios is an end-to-end monitoring, security, and notification framework. It is
designed around the central control server approach to monitoring. Nagios provides
many other services not included in Distributive, but may not be suitable for
some projects due to its size, complexity, and centralized architecture.
Distributive is simple, lightweight, and easy to configure, and doesn't provide
its own scheduling, dashboard, etc. It is designed to be used within frameworks
such as Sensu and Consul.

## Supported Frameworks

Distributive attempts to be as framework-agnostic as possible. It is known to
work well with both Sensu and Consul, which have similar architecture with
regards to their health checks.

[1]: https://www.consul.io/docs/agent/checks.html "Consul"
[2]: https://sensuapp.org/docs/0.18/checks "Sensu"
