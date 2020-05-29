---
title: "Ogre CLI Documentation"
date: 2020-5-6T16:22:42+08:00
lastmod: 2020-5-6T16:22:42+08:00
draft: false
description: "Documentation for the ogre CLI commands"
show_in_homepage: true
show_description: false
license: 'MIT'

tags: ['Documentation']
categories: ['programming', 'open source', 'ogre']

featured_image: ''
featured_image_preview: ''

comment: true
toc: true
autoCollapseToc: true
math: true
---

## Overview
#### What Ogre Is...
Ogre is a CLI and daemon delivering out of the box container monitoring by way
of docker label defined health checks. Ogre can be installed as a binary on the
host, or deployed using the public Docker image.

Ogre enables developers to quickly integrate a new or existing service with
existing reporting infrastructure without having to manage multiple configurations
and bespoke reporting implementations. Ogre makes it simple to report application
health by consolidating the collection and reporting mechanisms, leaving only
the responsibility of the health check definition to the developer.

Ogre currently integrates with the following reporting mechanisms:
- Prometheus
- Statsd
- Generic HTTP server (webhook)
- Logs  

#### What Ogre Is Not...
Ogre is not a stand-alone reporting tool, it is designed to integrate with popular
and custom solutions.

Ogre is not a process/container manager. No action is taken outside of sending
health check results to existing infrastructure.

Ogre is not (currently) a tool for reporting resource footprints or other pieces
of telemetry outside of container health checks, though it is possible this may
be the way the project heads.  

## Quick Start
#### Requirements
- Docker `> 1.9`
- User perms to CRUD files/dir under `/etc/` and `/var/`
#### Deploying on a host
For the CLI and daemon, you can clone and build yourself if you choose...
```
# clone the repo
https://github.com/ideal-co/ogre.git
cd ogre

# build the binaries
make build

# ensure the install was successful
ogre version

# start the daemon
ogre start
```
Now you'll want to provide a label on your services container, indicating what
the health check should be. (`-l` is the label flag)
```
docker run -dit --name foo-noodle                     \
-l ogre.health.ping.outside='ping -c 1 -W 1 8.8.8.8'  \
alpine:latest
```
By default, Ogre will log the check as JSON to `/var/log/ogred.log` if no
backend configuration is provided otherwise. The default interval at which
checks are run is `5s`.
```
tail -n 1 /var/log/ogred.log | jq .
{
  "CompletedCheck": null,
  "Destination": "",
  "Data": {
    "Container": "/foo-noodle",
    "Hostname": "daae3a5a717f",
    "Exit": 0,
    "StdOut": "PING 8.8.8.8 (8.8.8.8): 56 data bytes\n64 bytes from 8.8.8.8: seq=0 ttl=37 time=23.473 ms\n\n--- 8.8.8.8 ping statistics ---\n1 packets transmitted, 1 packets received, 0% packet loss\nround-trip min/avg/max = 23.473/23.473/23.473 ms\n",
    "StdErr": ""
  },
  "Err": null
}
```
The same `docker run` command could be given another label to indicate that this
health check should only be run every `1m` versus the default `5s`.
```
docker run -dit --name foo-noodle                     \
-l ogre.health.ping.outside='ping -c 1 -W 1 8.8.8.8'  \
-l ogre.format.health.interval='1m'                   \
alpine:latest
```
Multiple, more complex checks can be passed. These could also be part of your
`Dockerfile` or `docker-compose.yml` file, they do not need to be flags. 
```
docker run -dit --name rev-prox \
-l ogre.health.in.https.open="nmap --host-timeout 1s -p 443 127.0.0.1 | grep -i closed | wc -l | awk '{$1=$1;exit $1}'"  \
-l ogre.health.ex.dns.connect="nmap --host-timeout 1s -p 53 8.8.8.8 | grep -i closed | wc -l | awk '{$1=$1;exit $1}'"    \
-l ogre.health.service.check.script="./usr/local/bin/your_health_check_script.sh"                                        \
your_nginx_img:latest 
``` 
#### Running Ogre in a Container
... TODO ... 

## Building Ogre
If you're building from source, clone the repo and use the Makefile target
which will result in the `ogre` and `ogred` binary in `/usr/bin/local/`. This
will require you have Go version `1.14.2` or greater installed. Older Go versions
will most likely work but cannot be guaranteed.
```
make build
```
Likewise, you can build the binaries yourself by executing the go build commands:
```
go build ./cmd/ogre
go build ./cmd/ogred
```
Note, that wherever you wind up placing the `ogred` binary, you make sure to
provide the additional configuration, as the application defaults to expect the
bin in `/usr/local/bin/`.

## Using Public Docker Image
_coming soon..._

## Configuration


### Daemon
_coming soon..._
### Logs
_coming soon..._
### Backends
_coming soon..._

## Docker Health
_coming soon..._

## Host Health
_coming soon..._
