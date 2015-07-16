#!/usr/bin/env sh

# This script compiles and runs distributive, downloading any dependencies on
# the fly.

#### GLOBALS

version="0.2"
src="./src/github.com/CiscoCloud/distributive"
bindir="./bin/"

#### GET DEPENDENCIES

# Put them all in ./.godeps
. "./.envrc"
go get ./...

#### BUILD

if [ ! -e "$bindir" ]; then
    mkdir "$bindir"
fi
if [ ! -w "$bindir" ]
then
    echo "I always wished I were a better writer,"
    echo "but I can't even write to $bindir"
    exit 1
fi
if [ ! -r "$src" ]
then
    echo "This code is so bad it's unreadable! But really, can't read $src"
    exit 1
fi

# -X sets the value of a string variable in main, others are size optimizations
go build -ldflags "-w -s -O -X main.Version $version" $src
if [ -e "./distributive" ]; then
    mv ./distributive "$bindir"
else
    echo "I couldn't find that binary at ./distributive..."
    echo "I know I put it around here somewhere..."
    echo 'Oh go "find" it yourself:'
    find . -name "distributive" -type f -executable
fi

#### CLEAN UP

# For some reason, this weird dir gets made...
if [ -e "./bin:/" ]; then
    rm -r "./bin:/"
fi

if [ -d "./pkg/" ]; then
    rm -r "./pkg/"
fi

if [ -f "./src/github.com/CiscoCloud/distributive/distributive" ]; then
    rm "./src/github.com/CiscoCloud/distributive/distributive"
fi
