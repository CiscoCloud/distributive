#!/bin/bash

# This script packages Distributive into an RPM, using a gem-installed FPM for
# package creation. It provides a version number to the binary and optimizes
# for size.

#### GLOBALS

binary_dir=/usr/bin/
sample_dir=/etc/distributive.d/samples/
binary=../bin/distributive
samples=../samples

#### BUILD

cd ..
source build.sh
cd -

#### PACKAGE

gem_dir="~/.gem/ruby/2.2.0/bin/"
PATH=$PATH:$gem_dir
fpm -s dir -t rpm -n distributive -f -d go -m langston@aster.is -v $version --epoch 0 $binary=$binary_dir $samples=$sample_dir
