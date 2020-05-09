package health

const (
    ogre int = iota    // 0 -> 'ogre'
    space              // 1 -> 'health, format'
    subSpaceOne        // 2 -> 'in, ex, backend, health`
    subSpaceTwo        // 3 -> 'prometheus, grafana, stdout, interval, output'
    subSpaceThree      // 4 -> 'metric, job, result, type'

    health = "health"
    format = "format"

    internalCheck = "in"
    externalCheck = "ex"

    formatHeath = "health"
    formatBackend = "backend"

    formatBackendProm = "prometheus"
    prometheusMetric = "metric"
    prometheusJob = "job"

    formatHeathOutput = "output"
    formatHeathInterval = "interval"

    formatHealthOutputType = "type"
    formatHealthOutputResult = "result"
)

// HealthCheckResult is the interface by which various parts of the application
// will read and understand the results of a healthcheck command execution.
type HealthCheck interface {
    String() string
    ExitCode() int
    Passed() bool
}

type HealthCheckFormatter interface {
    Parse(map[string]string) []HealthCheck
}
