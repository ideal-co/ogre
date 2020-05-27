package backend

import (
	msg "github.com/lowellmower/ogre/pkg/message"
	"github.com/lowellmower/ogre/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

// PrometheusBackend
type PrometheusBackend struct {
	Registry    *prometheus.Registry
	CounterVec  *prometheus.CounterVec
	MetricPath  string
	Metric      string
	Label       string
	msgToMetric chan prometheus.Metric
}

func NewPrometheusBackend(server, metric, path string) (Platform, error) {
	pbe := &PrometheusBackend{MetricPath: path, msgToMetric: make(chan prometheus.Metric)}
	if len(metric) > 0 {
		pbe.Metric = metric
	} else {
		pbe.Metric = "ogre_health_check"
	}
	pbe.CounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: pbe.Metric,
			Help: "Ogre executed health checks.",
		},
		[]string{"host", "check", "health"},
	)
	//pbe.Registry = prometheus.NewPedanticRegistry()
	//pbe.Registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	prometheus.MustRegister(pbe.CounterVec)
	http.Handle(pbe.MetricPath, promhttp.Handler())
	go http.ListenAndServe(server, nil)

	return pbe, nil
}

func (p *PrometheusBackend) Send(m msg.Message) error {
	bem := m.(msg.BackendMessage)
	host := bem.Data.Hostname
	check := bem.CompletedCheck.String()
	result := bem.Data.Exit

	// we always want to attempt to register because the backend doesn't know
	// if this is a 'new' check result or not
	prometheus.Register(p.CounterVec)
	if result >= 1 {
		// to achieve binary reporting, we always reset the counter before adding
		p.CounterVec.Reset()
		p.CounterVec.With(prometheus.Labels{"host": host, "check": check, "health": "unhealthy"}).Add(1)
	} else {
		// if our exit code is 0, we reset the counter so as to stop reporting
		p.CounterVec.Reset()
	}

	return nil
}

func (p *PrometheusBackend) Type() types.PlatformType {
	return types.PrometheusBackend
}
