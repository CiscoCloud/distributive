NAME = $(shell awk -F\" '/^const Name/ { print $$2 }' main.go)
VERSION = $(shell awk -F\" '/^const Version/ { print $$2 }' main.go)

# Should we use a system-wide copy of glide, or a locally built one?
GLIDE = $(shell which glide > /dev/null && echo "glide" || echo "./glide")

all: glide deps build test package

# Build glide in this directory. This 1. works in travis and 2. doesn't
# clutter up developers' machines with unnecessary programs.
glide:
	which glide > /dev/null || go get -v github.com/Masterminds/glide
	which glide > /dev/null || go build -o glide github.com/Masterminds/glide

deps: glide
	$(GLIDE) install

updatedeps: glide
	$(GLIDE) update

build: deps
	@mkdir -p bin/
	go build -o bin/$(NAME)

test: glide deps
	#go test -v $(shell $(GLIDE) novendor)
	go test -v ./chkutil/... ./dockerstatus/... ./errutil/... ./fsstatus/... ./netstatus/... ./systemdstatus/... ./tabular/... .

package: build
	tar -zcvf bin/$(NAME).tar.gz bin/$(NAME)

.PHONY: all glide deps build test package
