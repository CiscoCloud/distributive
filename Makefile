TEST?=./...
NAME = $(shell awk -F\" '/^const Name/ { print $$2 }' main.go)
VERSION = $(shell awk -F\" '/^const Version/ { print $$2 }' main.go)

all: deps build

deps:
	go get -d -v ./...

updatedeps:
	go get -u -v ./...

build: deps
	@mkdir -p bin/
	go build -o bin/$(NAME)

test: deps
	go test $(go list ./... | grep -v /vendor/)

xcompile: deps test
	@rm -rf build/
	@mkdir -p build
	gox \
		-os="darwin" \
		-os="linux" \
		-output="build/{{.Dir}}_$(VERSION)_{{.OS}}_{{.Arch}}/$(NAME)"

package: xcompile
	$(eval FILES := $(shell ls build))
	@mkdir -p build/tgz
	for f in $(FILES); do \
		(cd $(shell pwd)/build && tar -zcvf tgz/$$f.tar.gz $$f); \
		echo $$f; \
	done

.PHONY: all deps updatedeps build test xcompile package
