package backend

import (
	"github.com/lowellmower/ogre/pkg/config"
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

func NewBackendClient(pType types.PlatformType, addr string) (Platform, error) {
	switch pType {
	case types.StatsdBackend:
		prefix := config.Daemon.GetString("backends.statsd.server.prefix")
		return NewStatsdClient(addr, prefix)
	default:
		// TODO (lmower): return some default backend type here
		return nil, nil
	}
}
