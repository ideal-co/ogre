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
    in chan msg.Message
    out chan msg.Message
    err chan msg.Message
}

func (bes *BackendService) Start() error {
    bes.listen()
    return nil
}

func (bes *BackendService) Stop() error {
    return nil
}

func (bes *BackendService) Type() types.ServiceType {
    return types.BackendService
}

func NewBackendService(out, in, errChan chan msg.Message) (*BackendService, error) {
    return &BackendService{
        Platforms: make(map[types.PlatformType]backend.Platform),
        ctx:       NewDefaultContext(),
        in:        in,
        out:       out,
        err:       errChan,
    }, nil
}

func (bes *BackendService) listen() {
    for {
        select {
        case m := <-bes.in:
            bem := m.(msg.BackendMessage)
            log.Service.WithField("service", types.BackendService).Tracef("backend listen got %+v", bem)

            if be, ok := bes.Platforms[bem.Destination]; ok {
                err := be.Send(m)
                if err != nil {
                    log.Service.Errorf("could not send message to %s: %s", bem.Destination, err)
                }
            } else {
                log.Service.Infof("%s", bem.CompletedCheck.String())
            }
            //bes.Platforms[bem.]
            //if err := sdb.Client.SetInt(bem.CompletedCheck.String(), int64(bem.CompletedCheck.ExitCode()), 0.5); err != nil {
            //    sdb.err <-msg.GenericMessage{Err: err}
            //}
        }
    }
}