[![Build Status](https://travis-ci.org/CiscoCloud/distributive.svg?branch=master)](https://travis-ci.org/CiscoCloud/distributive)

<!-- markdown-toc start - Don't edit this section. Run M-x markdown-toc-generate-toc again -->
**Table of Contents**

- [Overview](#overview)
- [Installation and Usage](#installation-and-usage)
    - [Installation/Building](#installationbuilding)
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

This readme documents the current (development) version of distributive.

Distributive is a tool for running distributed health checks in datacenters.
It was designed with Consul in mind, but is stack agnostic. It is simple
to configure (with YAML checklists) and easy to deploy and run. It has no
runtime dependencies, and can be shipped as a single binary.

Usually, some external server will ask the host to execute this program, reading
a checklist from a YAML file, and will record this program's exit code and
standard out. Distributive's output includes information about which checks
in a checklist failed, and how so.

The exit code meanings are defined as [Consul][consul], [Kubernetes][kubernetes],
[Sensu][sensu], and [Nagios][nagios] recognize them.

 * Exit code 0 - Checklist is passing
 * Exit code 1 - Checklist is warning
 * Any other code - Checklist is failing

As of right now, only exit codes 0 and 1 are used, even if a checklist fails.

Installation and Usage
======================

Installation/Building
---------------------

To install the development version (potentially unstable):
 1. Clone this repo: `git clone https://github.com/CiscoCloud/distributive && cd distributive`
 2. Get [Glide][glide].
 3. Install dependencies with `glide install`
 4. (Optional) Test with `go test $(glide novendor)`
 5. Build a binary with `go build .`
 6. Follow the "Usage" instructions below

Distributive currently only supports Linux.

Usage
-----

The default behavior is to run any checks specified via `--file`, `--url`,
`--stdin`, or `--directory` options, or all checks in /etc/distributive.d/ if no
location is specified.

```
$ distributive --help
[...]
GLOBAL OPTIONS:
   --verbosity          info | debug | fatal | error | panic | warn
   --file, -f           Read a checklist from a file
   --url, -u            Read a checklist from a URL
   --directory, -d      Read all of the checklists in this directory
   --stdin, -s          Read data piped from stdin as a checklist
   --no-cache           Don't use a cached version of a remote check, fetch it.
   --help, -h           show help
   --version, -v        print the version
```

Examples:

```
$ /path/to/distributive --verbosity="warn" -f ./samples/filesystem.yml
$ distributive --f="/etc/distributive/samples/network.yaml" --verbosity=debug
$ ./distributive -u "http://pastebin.com/raw.php?i=5c1BAxcX"
$ /distributive --verbosity="info"
$ /path/to/distributive -d "/etc/distributive.d/" # same as default behavior
$ cat samples/filesystem.yml | ./distributive -d "" -s=true --verbosity=fatal
```

Supported Frameworks
--------------------

Distributive attempts to be as framework-agnostic as possible. It is known to
work well with Consul, Kubernetes, Sensu, and Nagios, which have similar design
in how they detect passing and failing checks. There is documentation on how to
use Distributive with Consul on our [Github wiki][wiki].

Checks
=======

For the impatient, examples of every single implemented check are available in
the `samples/` directory, sorted by category. There is extensive documentation
for each check available on our [Github wiki][wiki].

If you'd like to see how Distributive is used in production environments, take
a look at the [RPM source][mantl-packaging], which includes checks used in
[Mantl][mantl].


Dependencies
============

Distributive itself has no dependencies; it is a standalone binary. Some checks,
however, rely on output from specific commands. These dependencies are outlined
for each check on our [Github wiki][wiki].

Comparison to Other Software
============================

Distributive was created with the idea of pushing responsibility to the nodes,
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
[Nagios exit code specifications] [nagios], and can be used just like any other
plugin. Its advantage over other plugins is that it is small, fast, and has no
dependencies.

Contributing and Getting Help
=============================

Contributing
------------

Thank you for your interest in contributing! To get started, please check out
[our wiki][wiki].

Getting Help
------------

All comments, questions, and contributions are always welcome. We strive to
provide expedient and detailed support for anyone using our software. Please
submit any requests via our [Github Issues Page][issues], where someone will
see it and get to work promptly.

License
=======
Copyright © 2015 Cisco Systems, Inc.

Licensed under the [Apache License, Version 2.0][license] (the "License").

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.


[license]: http://www.apache.org/licenses/LICENSE-2.0
[wiki]: https://github.com/CiscoCloud/distributive/wiki
[issues]: https://github.com/CiscoCloud/distributive/issues
[consul]: https://www.consul.io/docs/agent/checks.html
[sensu]: https://sensuapp.org/docs/0.18/checks
[nagios]: https://nagios-plugins.org/doc/guidelines.html#AEN78
[kubernetes]: http://kubernetes.io/v1.0/docs/user-guide/walkthrough/k8s201.html#health-checking
[mantl]: https://github.com/CiscoCloud/mantl
[glide]: https://github.com/Masterminds/glide
[mantl-packaging]: https://github.com/asteris-llc/mantl-packaging/tree/master/distributive
