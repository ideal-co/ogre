package health

import (
    "os/exec"
    "time"
)

type DockerHealthCheck struct {
    // the command to be run
    // LABEL ogre.health.in.unique.check.one={{"nc -vz 127.0.0.1 8000"}}
    Cmd exec.Cmd

    // where the check will be sent to from the daemon
    Destination string

    // the interval at which to run the healthcheck
    Interval time.Duration

    // information about how to operate the healthchecks
    Format HealthCheckFormatter
}

type DockerFormatter struct {
    // LABEL ogre.format.health.output.{{string}}
    Output interface{}
    // LABEL ogre.format.health.interval={{"5s"}}
    Interval string
}

func (dhc *DockerHealthCheck) String() string {
    return ""
}

func (dhc *DockerHealthCheck) ExitCode() int {
    return 0
}

func (dhc *DockerHealthCheck) Passed() bool {
    return true
}

func (dhcf DockerFormatter) Parse() HealthCheck {
    return &DockerHealthCheck{}
}
