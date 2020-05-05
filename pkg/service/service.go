package srvc

import (
	"context"
	"fmt"
	msg "github.com/lowellmower/ogre/pkg/message"
)

type ServiceType string

const (
	CLI    ServiceType = "CLI"
	Docker ServiceType = "docker"
)

// Service is the interface which all services will implement.
type Service interface {
	Type() ServiceType
	//Read(io.Reader) (msg.Message, error)
	Start() error
	Stop() error
}

// ServiceCallback is an anonymous function which is executed upon the closing
// of a srvc.Context and can be used as an additional means of reporting or
// clean up when a context is canceled.
type ServiceCallback func() error

type Context struct {
	Ctx context.Context
	Cancel context.CancelFunc
	Callback ServiceCallback
}

// NewService takes a ServiceType and a channel of msg.Message and returns a
// Service interface to be stored on the Daemon's service field, keyed by the
// ServiceType.
func NewService(s ServiceType, in, out, err chan msg.Message) (Service, error) {
	switch s {
	case Docker:
		return NewDockerService(in, out, err)
	default:
		return nil, fmt.Errorf("could not establish service type: %s", s)
	}
}

// NewDefaultContext returns a pointer to Context struct which is a wrapper
// around the standard libs context. The context associated with the daemon
// is the context which will communicate runtime state to the rest of the
// program, e.g. the running services.
func NewDefaultContext() *Context {
	ctx, cancel := context.WithCancel(context.Background())
	return &Context{
		Ctx: ctx,
		Cancel: cancel,
		Callback: nil,
	}
}

// Done calls the underlying context's Done method
func (c *Context) Done() <-chan struct{} {
	return c.Ctx.Done()
}

// Err calls the underlying context's Err method
func (c *Context) Err() error {
	return c.Ctx.Err()
}
