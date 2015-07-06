#!/bin/bash

# This script comiles and runs distributive, mirroring the way it's done in
# package/package.sh. Simply for ease of use.

version=0.1.3_dev
# -X sets the value of a string variable in main, others are size optimizations
go build -ldflags "-w -s -O -X main.Version $version" ./
sudo ./distributive -d "" -u "http://pastebin.com/raw.php?i=L0FhxKpG" --verbosity="debug"
