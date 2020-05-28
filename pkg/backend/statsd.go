package backend

import (
	"github.com/cactus/go-statsd-client/statsd"
	"github.com/lowellmower/ogre/pkg/log"
	msg "github.com/lowellmower/ogre/pkg/message"
	"github.com/lowellmower/ogre/pkg/types"
)

// StatsdBackend implements the Platform interface and is responsible for the
// sending of health check results to a statsd instance.
type StatsdBackend struct {
	Client statsd.Statter
}

// NewStatsdClient takes two strings, an address and a prefix and returns a new
// pointer to StatsdBackend which satisfies the Platform interface, or an error.
// If no prefix is provided, the field will be an empty string.
func NewStatsdClient(addr, prefix string) (Platform, error) {
	conf := &statsd.ClientConfig{
		Address: addr,
		Prefix:  prefix,
	}

	client, err := statsd.NewClientWithConfig(conf)
	if err != nil {
		return nil, err
	}

	be := &StatsdBackend{
		Client: client,
	}

	return be, nil
}

// Send is the StatsdBackend's implementation of the Platform interface Send
// method. Send takes a Message and sends a metric to a statsd backend equal
// to that of the exit code of the health check. An error is returned on a
// failure to send.
func (sdb *StatsdBackend) Send(m msg.Message) error {
	bem := m.(msg.BackendMessage)
	log.Daemon.Tracef("statsd client listen got %+v", bem)
	return sdb.Client.Inc(bem.CompletedCheck.String(), int64(bem.CompletedCheck.ExitCode()), 1.0)
}

// Type is the StatsdBackend implementation of the Platform interface Type
// and returns a PlatformType of type StatsdBackend.
func (sdb *StatsdBackend) Type() types.PlatformType {
	return types.StatsdBackend
}
