### Example Prometheus Integration
Within this directory contains all the necessary commands and files to run an
MVP prometheus setup with ogre. This is the same setup used in the 
[example video](https://youtu.be/680R_YYbaCQ). Please file an issue or PR if you
encounter errors following along.

It should be noted that all commands in this example presume you are running said
commands from within this directory, i.e. `ogre/examples/prometheus/`

### Installing Prometheus
```
GO15VENDOREXPERIMENT=1 go get github.com/prometheus/prometheus/cmd/...
```
Ensure that prometheus is installed:
```
prometheus --help
```
Note that you may need to run prometheus from whithin its install directory
given the resources for the UI, which it that is the case and the above
doesn't work, try:
```
cp prom.conf.yml cd $GOPATH/src/github.com/prometheus/prometheus/
cd $GOPATH/src/github.com/prometheus/prometheus/
go build cmd/...
./cmd/prometheus/prometheus --config.file=./prom.conf.yml
```
I found the install steps to be a bit lacking on the [Docker Hub](https://hub.docker.com/r/prom/prometheus/)
page, so you'll have to play around with it.

### Running Prometheus
Run prometheus with the configuration file provided in this directory:
```
prometheus --config.file=./prom.conf.yml
```

### Building Ogre
```
pushd ../.. && make build install && popd
```
Ensure that ogre is installed:
```
ogre version
Build Date: 2020-06-01-09:54:01
Git Commit: 055e089e4074b03c61892406b0da5c06e031093b
Version: 0.1.0
Go Version: go1.14.2
OS / Arch: darwin amd64
```

### Configuring Ogre
Run ogre with the config from this directory:
```
cp oged.conf.json /etc/ogre/ogre.d/
ogre start
```

### Build and Run the Docker Container
```
docker build --tag prom-ogre-example .
docker run -dt --name prom-ogre-example prom-ogre-example:latest
```

### Inspect
Navigate to `localhost:9090` to see the prometheus UI and search for the
metric name from the `ogred.conf.json` (i.e. `east_coast_ogre_checks`) and you
should see that they are reporting "unhealthy" because `health.sh` is simply
executing `exit 1`.

Toggle this by docker exec'ing into the `prom-ogre-example` container and
adjusting the exit code in the script (or make your own!):
```
docker exec -it prom-ogre-example sh
vi /usr/local/bin/health.sh
# change the exit code and save
```
Navigate back to UI and see that the check is no longer reporting unhealthy.
