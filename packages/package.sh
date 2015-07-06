#!/bin/bash

# This script packages Distributive into an RPM, using a gem-installed FPM for
# package creation. It provides a version number to the binary and optimizes
# for size.

version=0.1.3_dev
# -X sets the value of a string variable in main, others are size optimizations
go build -ldflags "-w -s -O -X main.Version $version" ../
# use FPM to build a package with the correct version number
~/.gem/ruby/2.2.0/bin/fpm -s dir -t rpm -n distributive -f -d go -m langston@aster.is -v $version --epoch 0 ../distributive=/usr/bin/ ../samples=/etc/distributive.d/samples/
