package health

// HealthCheckResult is the interface by which various parts of the application
// will read and understand the results of a healthcheck command execution.
type HealthCheck interface {
    String() string
    ExitCode() int
    Passed() bool
}

type HealthCheckFormatter interface {
    Parse(map[string]string) HealthCheck
}
