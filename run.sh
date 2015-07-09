#!/bin/bash
source build.sh
sudo ./bin/distributive -d "" -f samples/systemctl-fail.json --verbosity="debug"
