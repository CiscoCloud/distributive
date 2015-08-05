#!/usr/bin/env sh

# This script compiles and runs distributive, downloading any dependencies on
# the fly.

#### GLOBALS

version="0.2.1"
src="./src/github.com/CiscoCloud/distributive"
bindir="./bin/"

# for POSIX shell compliance: http://www.etalabs.net/sh_tricks.html
echo () { printf %s\\n "$*" ; }

# Description: echoes the contents of the parent dir of $1, then exits 1
error_list_parent() {
    echo "Well, if you think you could do a better job, then here!"
    parent=$(dirname "$1")
    ls -alh "$parent"
    exit 1
}

# Description: exit 1 with a message about not being about to read
cant_read_error() {
    echo "This code is so bad it's unreadable! But really, can't read $1"
    error_list_parent "$1"
}

# Description: Echo a can't write error and exit 1
# Arguments: $1 - dir
cant_write_error() {
    echo "I always wished I were a better writer, but I can't even write to $1"
    error_list_parent "$1"
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

# Description: exit 1 with a message if you can't read the resource
assert_readable() { [ ! -r "$1" ] && cant_read_error "$1" ; }

# Description: Check if an executable file is on the PATH (for use in if)
# Arguments: $1 - Name of executable
# Returns: 0 - executable is on PATH, 1 - " " not " "
executable_exists() {
    if command -v "$1" >/dev/null 2>&1; then # command exists
        return 0
    fi
    return 1
}

# Description: Get go package/depedencies, put them in ./.godeps
godep() {
  export GOPATH="$PWD/.godeps"
  export GOBIN="$PWD/.godeps/bin"
  go get "$1" 2> /dev/null
  export GOPATH="$PWD:$PWD/.godeps"
  export GOBIN="$PWD/bin:$PWD/.godeps/bin"
}

#### GET DEPENDENCIES

assert_writable "d" "$GOPATH"
assert_writable "d" "$GOBIN"
assert_readable "$src"
godep ./...

#### BUILD

assert_writable "d" "$bindir"
assert_readable "$src"

# -X sets the value of a string variable in main, others are size optimizations
go build -ldflags "-w -s -O -X main.Version $version" "$src"
executable="./distributive"
[ -e "$executable" ] && mv "$executable" "$bindir"

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

# Description: If $1 exists, rm -r it!
remove_if_exists() { [ -e "$1" ] && rm -r "$1" ; }

# For some reason, this weird dir gets made...
remove_if_exists "./bin:"
remove_if_exists "../distributive:"
remove_if_exists "./pkg/"
remove_if_exists "./src/github.com/CiscoCloud/distributive/distributive"
