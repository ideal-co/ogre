package backend

type PlatformType string

const (
    Pometheus PlatformType = "prometheus"
    StdOut PlatformType = "stdout"
    Grafana PlatformType = "grafana"
)

type Platform interface {
    String(string) string
}
