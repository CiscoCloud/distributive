FROM ubuntu:wily
MAINTAINER Langston Barrett <langston@aster.is> (@siddharthist)

# If this file doesn't immedately work for you, please submit a Github issue:
# https://github.com/CiscoCloud/distributive/issues

# This docker container should run and then stop immediately when the checklist
# has been completed.

# Note that Distributive doesn't have access to certain system information when
# run in a container.

RUN apt-get update
# network: net-tools
# misc: lm-sensors, php5-cli, module-init-tools
RUN apt-get install -y module-init-tools bash golang git php5-cli lm-sensors net-tools && apt-get clean

WORKDIR /gopath/src/github.com/CiscoCloud/distributive
RUN mkdir -p /gopath/{bin,src}
ENV GOPATH /gopath
ENV GOBIN /gopath/bin
ENV PATH $PATH:/gopath/bin
RUN go get github.com/tools/godep
ADD . /gopath/src/github.com/CiscoCloud/distributive
# Note: docker-machine on Windows / OS X sometimes gets its time out of sync,
# which can cause SSL verification failures. If this happens, `go get .`, will
# fail. If you run into this problem, run this command at your terminal:
# docker-machine ssh ${machine} 'sudo ntpclient -s -h pool.ntp.org'
# After that, you can retry `docker build .`.
RUN godep go install .
# Tests don't all pass in this Docker container, but they will soon!
#RUN godep go test  --short ./...

CMD ["distributive", "-d", "/gopath/src/github.com/CiscoCloud/distributive/samples/", "--verbosity", "info"]
