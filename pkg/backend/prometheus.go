package backend

import "fmt"

type Prometheus struct {
    Metric string
    Job string
    Labels []string
}

// m, j string, l []string
func NewPrometheusBackend() *Prometheus {
    return &Prometheus{}
}

func (p *Prometheus) String(label string) string {
    return fmt.Sprintf("%s{label=%s}", p.Metric, label)
}

func (p *Prometheus) URL() string {
    // /metrics/job/some_job
    return fmt.Sprintf("/metrics/job/%s", p.Job)
}
