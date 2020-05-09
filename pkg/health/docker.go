package health

import (
    "context"
    "github.com/lowellmower/ogre/pkg/backend"
    "github.com/lowellmower/ogre/pkg/log"
    "os/exec"
    "strings"
    "sync"
    "time"
)

// DockerHealthCheck satisfies the HeathCheck interface and encapsulates the
// behaviors for issuing health checks in or against docker containers
type DockerHealthCheck struct {
    Name string
    // the command to be run
    Cmd *exec.Cmd

    // context associated with the command to be run and the
    // corresponding cancel function
    Ctx context.Context
    cancel context.CancelFunc

    // where the check will be executed, i.e. inside of the
    // container or against the container from host
    Destination string

    // the interval at which to run the health check
    Interval time.Duration

    // information about how to operate the health checks
    Formatter *DockerFormatter

    mu sync.Mutex
}

type DockerFormatter struct {
    Output FormatOutput
    Interval time.Duration

    Backend backend.Platform
    BackendLabels map[string]string
}

type FormatOutput struct {
    Type string
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

func NewDockerFormatter(beLabels, opLabels map[string]string) *DockerFormatter {
    return &DockerFormatter{
        Output:        parseOutputFromLabels(opLabels),
        Backend:       parseBackendFromLabels(beLabels),
        BackendLabels: make(map[string]string),
    }
}

func NewDockerHealthCheck(labels map[string]string) []*DockerHealthCheck {
    var checks []*DockerHealthCheck
    formatter := newFormatterFromLabels(labels)

    for key, val := range labels {
        splitKey := strings.Split(key, ".")
        if splitKey[ogre] == "ogre" {
            switch splitKey[space] {
            case health:
                hc := parseHealthCheck(splitKey[subSpaceOne], splitKey[subSpaceOne:], val)
                hc.Formatter = formatter
                checks = append(checks, hc)
            }
        }
    }

    return checks
}

func newFormatterFromLabels(labels map[string]string) *DockerFormatter {
    fmtBackendMap := make(map[string]string)
    fmtHealthMap := make(map[string]string)

    for key, val := range labels {
        splitKey := strings.Split(key, ".")
        if splitKey[ogre] == "ogre" {
            switch splitKey[space] {
            case format:
                switch splitKey[subSpaceOne] {
                case formatBackend:
                    fmtBackendMap[strings.Join(splitKey[subSpaceOne:], ".")] = val
                case formatHeath:
                    fmtHealthMap[strings.Join(splitKey[subSpaceOne:], ".")] = val
                }
            }
        }
    }
    f := NewDockerFormatter(fmtBackendMap, fmtHealthMap)
    if interval, ok := labels["ogre.format.health.interval"]; ok {
        dur, err := time.ParseDuration(interval)
        if err != nil {
            log.Service.Errorf("could not parse time %s from label", interval)
        }
        f.Interval = dur
    }

    return f
}

func parseOutputFromLabels(outputLabels map[string]string) FormatOutput {
    var out FormatOutput
    for key, val := range outputLabels {
        splitKey := strings.Split(key, ".")
        switch splitKey[space] {
        case formatHeathOutput:
            switch splitKey[subSpaceOne] {
            case formatHealthOutputType:
                out.Type = val
            case formatHealthOutputResult:
                out.Result = val
            }
        }
    }

    return out
}

func parseBackendFromLabels(backendLabels map[string]string) backend.Platform {
    var be backend.Platform
    for key, val := range backendLabels {
        splitKey := strings.Split(key, ".")
        switch splitKey[space] {
        case formatBackendProm:
            if be == nil {
                be = backend.NewPrometheusBackend()
            }
            switch splitKey[subSpaceOne] {
            case prometheusMetric:
                be.(*backend.Prometheus).Metric = val
            case prometheusJob:
                be.(*backend.Prometheus).Job = val
            }
        }
    }

    return be
}


func parseHealthCheck(dest string, name []string, cmd string) *DockerHealthCheck {
    var c DockerHealthCheck
    c.Ctx, c.cancel = context.WithCancel(context.Background())
    switch dest {
    case internalCheck:
        c.Name = strings.Join(name[1:], "_")
        c.Destination = dest
        c.Cmd = getCommand(c.Ctx, cmd)
        return &c
    case externalCheck:
        c.Name = strings.Join(name[1:], "_")
        c.Destination = dest
        c.Cmd = getCommand(c.Ctx, cmd)
        return &c
    default:
        c.Name = strings.Join(name[0:], "_")
        c.Destination = "internal"
        c.Cmd = getCommand(c.Ctx, cmd)
        return &c
    }
}

func getCommand(ctx context.Context, com string) *exec.Cmd{
    comSlice := strings.Split(com, " ")
    if len(comSlice) < 1 {
        return nil
    }

    return exec.CommandContext(ctx, comSlice[0], comSlice[1:]...)
}
