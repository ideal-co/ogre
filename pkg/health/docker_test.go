package health

import (
	"context"
	"github.com/docker/docker/pkg/testutil/assert"
	"github.com/ideal-co/ogre/pkg/types"
	"testing"
	"time"
)

func TestNewDockerHealthCheck(t *testing.T) {
	testIO := []struct {
		name string
		in   map[string]string
		exp  []*DockerHealthCheck
	}{
		{
			name: "should return a single health check with defaults",
			in: map[string]string{
				"ogre.health.foo.check": "./usr/bin/foo.sh",
			},
			exp: []*DockerHealthCheck{
				{
					Name:        "foo_check",
					Cmd:         getCommand(context.Background(), "./usr/bin/foo.sh"),
					RawCmd:      []string{"./usr/bin/foo.sh"},
					Destination: "in",
					Interval:    time.Second * 5,
					Formatter:   newFormatterFromLabels(make(map[string]string)),
				},
			},
		},
		{
			name: "should return two health checks with defaults",
			in: map[string]string{
				"ogre.health.foo.check":     "./usr/bin/foo.sh",
				"ogre.health.foo.bar.check": "./usr/bin/foo_bar.sh",
			},
			exp: []*DockerHealthCheck{
				{
					Name:        "foo_check",
					Cmd:         getCommand(context.Background(), "./usr/bin/foo.sh"),
					RawCmd:      []string{"./usr/bin/foo.sh"},
					Destination: "in",
					Interval:    time.Second * 5,
					Formatter:   newFormatterFromLabels(make(map[string]string)),
				},
				{
					Name:        "foo_bar_check",
					Cmd:         getCommand(context.Background(), "./usr/bin/foo_bar.sh"),
					RawCmd:      []string{"./usr/bin/foo_bar.sh"},
					Destination: "in",
					Interval:    time.Second * 5,
					Formatter:   newFormatterFromLabels(make(map[string]string)),
				},
			},
		},
		{
			name: "should return a single health check with format configuration statsd",
			in: map[string]string{
				"ogre.health.foo.check":      "./usr/bin/foo.sh",
				"ogre.format.backend.statsd": "true",
			},
			exp: []*DockerHealthCheck{
				{
					Name:        "foo.check",
					Cmd:         getCommand(context.Background(), "./usr/bin/foo.sh"),
					RawCmd:      []string{"./usr/bin/foo.sh"},
					Destination: "in",
					Interval:    time.Second * 5,
					Formatter: newFormatterFromLabels(map[string]string{
						"ogre.format.backend.statsd": "true",
					}),
				},
			},
		},
		{
			name: "should return a single health check with all format configuration",
			in: map[string]string{
				"ogre.health.foo.check":                 "./usr/bin/foo.sh",
				"ogre.format.backend.prometheus.metric": "foo_metric",
				"ogre.format.backend.prometheus.label":  "foo_job",
				"ogre.format.health.output.type":        "string",
				"ogre.format.health.output.result":      "return",
			},
			exp: []*DockerHealthCheck{
				{
					Name:        "foo_check",
					Cmd:         getCommand(context.Background(), "./usr/bin/foo.sh"),
					RawCmd:      []string{"./usr/bin/foo.sh"},
					Destination: "in",
					Interval:    time.Second * 5,
					Formatter: newFormatterFromLabels(map[string]string{
						"ogre.format.backend.prometheus.metric": "foo_metric",
						"ogre.format.backend.prometheus.label":  "foo_job",
						"ogre.format.health.output.type":        "string",
						"ogre.format.health.output.result":      "return",
					}),
				},
			},
		},
	}
	for _, io := range testIO {
		t.Run(io.name, func(t *testing.T) {
			dhc := NewDockerHealthCheck(io.in)
			for idx, hc := range dhc {
				assert.Equal(t, hc.Name, io.exp[idx].Name)
				assert.Equal(t, hc.Cmd.String(), io.exp[idx].Cmd.String())
				assert.DeepEqual(t, hc.RawCmd, io.exp[idx].RawCmd)
				assert.Equal(t, hc.Interval, io.exp[idx].Interval)
				assert.DeepEqual(t, hc.Formatter, io.exp[idx].Formatter)
			}
		})
	}
}

