FROM golang:1.14-buster AS build-stage

ARG GIT_COMMIT
ARG VERSION
LABEL GIT_COMMIT=$GIT_COMMIT
LABEL VERSION=$VERSION
LABEL REPO="https://github.com/ideal-co/ogre"

ENV PROJPATH=/go/src/github.com/ideal-co/ogre

ENV PATH=$PATH:$GOROOT/bin:$GOPATH/bin

COPY . /go/src/github.com/ideal-co/ogre
WORKDIR /go/src/github.com/ideal-co/ogre

RUN make build

FROM ubuntu:latest

ENV PATH=$PATH:/usr/local/bin

COPY --from=build-stage /go/src/github.com/ideal-co/ogre/ogre /usr/local/bin/
COPY --from=build-stage /go/src/github.com/ideal-co/ogre/ogred /usr/local/bin/

RUN chmod +x /usr/local/bin/ogre
RUN chmod +x /usr/local/bin/ogred

ENTRYPOINT ["ogred"]
