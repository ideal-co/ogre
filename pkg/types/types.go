package types

type PlatformType string

const (
    PrometheusBackend PlatformType = "prometheus"
    StatsdBackend PlatformType = "statsd"
    CollectdBackend PlatformType = "collectd"
    HTTPBackend PlatformType = "http"
    GrafanaBackend PlatformType = "grafana"
    GraphiteBackend PlatformType = "graphite"
)


type MessageType string

const (
    DaemonMessage MessageType = "daemon"
    DockerMessage MessageType = "docker"
    BackendMessage MessageType = "backend"
    HostMessage MessageType = "host"
)

type ServiceType string

const (
    CLI    ServiceType = "CLI"
    DockerService ServiceType = "docker"
    BackendService ServiceType = "backend"
)