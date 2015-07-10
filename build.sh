#!/bin/bash

# This script compiles and runs distributive, downloading any dependencies on
# the fly.

#### GLOBALS

version=0.2
src=./src/github.com/CiscoCloud/distributive

#### GET DEPENDENCIES

function download_to_dir() {
    dest=$1
    url=$2
    filename=`basename $url`

    if [ ! -w $dest ]
    then
        mkdir $dest
    fi

    curl -O $url
    chmod +x $filename
    if [ ! -w $dest ]
    then
        sudo mv $filename $dest
    else
        mv $filename $dest
    fi
}

# Install gpm if the user doesn't have it
gpm_dir=./.gpm
gpm=$gpm_dir/gpm
if [ ! -x $gpm ]
then
    download_to_dir $gpm_dir https://raw.githubusercontent.com/pote/gpm/v1.3.2/bin/gpm
fi

# Install gvp if the user doesn't have it
gvp_dir=./.gvp
gvp=$gvp_dir/gvp
if [ ! -x $gvp ]
then
    mkdir $gvp_dir
    download_to_dir $gvp_dir https://raw.githubusercontent.com/pote/gvp/v0.2.0/bin/gvp
fi

# install depedencies
source $gvp
$gpm install

#### BUILD

# -X sets the value of a string variable in main, others are size optimizations
bindir=./bin
mkdir -p ./bin 2&> /dev/null
if [ ! -w ./bin ]
then
    echo "Can't write to ./bin, please change its permissions"
    exit 1
fi
if [ ! -d ./bin ]
then
    mkdir ./bin
fi
if [ ! -r src ]
then
    echo "Couldn't read source files in $src"
    exit 1
fi
go build -ldflags "-w -s -O -X main.Version $version" $src
mv ./distributive ./bin
