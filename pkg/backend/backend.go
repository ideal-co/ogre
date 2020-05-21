package backend

import (
	"fmt"
	"github.com/lowellmower/ogre/pkg/config"
	"github.com/lowellmower/ogre/pkg/log"
	msg "github.com/lowellmower/ogre/pkg/message"
	"github.com/lowellmower/ogre/pkg/types"
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
		// check to see if there is a configured prefix for statsd
		return NewStatsdClient(conf.Server, conf.Prefix)
	case types.HTTPBackend:
		//format := config.Daemon.GetString("backends.http.format")
		return NewHTTPBackend(conf.Server, conf.ResourcePath, conf.Format)
	case types.DefaultBackend:
		// our default backend should be the service log but without the logrus
		// formatting when messages are written.
		return NewDefaultBackend(log.Service.Out)
	default:
		// we didn't get a recognized backend type
		return nil, fmt.Errorf("could not establish a backend for address: %s", conf.Server)
	}
}
