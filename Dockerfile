FROM golang:1.5-onbuild

MAINTAINER SpectoLabs

ADD . /go/src/github.com/spectolabs/twitter-app

ENV GO15VENDOREXPERIMENT 1

RUN go install github.com/spectolabs/twitter-app

ENTRYPOINT /go/bin/twitter-app

EXPOSE 8080
