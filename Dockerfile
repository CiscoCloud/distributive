FROM gliderlabs/alpine:3.1
MAINTAINER Langston Barrett <lagnston@aster.is> (@siddharthist)

# This docker container should run and then stop immediately when the checklist
# has been completed

# If this file doesn't immedately work for you, please submit a Github issue:
# https://github.com/CiscoCloud/distributive/issues

RUN apk update && apk add go && rm -rf /var/cache/apk/*

ENV GOROOT /usr/lib/go
ENV GOPATH /distributive
ENV GOBIN /distributive/bin
ENV PATH $PATH:$GOROOT/bin:$GOPATH/bin

WORKDIR /distributive
ADD . /distributive
RUN go build /distributive/distributive.go

CMD [/distributive/distributive -f /distributive/samples/filesystem.json]
