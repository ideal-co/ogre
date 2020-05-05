# Build Stage
FROM lowellmower/alpine-ogre-build:1.13 AS build-stage

LABEL app="build-ogre"
LABEL REPO="https://github.com/lowellmower/ogre"

ENV PROJPATH=/go/src/github.com/lowellmower/ogre

# Because of https://github.com/docker/docker/issues/14914
ENV PATH=$PATH:$GOROOT/bin:$GOPATH/bin

ADD . /go/src/github.com/lowellmower/ogre
WORKDIR /go/src/github.com/lowellmower/ogre

RUN make build-alpine

# Final Stage
FROM lowellmower/alpine-base:latest

ARG GIT_COMMIT
ARG VERSION
LABEL REPO="https://github.com/lowellmower/ogre"
LABEL GIT_COMMIT=$GIT_COMMIT
LABEL VERSION=$VERSION

# Because of https://github.com/docker/docker/issues/14914
ENV PATH=$PATH:/opt/ogre/bin

WORKDIR /opt/ogre/bin

COPY --from=build-stage /go/src/github.com/lowellmower/ogre/bin/ogre /opt/ogre/bin/
RUN chmod +x /opt/ogre/bin/ogre

# Create appuser
RUN adduser -D -g '' ogre
USER ogre

ENTRYPOINT ["/usr/bin/dumb-init", "--"]

CMD ["/opt/ogre/bin/ogre"]
