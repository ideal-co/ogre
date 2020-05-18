package srvc

import (
	"github.com/lowellmower/ogre/pkg/backend"
	"github.com/lowellmower/ogre/pkg/log"
	msg "github.com/lowellmower/ogre/pkg/message"
	"github.com/lowellmower/ogre/pkg/types"
)

// BackendService satisfies the Service interface and is responsible for routing
// Messages on its 'in' channel to the various backend platforms
type BackendService struct {
	Platforms map[types.PlatformType]backend.Platform

	ctx *Context
	in  chan msg.Message
	out chan msg.Message
	err chan msg.Message
}

// Start is the BackendService implementation of the Service interface's Start
// method and calls out to a private method listen.
func (bes *BackendService) Start() error {
	log.Service.Infof("starting %s service", bes.Type())
	bes.listen()
	return nil
}

// Stop is the BackendService implementation of the Service interface's Stop
func (bes *BackendService) Stop() error {
	log.Service.Infof("stopping %s service", bes.Type())
	bes.ctx.Cancel()
	return nil
}

// Type is the BackendService implementation of the Service interface's Type
func (bes *BackendService) Type() types.ServiceType {
	return types.BackendService
}

// NewBackendService takes three channels of Messages which connect to the
// Daemon and returns a pointer to a BackendService and an error which is nil
// upon success.
func NewBackendService(out, in, errChan chan msg.Message) (*BackendService, error) {
	return &BackendService{
		Platforms: make(map[types.PlatformType]backend.Platform),
		ctx:       NewDefaultContext(),
		in:        in,
		out:       out,
		err:       errChan,
	}, nil
}

// listen is kicked off from the Service interface Run method which will begin
// an infinite loop
func (bes *BackendService) listen() {
	for {
		select {
		case <-bes.ctx.Done():
			return
		case m := <-bes.in:
			bem := m.(msg.BackendMessage)
			log.Service.WithField("service", bem.Type()).Tracef("backend listen got %+v", bem)

			if be, ok := bes.Platforms[bem.Destination]; ok {
				if err := be.Send(m); err != nil {
					log.Service.Errorf("could not send message to %s from %s: %s", bem.Destination, bes.Type(), err)
				}
			}
		}
	}
}
