#!/bin/bash

# This script packages Distributive into an RPM, using a gem-installed FPM for
# package creation. It provides a version number to the binary and optimizes
# for size.

#### GLOBALS

bin_location=/usr/bin/
sample_location=/etc/distributive.d/samples/

#### BUILD

cd ..
source build.sh
cd -

#### PACKAGE

# use FPM to build a package with the correct version number
PATH=$PATH:~/.gem/ruby/2.2.0/bin/
fpm -s dir -t rpm -n distributive -f -d go -m langston@aster.is -v $version --epoch 0 ../distributive=$bin_location ../samples=$sample_location
