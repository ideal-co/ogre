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

featured_image: '/img/ogre-green-l.svg'
featured_image_preview: '/img/ogre-green-l.svg'

comment: true
toc: true
autoCollapseToc: false
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
Note that, wherever you wind up placing the `ogred` binary, you make sure to
provide the additional configuration, as the application defaults to expect the
bin in `/usr/local/bin/`.

## Using Public Docker Image
If you would prefer to run ogre in a container, there is a public image available
to do so. To get started, pull the image:
```
docker pull idealco/ogre:latest
```
Once you've pulled the image locally, you can run it using the default config by
issuing a docker run command with the docker daemon socket mounted:
```
docker run -dit --name ogre -v /run/docker.sock:/run/docker.sock idealco/ogre:latest
```
Ogre will write check results to the log at `/var/log/ogred.log` which you can
manually inspect by issuing a docker exec command:
```
docker exec -it ogre tail -f /var/log/ogred.log
```
In another shell, you can start a basic container with an idempotent health check:
```
docker run -dit --name foo-noodle -l ogre.health.echo='echo hello world' alpine:latest
```
You should then see information showing up in the logs you are tailing in the other
shell.

If you are providing custom configuration (which you likely are or should be) you
will need to mount that config in via another volume:
```
docker run -dit --name ogre                           \ 
-v /run/docker.sock:/run/docker.sock                  \
-v /foo/my.ogre.conf:/etc/ogre/ogre.d/ogred.conf.json \
idealco/ogre:latest
```
You can issue ogre commands to the CLI via the docker exec as well:
```
docker exec ogre ogre version
Version: 0.1.0
Git Commit: 055e089e4074b03c61892406b0da5c06e031093b
Build Date: 2020-06-01-15:45:04
Go Version: go1.14.3
OS / Arch: linux amd64
```

## Configuration
Configuration is parsed every time the CLI is run. If a configuration file at
`/etc/ogre/ogre.d/ogred.conf.json` does not exist, a default configuration will
be written from source into memory and ultimately to disk at that location.

Should the file already exist, that file will be used to inform the configuration
of the application. The configuration must be in valid JSON and can be checked by
running `ogre config list` which should panic if invalid and otherwise print:
```
ogre config list
{
    "dockerd_socket": "/run/docker.sock",
    "containerd_socket": "/run/containerd/containerd.sock",
    "ogred_socket": "/var/run/ogred.sock",
    "ogred_pid": "/etc/ogre/ogred.pid",
    "log": {
        "level": "trace",
        "file": "/var/log/ogred.log",
        "silent": false,
        "report_caller": false
    }
}
```
Below is a complete configuration file with all possible backends configured:
```
{
    "dockerd_socket": "/run/docker.sock",
    "containerd_socket": "/run/containerd/containerd.sock",
    "ogred_socket": "/var/run/ogred.sock",
    "ogred_pid": "/etc/ogre/ogred.pid",
    "log": {
        "level": "trace",
        "file": "/var/log/ogred.log",
        "silent": false,
        "report_caller": false
    },
    "backends": [
        {
            "type": "statsd",
            "server": "127.0.0.1:8125",
            "prefix": "ogre"
        },
        {
            "type": "prometheus",
            "server": "127.0.0.1:9099",
            "metric": "east_coast_ogre_checks",
            "resource_path": "/metrics"
        },
        {
            "type": "http",
            "server": "127.0.0.1:9009",
            "format": "json",
            "resource_path": "/health"
        }
    ]
}
``` 
### Daemon Configuration
The daemon is configured by the JSON block which is provided as the default: 
```
{
    "dockerd_socket": "/run/docker.sock",
    "ogred_socket": "/var/run/ogred.sock",
    "ogred_pid": "/etc/ogre/ogred.pid",
    "ogred_bin": "/usr/local/bin/"
    "log": {
        "level": "trace",
        "file": "/var/log/ogred.log",
        "silent": false,
        "report_caller": false
    }
}
```
#### `dockerd_socket`
- Default: `/run/docker.sock`
- Desc: The location of the docker daemon unix socket
- Required: `true`
#### `ogred_socket`
- Default: `/var/run/ogred.sock`
- Desc: The location of the ogre daemon unix socket
- Required: `false`
#### `ogred_pid`
- Default: `/etc/ogre/ogred.pid`
- Desc: The location of the ogre daemon PID file
- Required: `false`
#### `ogred_bin`
- Default: `/usr/local/bin/`
- Desc: The location of the ogre daemon binary
- Required: `false`
#### `log`
- Default: see log config section
- Desc: The log configuration for the daemon
- Required: `false`

### Log Configuration
```
"log": {
    "level": "trace",
    "file": "/var/log/ogred.log",
    "silent": false,
    "report_caller": false
}
```
#### `level`
- Values: `info`,`warn`,`error`,`trace`
- Default: `info`
- Desc: The log level by which to report/log information
- Required: `true`
#### `file`
- Default: `/var/log/ogred.log`
- Desc: The file where ogre application will log
- Required: `true`
#### `silent`
- Values: `true`,`false`
- Default: `false`
- Desc: Indicate if all logging should be silent of not
- Required: `false`
#### `report_caller`
- Values: `true`,`false`
- Default: `false`
- Desc: Report the log line and information about code exec
- Required: `false`

