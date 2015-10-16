FROM gliderlabs/alpine:3.2
MAINTAINER Langston Barrett <langston@aster.is> (@siddharthist)

# If this file doesn't immedately work for you, please submit a Github issue:
# https://github.com/CiscoCloud/distributive/issues

# This docker container should run and then stop immediately when the checklist
# has been completed.

# Note that Distributive doesn't have access to certain system information when
# run in a container.

RUN apk update && apk add go git && rm -rf /var/cache/apk/*

WORKDIR /gopath/src/github.com/CiscoCloud/distributive
RUN mkdir -p /gopath/{bin,src}
ENV GOPATH /gopath
ENV GOBIN /gopath/bin
ENV PATH $PATH:/gopath/bin
RUN go get github.com/tools/godep
ADD . /gopath/src/github.com/CiscoCloud/distributive
# Note: docker-machine on Windows / OS X sometimes gets its time out of sync, which can cause SSL verification failures.
# If this happens, `go get .`, will fail. If you run into this problem, uncomment out the following line and re-run `docker build .`.
# RUN ntpd -d -q -n -p 0.pool.ntp.org
RUN go get .
RUN go build .

CMD ["distributive", "-f", "/gopath/src/github.com/CiscoCloud/distributive/samples/filesystem.json", "-d", "", "--verbosity", "info"]
