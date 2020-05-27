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
	Ctx    context.Context
	cancel context.CancelFunc

	// where the check will be executed, i.e. inside of the
	// container or against the container from host
	Destination string

	// the interval at which to run the health check
	Interval time.Duration

	// information about how to operate the health checks
	Formatter *DockerFormatter

	// the outcome of a health check's command
	Result *ExecResult

	mu sync.Mutex
}

// DockerFormatter encapsulates the format for the output and backend as it is
// parsed from the labels returned from the Docker API for a single container.
type DockerFormatter struct {
	Output   FormatOutput
	Platform FormatPlatform
}

// ExecResult is the encapsulating struct used to capture the output from a command
// run internal or external to a container.
type ExecResult struct {
	Hostname string
	Exit     int
	StdOut   string
	StdErr   string
}

// FormatOutput is the struct representation of the ogre.format.output.$ labels.
type FormatOutput struct {
	Type   string
	Result string
}

// FormatPlatform is the struct representation of the ogre.format.backend.$ labels.
type FormatPlatform struct {
	Target types.PlatformType
	// Metric will only be set for types.PrometheusBackend
	Metric string
	// Label will only be set for types.PrometheusBackend
	Label string
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

// newDockerFormatter takes two maps of labels where key and value are both
// strings and parses out the FormatOutput and FormatPlatform from them. If
// any values are missing or configured incorrectly, default values will be
// used in an effort to provide functionality out of the box.
func newDockerFormatter(platLabels, opLabels map[string]string) *DockerFormatter {
	df := &DockerFormatter{
		Output:   parseOutputFromLabels(opLabels),
		Platform: parsePlatformFromLabels(platLabels),
	}

	return df
}

// NewDockerHealthCheck takes a map of labels where key and value are both
// strings and returns a slice of pointers to DockerHealthCheck. The labels
// passed are from the Docker API and represent the labels of a container
// running on the same host as ogred.
func NewDockerHealthCheck(labels map[string]string) []*DockerHealthCheck {
	var checks []*DockerHealthCheck
	formatter := newFormatterFromLabels(labels)

	for key, val := range labels {
		splitKey := strings.Split(key, ".")
		if splitKey[ogre] == "ogre" {
			// if we got incomplete values passed, bail
			if len(splitKey) <= space {
				break
			}
			switch splitKey[space] {
			case health:
				hc := &DockerHealthCheck{}
				hc.Formatter = formatter
				hc.parseHealthCheck(splitKey[subSpaceOne], splitKey[subSpaceOne:], val)
				if interval, ok := labels["ogre.format.health.interval"]; ok {
					dur, err := time.ParseDuration(interval)
					hc.Interval = dur
					if err != nil {
						log.Service.Errorf("could not parse time %s from label", interval)
						hc.Interval = 5 * time.Second
					}

				}
				hc.setDefaultIfEmpty()
				checks = append(checks, hc)
			}
		}
	}

	return checks
}

func (dhc *DockerHealthCheck) setDefaultIfEmpty() {
	// ogre.health.{in, ex}.check.name
	// default location to run a check
	if len(dhc.Destination) == 0 {
		log.Service.Info("health check destination was empty, using default 'in' (internal)")
		dhc.Destination = "in"
	}
	// ogre.format.health.{interval}
	// defaults to 5s check
	if dhc.Interval == 0 {
		log.Service.Info("health check interval was empty, using default '5s' (5 seconds)")
		dhc.Interval = time.Second * 5
	}
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
	f := newDockerFormatter(fmtBackendMap, fmtHealthMap)
	log.Daemon.Tracef("FORMATTER: %+v", *f)
	return f
}

func parseOutputFromLabels(outputLabels map[string]string) FormatOutput {
	fo := FormatOutput{}
	for key, val := range outputLabels {
		splitKey := strings.Split(key, ".")
		// if we got incomplete values passed, bail
		if len(splitKey) <= space {
			break
		}
		switch splitKey[space] {
		case formatHeathOutput:
			// if we got incomplete values passed, bail
			if len(splitKey) <= subSpaceOne {
				break
			}
			switch splitKey[subSpaceOne] {
			case formatHealthOutputType:
				fo.Type = val
			case formatHealthOutputResult:
				fo.Result = val
			}
		}
	}
	fo.setDefaultIfEmpty()

	return fo
}

// setDefaultIfEmpty ensures that if labels were not passed to configure these
// fields, we use some default values in their place for FormatOutput.
func (fo *FormatOutput) setDefaultIfEmpty() {
	if len(fo.Type) == 0 {
		log.Service.Info("format output type missing, using default 'int' (integer)")
		fo.Type = "int"
	}
	if len(fo.Result) == 0 {
		log.Service.Info("format output result missing, using default 'exit' (exit code)")
		fo.Result = "exit"
	}
}

func parsePlatformFromLabels(backendLabels map[string]string) FormatPlatform {
	fp := &FormatPlatform{}
	for key, val := range backendLabels {
		splitKey := strings.Split(key, ".")
		// if we got incomplete values passed, bail
		if len(splitKey) <= space {
			break
		}
		switch splitKey[space] {
		case formatBackendStatsd:
			fp.Target = types.StatsdBackend
		case formatBackendHTTP:
			fp.Target = types.HTTPBackend
		case formatBackendProm:
			fp.Target = types.PrometheusBackend
			// if no other values were provided, bail
			// ogre.format.backend.{prometheus}="true"
			if len(splitKey) <= subSpaceOne {
				break
			}

			switch splitKey[subSpaceOne] {
			case prometheusMetric:
				fp.Metric = strings.Join(strings.Split(val, " "), "_")
			case prometheusLabel:
				fp.Label = strings.Join(strings.Split(val, " "), "_")
			}

		}
	}
	fp.setDefaultIfEmpty()

	return *fp
}

// setDefaultIfEmpty ensures that if labels were not passed to configure these
// fields, we use some default values in their place for FormatPlatform.
func (fp *FormatPlatform) setDefaultIfEmpty() {
	switch fp.Target {
	case types.StatsdBackend:
		// Do nothing, we only need to indicate this is our desired platform
		return
	case types.HTTPBackend:
		// Do nothing, we only need to indicate this is our desired platform
		return
	case types.PrometheusBackend:
		if len(fp.Label) == 0 {
			log.Service.Info("format backend prometheus job missing, using default name 'ogre_job'")
			fp.Label = "ogre_job"
		}
		if len(fp.Metric) == 0 {
			log.Service.Info("format backend prometheus metric missing, using default name 'ogre_metric'")
			fp.Metric = "ogre_metric"
		}
	case types.CollectdBackend:
		// TODO (lmower): issue #9
	default:
		log.Service.Info("format backend missing, will send health checks to log")
		fp.Target = types.DefaultBackend
	}
}

func (dhc *DockerHealthCheck) parseHealthCheck(dest string, name []string, cmd string) {
	dhc.Ctx, dhc.cancel = context.WithCancel(context.Background())
	switch dest {
	case internalCheck:
		// Docker label values passed in brackets
		// ogre.health.in.{0...n}=""
		dhc.formatNameByPlatform(name[1:])
		dhc.Destination = dest
		dhc.Cmd = getCommand(dhc.Ctx, cmd)
		dhc.RawCmd = strings.Split(cmd, " ")
	case externalCheck:
		// Docker label values passed in brackets
		// ogre.health.ex.{0...n}=""
		dhc.formatNameByPlatform(name[1:])
		dhc.Destination = dest
		dhc.Cmd = getCommand(dhc.Ctx, cmd)
		dhc.RawCmd = strings.Split(cmd, " ")
	default:
		// Docker label values passed in brackets
		// ogre.health.{0...n}=""
		dhc.formatNameByPlatform(name[0:])
		dhc.Destination = internalCheck
		dhc.Cmd = getCommand(dhc.Ctx, cmd)
		dhc.RawCmd = strings.Split(cmd, " ")
	}
}

// formatNameByPlatform will adjust the separating token on a health check name
// based upon platform or will set as the default which is an underscore.
func (dhc *DockerHealthCheck) formatNameByPlatform(name []string) {
	switch dhc.Formatter.Platform.Target {
	case types.StatsdBackend:
		// ogre.health.{in, ex}.some.check.name -> some.check.name
		dhc.Name = strings.Join(name, ".")
	default:
		// ogre.health.{in, ex}.some.check.name -> some_check_name
		dhc.Name = strings.Join(name, "_")
	}
}

// getCommand takes a context.Context and a string of space separated values
// and returns a pointer to a command. If the slice of arguments returned
// from splitting the string results in a zero length slice, we will return a
// nil value for the command. Both the context.Context passed and the returned
// exec.Cmd pointer are values associated with the corresponding fields of the
// HealthCheck which is being initialized.
func getCommand(ctx context.Context, com string) *exec.Cmd {
	comSlice := strings.Split(com, " ")
	if len(comSlice) < 1 {
		return nil
	}

	return exec.CommandContext(ctx, comSlice[0], comSlice[1:]...)
}
