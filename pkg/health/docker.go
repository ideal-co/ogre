package health

import (
    "context"
    "github.com/lowellmower/ogre/pkg/log"
    "github.com/lowellmower/ogre/pkg/types"
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
    // the slice string representation of the command
    RawCmd []string

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

    // the outcome of a health check's command
    Result ExecResult

    mu sync.Mutex
}

type DockerFormatter struct {
    Output FormatOutput
    Platform FormatPlatform
}

// ExecResult is the encapsulating struct used to capture the output from a command
// run internal or external to a container.
type ExecResult struct {
    Exit int
    StdOut string
    StdErr string
}

type FormatOutput struct {
    Type string
    Result string
}

type FormatPlatform struct {
    Target types.PlatformType
    // Metric will only be set for types.PrometheusBackend
    Metric string
    // Job will only be set for types.PrometheusBackend
    Job string
}

func (dhc *DockerHealthCheck) String() string {
    return dhc.Name
}

func (dhc *DockerHealthCheck) ExitCode() int {
    return dhc.Result.Exit
}

func (dhc *DockerHealthCheck) Passed() bool {
    return dhc.Result.Exit == 0
}

func NewDockerFormatter(platLabels, opLabels map[string]string) *DockerFormatter {
    return &DockerFormatter{
        Output:        parseOutputFromLabels(opLabels),
        Platform:      parsePlatformFromLabels(platLabels),
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
                if interval, ok := labels["ogre.format.health.interval"]; ok {
                    dur, err := time.ParseDuration(interval)
                    hc.Interval = dur
                    if err != nil {
                        log.Service.Errorf("could not parse time %s from label", interval)
                        hc.Interval = 5 * time.Second
                    }

                }
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

func parsePlatformFromLabels(backendLabels map[string]string) FormatPlatform {
   fp := FormatPlatform{}
   for key, val := range backendLabels {
       splitKey := strings.Split(key, ".")
       switch splitKey[space] {
       case formatBackendStatsd:
           fp.Target = types.StatsdBackend
       case formatBackendProm:
           fp.Target = types.PrometheusBackend
           switch splitKey[subSpaceOne] {
           case prometheusMetric:
               fp.Metric = val
           case prometheusJob:
               fp.Job = val
           }
       }
   }

   return fp
}


func parseHealthCheck(dest string, name []string, cmd string) *DockerHealthCheck {
    var c DockerHealthCheck
    c.Ctx, c.cancel = context.WithCancel(context.Background())
    switch dest {
    case internalCheck:
        c.Name = strings.Join(name[1:], "_")
        c.Destination = dest
        c.Cmd = getCommand(c.Ctx, cmd)
        c.RawCmd = strings.Split(cmd, " ")
        return &c
    case externalCheck:
        c.Name = strings.Join(name[1:], "_")
        c.Destination = dest
        c.Cmd = getCommand(c.Ctx, cmd)
        c.RawCmd = strings.Split(cmd, " ")
        return &c
    default:
        c.Name = strings.Join(name[0:], "_")
        c.Destination = "internal"
        c.Cmd = getCommand(c.Ctx, cmd)
        c.RawCmd = strings.Split(cmd, " ")
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
