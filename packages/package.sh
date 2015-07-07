#!/bin/bash

# This script packages Distributive into an RPM, using a gem-installed FPM for
# package creation. It provides a version number to the binary and optimizes
# for size.

# Global variables
version=0.1.3_dev
bin_location=/usr/bin/
sample_location=/etc/distributive.d/samples/

# Install gpm if user doesn't have it
if [ ! -x /usr/local/bin/gpm ]
then
    wget https://raw.githubusercontent.com/pote/gpm/v1.3.2/bin/gpm
    chmod +x gpm
    sudo mv gpm /usr/local/bin
fi

# install depedencies
cd "$(dirname "$0")" # workdir = dir of this script
cd ..
gpm install ..
cd -

# -X sets the value of a string variable in main, others are size optimizations
go build -ldflags "-w -s -O -X main.Version $version" ../
# use FPM to build a package with the correct version number
PATH=$PATH:~/.gem/ruby/2.2.0/bin/
fpm -s dir -t rpm -n distributive -f -d go -m langston@aster.is -v $version --epoch 0 ../distributive=$bin_location ../samples=$sample_location
