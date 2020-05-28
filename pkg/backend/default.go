package backend

import (
	msg "github.com/lowellmower/ogre/pkg/message"
	"github.com/lowellmower/ogre/pkg/types"
	"io"
)

// DefaultBackend is the backend which should always be initialized in an effort
// to make ogre useful out of the box. By establishing this default, we enable
// health checks to report to a log should there be no alternate backend such as
// statsd, collectd, prometheus, etc.
type DefaultBackend struct {
	Logger io.Writer
}

// NewDefaultBackend returns a Platform interface which is a pointer to a struct
// of type DefaultBackend and an error.
func NewDefaultBackend(log io.Writer) (Platform, error) {
	return &DefaultBackend{Logger: log}, nil
}

// Type is the DefaultBackend implementation of the Platform interface Type
// and returns a PlatformType of type DefaultBackend.
func (dbe *DefaultBackend) Type() types.PlatformType {
	return types.DefaultBackend
}

// Send implementation for DefaultBackend will serialize the backend messgae and
// write it to the io.Writer, which, is the same writer used by the log.Service
// instance of logrus.Logger. However, the serialized message is written as JSON
// and without any of the logrus formatting. Error is returned if encountered.
func (dbe *DefaultBackend) Send(m msg.Message) error {
	bem := m.(msg.BackendMessage)
	data, err := bem.Serialize()
	if err != nil {
		return err
	}
	_, err = dbe.Logger.Write(data)
	return err
}
