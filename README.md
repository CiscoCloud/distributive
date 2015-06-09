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

## Installation and Usage

To install the development version (potentially unstable):
 1. Clone this repo: `git clone https://github.com/CiscoCloud/distributive`
 2. Build a binary: `cd distributive && go build .`
 3. Run the binary: `./distributive -f ./samples/filesystem.json`

We also provide premade RPM packages on
[Bintray](https://bintray.com/ciscocloud/rpm/Distributive/view#files). The
binary will be installed to `/bin/distributive` and the samples to
`/usr/share/distributive/samples/`. After installing the RPM, usage is simple:
`distributive -f /path/to/check.json`.

## JSON format

General field names
=======

 * `"Name"` : Descriptive name for a check/list (string)
 * `"Notes"` : Human-readable description of this check/list (not used by Distributive).
 * `"Check"` : Type of check to be run (string)
 * `"Parameters"` : Parameters to pass to the check (array of string)

Checks
=======

All check parameters should be formatted as a list of JSON strings.

Filesystem
----------
 * `"file"` : Is there a file at this path?
 * `"directory"` : Is there a directory at this path?
 * `"symlink"` : Is there a symlink at this path?
 * `"checksum"`: Using this algorithm and given this sum, is this file valid (three parameters)?

Packages
--------

 * `"installed"` : Is this program installed on the server?
 * `"ppa"` : Is the PPA at this URL present?
 * `"yumRepo"` : Is the Yum repo with this (short) name configured?
 * `"yumRepoURL"` : Is the Yum repo with this URL configured?

Network
-------

 * `"port"` : Is this port in an open state?
 * `"interface"` : Does this network interface exist?
 * `"up"` : Is this network interface up?
 * `"ip4"` : Does this interface have the specified IP address (two parameters)?
 * `"ip6"` : Does this interface have the specified IP address (two parameters)?
 * `"gateway"` : Does the default gateway have the specified IP address?
 * `"gatewayInterface"` : Is the default gateway operating on this interface?
 * `"TCP"` : Can this host be reached via a TCP connection?
 * `"UDP"` : Can this host be reached via a UDP connection?

Users and Groups
----------------

For all of the following checks, the user can either be specified by their
username or by their UID. Except `"userInGroup"`, which requires a username.

 * `"groupExists"` : Does a group by this name exist on the host?
 * `"groupId"` : Does a group by this name have this group id?
 * `"userInGroup"` : Is this user a member of this group?
 * `"userExists"` : Does this user exist?
 * `"userHasUID"` : Does this user have this UID?
 * `"userHasGID"` : Does this user have this primary GID?
 * `"userHasUsername"` : Does this user have this username?
 * `"userHasName"` : Does this user have this name?
 * `"userHasHomeDir"` : Is this the path of this user's home directory?

System/Misc
-----------

 * `"command"` : Run a shell command.
 * `"running"` : Is this service running on the server?
 * `"temp"` : Does the CPU temp exceed this integer (Celcius)?
 * `"module"` : Is this kernel module activated?
 * `"kernelParameter"` : Is this kernel parameter specified?
 * `"dockerImage"` : Does this Docker image exist on the host?
 * `"dockerRunning"` : Is this Docker container running (must include version,
 e.g. user/container:latest)?

#### Dependencies for certain checks

All of these checks should run on any Linux system (and most on any Unix system)
without any external dependencies other than these:

 * `"temp"` depends on the package lm_sensors.
 * `"installed"` depends on any of the three following package managers: dpkg, rpm, or pacman.
 * `"dockerImage"`, `"dockerRunning"` depend on Docker.

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
