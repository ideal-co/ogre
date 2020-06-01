package backend

import (
	"github.com/ideal-co/ogre/pkg/log"
	msg "github.com/ideal-co/ogre/pkg/message"
	"github.com/ideal-co/ogre/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

// PrometheusBackend satisfies the Platform interface and is responsible for
// exposing health check metrics to be scraped by a prometheus instance via an
// http endpoint.
type PrometheusBackend struct {
	CounterVec *prometheus.CounterVec
	MetricPath string
	Metric     string
	Label      string
}

// NewPrometheusBackend takes three strings, a server (address) to listen on, a
// metric name to register a prometheus collector under, and a resource path by
// which a prometheus instance will scrap for metrics. The server string will be
// passed as a colin separated address and port, i.e. 127.0.0.1:9099. The path
// will be passed as a forward slash prefixed string representing the full path
// to which your prometheus instance expects to scrape for metrics, most setups
// use the resource path '/metrics' and will be the default.
func NewPrometheusBackend(server, metric, path string) (Platform, error) {
	pbe := &PrometheusBackend{MetricPath: path}
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

	// register the collector for prometheus to scrape
	if err := prometheus.Register(pbe.CounterVec); err != nil {
		return nil, err
	}

	// expose the handler and start the server to be scraped
	http.Handle(pbe.MetricPath, promhttp.Handler())
	go func() {
		err := http.ListenAndServe(server, nil)
		if err != nil {
			log.Daemon.Errorf("error starting prometheus server: %s", err)
		}
	}()

	return pbe, nil
}

// Send is the PrometheusBackend implementation of the Platform interface. Send
// will take a Message and present a prometheus metric to be scraped. This will
// expose to a prometheus instance a value of 1 should a health check be failing
// and will other wise not report.
func (p *PrometheusBackend) Send(m msg.Message) error {
	bem := m.(msg.BackendMessage)
	host := bem.Data.Hostname
	check := bem.CompletedCheck.String()
	result := bem.Data.Exit

	// we always want to attempt to register because the backend doesn't know
	// if this is a 'new' check result or not, we don't concern ourselves with
	// the error returned which would most often be an AlreadyRegisteredError
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

// Type is the PrometheusBackend implementation of the Platform interface Type
// and returns a PlatformType of type PrometheusBackend.
func (p *PrometheusBackend) Type() types.PlatformType {
	return types.PrometheusBackend
}
