#!/bin/bash

# This script compiles and runs distributive, downloading any dependencies on
# the fly.

#### GLOBALS

version=0.1.3_dev

#### GET DEPENDENCIES

# Install gpm if user doesn't have it
if [ ! -x /usr/local/bin/gpm ]
then
    wget https://raw.githubusercontent.com/pote/gpm/v1.3.2/bin/gpm
    chmod +x gpm
    sudo mv gpm /usr/local/bin
fi

# install depedencies
gpm install

#### BUILD

# -X sets the value of a string variable in main, others are size optimizations
go build -ldflags "-w -s -O -X main.Version $version" ./
