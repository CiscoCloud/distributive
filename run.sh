#!/bin/bash

# This script comiles and runs distributive, mirroring the way it's done in
# package/package.sh. Simply for ease of use.

VERSION=0.1.3_dev
# -X sets the value of a string variable in main, others are size optimizations
go build -ldflags "-w -s -O -X main.Version $VERSION" ./
sudo ./distributive -d "" --verbosity="debug" -f samples/network.json