### Backends
There are three supported backend types outside of the default log.
#### Prometheus
```
    {
        "type": "prometheus",
        "server": "127.0.0.1:9099",
        "metric": "east_coast_ogre_checks",
        "resource_path": "/metrics"
    }
```
#### `type`
- Values: `prometheus`
- Default: n/a
- Required: `true`
- Desc: Indicate to ogre a prometheus instance expects to scrape for health metrics

#### `server`
- Values: `ip|domain:port`
- Default: n/a
- Required: `true`
- Desc: The address to expose metrics to prometheus

#### `metric`
- Values: user defined
- Default: `ogre_health`
- Required: `false`
- Desc: The metric name used for the collector registered

#### `resource_path`
- Values: user defined
- Default: `/metrics`
- Required: `false`
- Desc: The resource path prometheus expects to scrape

#### Statsd
```
    {
        "type": "statsd",
        "server": "127.0.0.1:8125",
        "prefix": "ogre"
    }
```
#### `type`
- Values: `statsd`
- Default: n/a
- Required: `true`
- Desc: Indicate to ogre a statsd instance can accept health metrics

#### `server`
- Values: `ip|domain:port`
- Default: n/a
- Required: `true`
- Desc: The statsd address ogre will send metrics to

#### `prefix`
- Values: user defined
- Default: `ogre`
- Required: `false`
- Desc: The prefix used in the dot separated notation of the metric 

#### HTTP
```
    {
        "type": "http",
        "server": "127.0.0.1:9009",
        "format": "json",
        "resource_path": "/health"
    }
```
#### `type`
- Values: `http`
- Default: n/a
- Required: `true`
- Desc: Indicate to ogre an HTTP server will accept POST requests for metrics

#### `server`
- Values: `ip|domain:port`
- Default: n/a
- Required: `true`
- Desc: The address of the HTTP server ogre will send JSON encoded metrics to

#### `format`
- Values: `json`
- Default: `json`
- Required: `false`
- Desc: The content type to send

#### `resource_path`
- Values: user defined
- Default: `/health`
- Required: `false`
- Desc: Any additional resource pathing to be appended to the `server` config

All three require at minimum `server` and `type` configurations. All three can 
be used in unison, or the `backends` stanza can be omitted entirely. See below
for detail.
```
"backends": [
    {
        "type": "prometheus",
        "server": "127.0.0.1:9099",
        "metric": "east_coast_ogre_checks",
        "resource_path": "/metrics"
    },
    {
        "type": "statsd",
        "server": "127.0.0.1:8125",
        "prefix": "ogre"
    },
    {
        "type": "http",
        "server": "127.0.0.1:9009",
        "format": "json",
        "resource_path": "/health"
    }
]
```
#### `type`
- Values: `prometheus`,`statsd`, `http`
- Default: n/a
- Required: `true`
- Desc: The backend type which ogre will communicate health results to

#### `server`
- Values: `ip:port`
- Default: n/a
- Required: `true`
- Desc: The address at which to send or expose health results


## Dockerfile Configuration
```dockerfile
FROM alpine

# if the command is to be run internal to the container, the label immediately
# following 'health' must be 'in'
LABEL ogre.health.in.unique.check.one="ping -c 1 127.0.0.1"

# if the command is to be run external to the container, the label immediately
# following 'health' must be 'ex' - note that if you are not running ogre on
# the host, i.e. running in a container, the ogre container must also be part
# of any custom network configurations you may have, if any, in order to reach
# a container.  
LABEL ogre.health.ex.unique.check.two="nc -vz 172.17.0.3 8000"

# if neither 'in' nor 'ex' is present, ogre will attempt to make the check from
# inside the container
LABEL ogre.health.unique.check.three="echo inside"

# to enable the health checks to be reported to prometheus use the label
# LABEL ogre.format.backend.prometheus="true"

# Currently the below is not supported but may be offered as another unique label
# in a future iteration, for now use the ogre config file
# LABEL ogre.format.backend.prometheus.metric="metric name"

# Currently the below is not supported but may be offered as another unique label
# in a future iteration, for now use the ogre config file
# LABEL ogre.format.backend.prometheus.label="label name"

# enable the statsd backend
# LABEL ogre.format.backend.statsd="true"

# enable a generic http backend which can accept json via a POST request
# LABEL ogre.format.backend.http="true"

# if you could like to collect the output of healthchecks and send that value you
# can format the health checks like below
#
# (collect output from the command as a string)
# LABEL ogre.format.health.output.type="string"
#
# (collect output from the command as an int)
# LABEL ogre.format.health.output.type="int"
#
# (collect output from the command as a float)
# LABEL ogre.format.health.output.type="float"

# inform ogre to collect the exit code of health checks (default)
LABEL ogre.format.health.output.result="exit"

# inform ogre to collect the return of the command itself, i.e. if you had a
# check like `ls /proc | wc -l` 
# LABEL ogre.format.health.output.result="return"

# the interval at which checks should be run
LABEL ogre.format.health.interval="5s"

ENTRYPOINT ["nc", "-lke", "127.0.0.1", "8000"]
```

## Host Health
_coming soon..._