func TestNewFormatterFromLabels(t *testing.T) {
	testIO := []struct {
		name string
		in   map[string]string
		exp  *DockerFormatter
	}{
		{
			name: "should make a formatter with default values",
			in:   make(map[string]string),
			exp: &DockerFormatter{
				Output: FormatOutput{
					Type:   "int",
					Result: "exit",
				},
				Platform: FormatPlatform{
					Target: types.DefaultBackend,
				},
			},
		},
		{
			name: "should make a formatter with default value for output type",
			in: map[string]string{
				"ogre.format.health.output.result": "return",
			},
			exp: &DockerFormatter{
				Output: FormatOutput{
					Type:   "int",
					Result: "return",
				},
				Platform: FormatPlatform{
					Target: types.DefaultBackend,
				},
			},
		},
		{
			name: "should make a formatter with default value for output result",
			in: map[string]string{
				"ogre.format.health.output.type": "float",
			},
			exp: &DockerFormatter{
				Output: FormatOutput{
					Type:   "float",
					Result: "exit",
				},
				Platform: FormatPlatform{
					Target: types.DefaultBackend,
				},
			},
		},
		{
			name: "should not panic for output having invalid values",
			in: map[string]string{
				"ogre.format.health.output": "foo",
			},
			exp: &DockerFormatter{
				Output: FormatOutput{
					Type:   "int",
					Result: "exit",
				},
				Platform: FormatPlatform{
					Target: types.DefaultBackend,
				},
			},
		},
		{
			name: "should make a formatter for statsd with default values",
			in: map[string]string{
				"ogre.format.backend.statsd": "true",
			},
			exp: &DockerFormatter{
				Output: FormatOutput{
					Type:   "int",
					Result: "exit",
				},
				Platform: FormatPlatform{
					Target: types.StatsdBackend,
				},
			},
		},
		{
			name: "should make a formatter for prometheus with default job",
			in: map[string]string{
				"ogre.format.backend.prometheus.metric": "foo_bar",
			},
			exp: &DockerFormatter{
				Output: FormatOutput{
					Type:   "int",
					Result: "exit",
				},
				Platform: FormatPlatform{
					Target: types.PrometheusBackend,
					Metric: "foo_bar",
					Label:  "ogre_job",
				},
			},
		},
		{
			name: "should make a formatter for prometheus with default metric",
			in: map[string]string{
				"ogre.format.backend.prometheus.label": "foo_bar",
			},
			exp: &DockerFormatter{
				Output: FormatOutput{
					Type:   "int",
					Result: "exit",
				},
				Platform: FormatPlatform{
					Target: types.PrometheusBackend,
					Metric: "ogre_metric",
					Label:  "foo_bar",
				},
			},
		},
		{
			name: "should make a formatter for prometheus with default values",
			in: map[string]string{
				"ogre.format.backend.prometheus": "true",
			},
			exp: &DockerFormatter{
				Output: FormatOutput{
					Type:   "int",
					Result: "exit",
				},
				Platform: FormatPlatform{
					Target: types.PrometheusBackend,
					Metric: "ogre_metric",
					Label:  "ogre_job",
				},
			},
		},
		{
			name: "should make a fully configured formatter for prometheus",
			in: map[string]string{
				"ogre.format.backend.prometheus.metric": "foo_metric",
				"ogre.format.backend.prometheus.label":  "foo_job",
				"ogre.format.health.output.type":        "string",
				"ogre.format.health.output.result":      "return",
			},
			exp: &DockerFormatter{
				Output: FormatOutput{
					Type:   "string",
					Result: "return",
				},
				Platform: FormatPlatform{
					Target: types.PrometheusBackend,
					Metric: "foo_metric",
					Label:  "foo_job",
				},
			},
		},
	}
	for _, io := range testIO {
		t.Run(io.name, func(t *testing.T) {
			fmtr := newFormatterFromLabels(io.in)
			assert.DeepEqual(t, fmtr, io.exp)
		})
	}
}

func TestParseOutputFromLabels(t *testing.T) {
	testIO := []struct {
		name string
		in   map[string]string
		exp  FormatOutput
	}{
		{
			name: "should return a default from empty labels",
			in:   make(map[string]string),
			exp: FormatOutput{
				Type:   "int",
				Result: "exit",
			},
		},
		{
			name: "should return a type and default",
			in: map[string]string{
				"health.output.type": "string",
			},
			exp: FormatOutput{
				Type:   "string",
				Result: "exit",
			},
		},
		{
			name: "should return a result and default",
			in: map[string]string{
				"health.output.result": "return",
			},
			exp: FormatOutput{
				Type:   "int",
				Result: "return",
			},
		},
	}
	for _, io := range testIO {
		t.Run(io.name, func(t *testing.T) {
			fmtr := parseOutputFromLabels(io.in)
			assert.DeepEqual(t, fmtr, io.exp)
		})
	}
}

func TestParsePlatformFromLabels(t *testing.T) {
	testIO := []struct {
		name string
		in   map[string]string
		exp  FormatPlatform
	}{
		{
			name: "should return a default from empty labels",
			in:   make(map[string]string),
			exp: FormatPlatform{
				Target: types.DefaultBackend,
			},
		},
		{
			name: "should return a statsd target",
			in: map[string]string{
				"backend.statsd": "true",
			},
			exp: FormatPlatform{
				Target: types.StatsdBackend,
			},
		},
		{
			name: "should return a prometheus target and defaults",
			in: map[string]string{
				"backend.prometheus": "true",
			},
			exp: FormatPlatform{
				Target: types.PrometheusBackend,
				Metric: "ogre_metric",
				Label:  "ogre_job",
			},
		},
		{
			name: "should return a prometheus target a metric, and a default",
			in: map[string]string{
				"backend.prometheus":        "true",
				"backend.prometheus.metric": "foo_metric_name",
			},
			exp: FormatPlatform{
				Target: types.PrometheusBackend,
				Metric: "foo_metric_name",
				Label:  "ogre_job",
			},
		},
		{
			name: "should return a prometheus target a metric and job",
			in: map[string]string{
				"backend.prometheus":        "true",
				"backend.prometheus.metric": "foo_metric_name",
				"backend.prometheus.label":  "foo_job_name",
			},
			exp: FormatPlatform{
				Target: types.PrometheusBackend,
				Metric: "foo_metric_name",
				Label:  "foo_job_name",
			},
		},
	}
	for _, io := range testIO {
		t.Run(io.name, func(t *testing.T) {
			fmtr := parsePlatformFromLabels(io.in)
			assert.DeepEqual(t, fmtr, io.exp)
		})
	}
}
