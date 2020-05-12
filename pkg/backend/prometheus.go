package backend

import (
	msg "github.com/lowellmower/ogre/pkg/message"
	"github.com/lowellmower/ogre/pkg/types"
)

type PrometheusBackend struct {
	Metric   string
	Job      string
	Labels   []string
	platType types.PlatformType
}

func (p *PrometheusBackend) Send(m msg.Message) error {
	return nil
}

func (p *PrometheusBackend) Type() types.PlatformType {
	return types.PrometheusBackend
}
