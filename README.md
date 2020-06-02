[![CircleCI](https://circleci.com/gh/ideal-co/ogre.svg?style=shield)](https://circleci.com/gh/ideal-co/ogre)
[![Docker Pulls](https://img.shields.io/docker/pulls/idealco/ogre.svg?maxAge=604800)]
# Ogre
<img align="right" width="300" height="300" src=https://github.com/ideal-co/ogre-assets/blob/master/images/ogre-green-300.png>
Simplified health monitoring and telemetry for Docker containers in modern day
distributed systems.

## Getting Started
- See [quick start section](https://lowellmower.com/1/01/ogre-doc/#quick-start) of the documentation.
- [Getting started with statsd](). (coming soon...)
- [Getting started with a generic HTTP backend](). (coming soon...)
- [Getting started with prometheus](). (coming soon...)

## Building Ogre

Ensure you have Go installed. Running Ogre then should be as simple as:
```console
$ make build
$ ./ogre version
```

### Testing

``make test``

### Documentation

Local docs can be found under the [docs/README.md](./docs/README.md) which will
also be in sync with the [hosted docs here](https://lowellmower.com/1/01/ogre-doc/).
