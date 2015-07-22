#!/usr/bin/env sh

# This script compiles and runs distributive, downloading any dependencies on
# the fly.

#### GLOBALS

version="0.2.1"
src="./src/github.com/CiscoCloud/distributive"
bindir="./bin/"

# for POSIX shell compliance: http://www.etalabs.net/sh_tricks.html
echo () { printf %s\\n "$*" ; }

# Description: Echo a can't write error and exit 1
# Arguments: $1 - dir
cant_write_error() {
    echo "I always wished I were a better writer,"
    echo "but I can't even write to $1"
    exit 1
}

# Description: If the resource doesn't exist, attempt to create it.
# Arguments: $1 - type ("f" | "d"), $2 - path
assert_writable() {
    if [ "$1" = "f" ]; then
        [ ! -e "$1" ] && touch "$2"
    elif [ "$1" = "d" ]; then
        [ ! -e "$1" ] && mkdir -p "$2"
    else
        echo "Internal error: invalid argument to assert_exists: $1"
        exit 1
    fi
    [ ! -w "$2" ] && cant_write_error "$2"
}

# Description: returns (echoes) all stderr and stdout from a command
capture_output() {
    temp_file="./.tmp"
    assert_writable "f" "$temp_file"
    # Redirect stdout, stderr to $temp_file, read, remove (POSIX compliant)
    $1 >$temp_file 2>&1
    output=$(<$temp_file)
    rm $temp_file
    echo "$output"
}

# Description: Check if an executable file is on the PATH (for use in if)
# Arguments: $1 - Name of executable
# Returns: 0 - executable is on PATH, 1 - " " not " "
executable_exists() {
    if command -v "$1" >/dev/null 2>&1; then # command exists
        return 0
    fi
    return 1
}

#### GET DEPENDENCIES

# Put them all in ./.godeps
GOPATH="$PWD/.godeps/"
GOBIN="$PWD/.godeps/bin"
assert_writable "d" "$GOPATH"
assert_writable "d" "$GOBIN"
get_output=$(capture_output "go get ./...")

# Restore env variables to what they should be to build (include ./src)
GOPATH=$GOPATH:$PWD
GOBIN="$GOPATH/bin:$PWD/bin"
. "./.envrc"

#### BUILD

assert_writable "d" "$bindir"
if [ ! -r "$src" ]; then
    echo "This code is so bad it's unreadable! But really, can't read $src"
    exit 1
fi

# -X sets the value of a string variable in main, others are size optimizations
build_output=$(go build -ldflags "-w -s -O -X main.Version $version" $src)
executable="./distributive"
[ -e "$executable" ] && mv "$executable" "$bindir"

#### REPORT BUILD ERRORS

if [ ! -f "$bindir/$executable" ]; then
    echo "Looks like the build failed. Here's the output of go get ./..."
    echo "$get_output"
    echo "And here it is for the build:"
    echo "$build_output"
fi

#### COMPRESS WITH UPX

if [ "$1" = "compress" ]  ; then
    if [ ! -f "$bindir/$executable" ]; then
        echo "Couldn't find executable to compress at $bindir/$executable"
        exit 1
    fi
    if executable_exists "upx" && executable_exists "goupx"; then
        goupx --no-upx "$bindir/$executable"
        upx --color --ultra-brute "$bindir/$executable"
    else
        echo "Couldn't find either UPX or goupx"
        exit 1
    fi
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
