FROM gliderlabs/alpine:3.1
MAINTAINER Langston Barrett <lagnston@aster.is> (@siddharthist)

RUN apk update && apk add go && rm -rf /var/cache/apk/*

ENV GOROOT /usr/lib/go
ENV GOPATH /distributive
ENV GOBIN /distributive/bin
ENV PATH $PATH:$GOROOT/bin:$GOPATH/bin

WORKDIR /distributive
ADD . /distributive
RUN go build /distributive/distributive.go

CMD [/distributive/distributive -f /gopath/src/distributve/samples/sleep.json]
