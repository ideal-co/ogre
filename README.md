[![CircleCI](https://circleci.com/gh/ideal-co/ogre.svg?style=shield)](https://circleci.com/gh/ideal-co/ogre)
![Docker Pulls](https://img.shields.io/docker/pulls/idealco/ogre.svg?maxAge=604800)
# Ogre
<img align="right" width="300" height="300" src=https://github.com/ideal-co/ogre-assets/blob/master/images/ogre-green-300.png>
Simplified health monitoring and telemetry for Docker containers in modern day
distributed systems.

## Note:
This project is currently in **Beta** and is **not** production ready. Please use at
will but with that understanding. We are aiming for end of June for a hardened
release which will be version `1.0.0`.

Please open issues for bugs, breaks, panics, code reviews, and puppy pictures...
the feedback is valuable and appreciated.

## Getting Started
- See [quick start section](https://lowellmower.com/1/01/ogre-doc/#quick-start) of the documentation.
- [Getting started with statsd](https://youtu.be/MjhH5YD570U)
- [Getting started with a generic HTTP backend](https://youtu.be/jZ3DDbNvkX4)
- [Getting started with prometheus](https://youtu.be/680R_YYbaCQ)

## Building Ogre

Ensure you have Go installed. Running Ogre then should be as simple as:
```console
$ make build install
$ ogre version
```

### Testing

``make test``

### Documentation

Local docs can be found under the [docs/README.md](./docs/README.md) which will
also be in sync with the [hosted docs here](https://lowellmower.com/1/01/ogre-doc/).
