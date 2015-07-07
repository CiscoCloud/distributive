#!/bin/bash

# This script comiles and runs distributive, mirroring the way it's done in
# package/package.sh. Simply for ease of use.

# Global variables
version=0.1.3_dev

# Install gpm if user doesn't have it
if [ ! -x /usr/local/bin/gpm ]
then
    wget https://raw.githubusercontent.com/pote/gpm/v1.3.2/bin/gpm
    chmod +x gpm
    sudo mv gpm /usr/local/bin
fi

# install depedencies
gpm install

# -X sets the value of a string variable in main, others are size optimizations
go build -ldflags "-w -s -O -X main.Version $version" ./
sudo ./distributive -d samples --verbosity="debug"
