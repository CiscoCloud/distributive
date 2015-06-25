<!-- markdown-toc start - Don't edit this section. Run M-x markdown-toc/generate-toc again -->
**Table of Contents**

- [Overview](#overview)
- [Installation and Usage](#installation-and-usage)
    - [Installation](#installation)
    - [Usage](#usage)
    - [Supported Frameworks](#supported-frameworks)
- [Checks](#checks)
- [Dependencies](#dependencies)
- [Comparison to Other Software](#comparison-to-other-software)
    - [Serverspec](#serverspec)
    - [Nagios](#nagios)
- [Contributing and Getting Help](#contributing-and-getting-help)
    - [Contributing](#contributing)
    - [Getting Help](#getting-help)
- [License](#license)

<!-- markdown-toc end -->

Overview
========

[![GoDoc](https://img.shields.io/badge/api-Godoc-blue.svg?style=flat-square)](https://godoc.org/github.com/CiscoCloud/distributive/tabular)

This readme documents the current (development) version of distributive.

Distributive is a tool for running distributed health checks in datacenters.
It was designed with Consul in mind, but is platform agnostic. It is simple
to configure (with JSON checklists) and easy to deploy and run. It has no
dependencies, and can be shipped as a speedy 7MB (yes, megabytes!) binary.

Usually, some external server will ask the host to execute this program, reading
a checklist from a JSON file, and will record this program's exit code and
standard out. Distributive's output includes information about which checks
in a checklist failed, and how so.

The exit code meanings are defined as [Consul] [1], [Sensu] [2], and
[Nagios] [3] recognize them.

 * Exit code 0 - Checklist is passing
 * Exit code 1 - Checklist is warning
 * Any other code - Checklist is failing

As of right now, only exit codes 0 and 1 are used, even if a checklist fails.

Installation and Usage
======================

Installation
------------
To install the development version (potentially unstable):
 1. Clone this repo: `git clone https://github.com/CiscoCloud/distributive`
 2. Build a binary: `cd distributive && go build .`
 3. Run the binary (as outlined in "Usage") with `./distributive`.

We also provide premade RPM packages on
[Bintray](https://bintray.com/ciscocloud/rpm/Distributive/view#files). The
binary will be installed to `/bin/distributive` and the samples to
`/etc/distributive/samples/`.

Usage
-----

The default behavior is to run all checks in /etc/distributive.d/ (the default
directory give to the `-a` option), in addition to any specified via the `-f`
or `-u` options.

```
$ distributive --help
Usage of ./distributive:
  -a="/etc/distributive.d/": Run all the checks in the specified directory
  -f="": Use the health check located at this path
  -log-level="warn": Output verbosity level (valid values are info | debug | fatal | error | panic | warn
  -u="": Use the health check located at this URL
  -version=false: Get the version of distributive this binary was built from
```

Examples:

```
$ /path/to/distributive -v=2 -f ./samples/filesystem.json
$ distributive -v=0 -a="" --f="/etc/distributive/samples/network.json"
$ ./distributive -a="" -u "http://pastebin.com/raw.php?i=5c1BAxcX"
$ /distributive
$ /path/to/distributive -a "/etc/distributive.d/"
```

Supported Frameworks
--------------------

Distributive attempts to be as framework-agnostic as possible. It is known to
work well with Sensu, Consul, and Nagios, which have similar design in how they
detect passing and failing checks. There is documentation on how to use
Distributive with Consul on this project's
[Github wiki](https://github.com/CiscoCloud/distributive/wiki/Working-with-Consul).


Checks
=======

For the impatient, examples of every single implemented check are available in
the `samples/` directory, sorted by category. There is extensive documentation
for each check available on this project's
[Github wiki](https://github.com/CiscoCloud/distributive/wiki).


Dependencies
============

Distributive itself has no dependencies; it is compiled as a standalone Go
binary. Some checks, however, rely on output from specific packages. These
dependencies are outlined for each check on this project's
[Github wiki](https://github.com/CiscoCloud/distributive/wiki/Checks-and-Checklists).

Comparison to Other Software
============================

Distributive was created with the idea of pushing responsibiliy to the nodes,
It was also designed around the idea of constantly changing infrastructure, with
servers being added and destroyed constantly, changing IP addresses, and even
changing roles. Integration with Consul provides even
[greater flexibility](https://www.consul.io/intro/vs/nagios-sensu.html).

Serverspec
----------

Serverspec runs on a single control server, and requires each check to be in a
directory matching the hostname of the machine to run it on. Distributive was
designed for dynamic systems with changing IPs, which can report into Consul,
Sensu, or another framework as soon as they are ready, and require little or no
centralized configuration. Additionally, Distributive attempts to rely as little
as possible on external tools/commands, using mostly just the Go standard library.

Nagios
------

Nagios is an end-to-end monitoring, security, and notification framework. It
provides many services not included in Distributive, and solves a very different
problem.  Distributive is simple, lightweight, and easy to configure, and
doesn't provide its own scheduling, dashboard, etc. It is designed to be used
within frameworks such as Sensu and Consul. Luckily, Distributive conforms to
[Nagios exit code specifications] [3], and can be used just like any other
plugin. Its advantage over other plugins is that it is small, fast, and has no
dependencies.

Contributing and Getting Help
=============================

Contributing
------------

Thank you for your interest in contributing! To get started, please check out
[this page on our wiki](https://github.com/CiscoCloud/distributive/wiki/How-It-Works-%28and-So-Can-You!%29).

Getting Help
------------

Feature requests, documentation requests, help installing and using, pull
requests, and other comments or questions are all always welcome. We strive to
provide expedient and detailed support for anyone using our software. Please
submit any requests via our
[Github Issues Page](https://github.com/CiscoCloud/distributive/issues),
where someone will see it and get to work promptly.

License
=======
Copyright © 2015 Cisco Systems, Inc.

Licensed under the [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0) (the "License").

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.


[1]: https://www.consul.io/docs/agent/checks.html "Consul"
[2]: https://sensuapp.org/docs/0.18/checks "Sensu"
[3]: https://nagios-plugins.org/doc/guidelines.html#AEN78 "Nagios"
