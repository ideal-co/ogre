package backend

import (
	"fmt"
	"github.com/ideal-co/ogre/pkg/config"
	"github.com/ideal-co/ogre/pkg/log"
	msg "github.com/ideal-co/ogre/pkg/message"
	"github.com/ideal-co/ogre/pkg/types"
)

type Platform interface {
	Type() types.PlatformType
	BackendClient
}

type BackendClient interface {
	Send(msg.Message) error
}

// NewBackendClient takes a types.PlatformType and an address and returns a
// typed backend.Platform interface and an error which will be nil upon
// successful initialization of the Platform.
func NewBackendClient(pType types.PlatformType, conf config.BackendConfig) (Platform, error) {
	switch pType {
	case types.StatsdBackend:
		return NewStatsdClient(conf.Server, conf.Prefix)
	case types.HTTPBackend:
		return NewHTTPBackend(conf.Server, conf.ResourcePath, conf.Format)
	case types.PrometheusBackend:
		return NewPrometheusBackend(conf.Server, conf.Metric, conf.ResourcePath)
	case types.DefaultBackend:
		// our default backend should be the service log but without the logrus
		// formatting when messages are written.
		return NewDefaultBackend(log.Daemon.Out)
	default:
		// we didn't get a recognized backend type
		return nil, fmt.Errorf("could not establish a backend for address: %s", conf.Server)
	}
}
