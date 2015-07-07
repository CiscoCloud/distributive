#!/bin/bash
source build.sh
sudo ./distributive -d "" -f samples/systemctl-fail.json --verbosity="debug"
