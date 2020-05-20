FROM golang:1.14-buster

RUN apt install make git gcc

LABEL REPO="https://github.com/lowellmower/ogre"

ENV PROJPATH=/go/src/github.com/lowellmower/ogre
# Because of https://github.com/docker/docker/issues/14914
ENV PATH=$PATH:$GOROOT/bin:$GOPATH/bin

COPY . /go/src/github.com/lowellmower/ogre
WORKDIR /go/src/github.com/lowellmower/ogre

CMD make build
