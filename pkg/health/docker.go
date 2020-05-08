package health

import (
    "os/exec"
    "time"
)

// DockerHealthCheck satisfies the HeathCheck interface and encapsulates the
// behaviors for issuing healthchecks in or against docker containers
type DockerHealthCheck struct {
    // the command to be run
    // LABEL ogre.health.in.unique.check.one={{"nc -vz 127.0.0.1 8000"}}
    Cmd exec.Cmd

    // where the check will be executed, i.e. internal or external to a
    // container
    // LABEL ogre.health.{{in}}.unique.check.one="nc -vz 127.0.0.1 8000"
    Destination string

    // the interval at which to run the healthcheck
    Interval time.Duration

    // information about how to operate the healthchecks
    Format HealthCheckFormatter
}

type DockerFormatter struct {
    // LABEL ogre.format.health.output.type={{"string"}}
    Output interface{}
    // LABEL ogre.format.health.interval={{"5s"}}
    Interval string
    // LABEL ogre.format.health.output.result={{"exit"}}
    Result string
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

func (dhcf DockerFormatter) Parse(labels map[string]string) []HealthCheck {
    return []HealthCheck{&DockerHealthCheck{Format: dhcf}}
}

func NewDockerHealthCheck(labels map[string]string) []HealthCheck {
    fmttr := DockerFormatter{}
    dhc := fmttr.Parse(labels)
    return dhc
}
