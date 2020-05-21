package types

import "errors"

// PlatformType is a string which is used in the constants of this package to
// implement a typing of sorts on platforms. Anything which implements the
// backend.Platform interface will return this type from its implementation of
// the Type() method.
type PlatformType string

const (
	PrometheusBackend PlatformType = "prometheus"
	StatsdBackend     PlatformType = "statsd"
	CollectdBackend   PlatformType = "collectd"
	HTTPBackend       PlatformType = "http"
	GrafanaBackend    PlatformType = "grafana"
	GraphiteBackend   PlatformType = "graphite"
	DefaultBackend    PlatformType = "log"
)

// MessageType is a string which is used in the constants of this package to
// implement a typing of sorts on messages. Anything which implements the
// msg.Message interface will return this type from its implementation of the
// Type() method.
type MessageType string

const (
	DaemonMessage  MessageType = "daemon"
	DockerMessage  MessageType = "docker"
	BackendMessage MessageType = "backend"
	HostMessage    MessageType = "host"
)

// ServiceType is a string which is used in the constants of this package to
// implement a typing of sorts on services. Anything which implements the
// srvc.Service interface will return this type from its implementation of
// the Type() method.
type ServiceType string

const (
	CLI            ServiceType = "CLI"
	DockerService  ServiceType = "docker"
	BackendService ServiceType = "backend"
)

// Error types
var ErrNoCheck = errors.New("no check health was present or parsed")