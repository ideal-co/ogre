package srvc

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"github.com/lowellmower/ogre/pkg/log"
	msg "github.com/lowellmower/ogre/pkg/message"
)

// DockerAPIClient is an interface which wraps a subset of a number of the
// Docker API interfaces. By limiting our client to only the interfaces we
// use, we are able to reduce code footprint and more easily test.
type DockerAPIClient interface {
	// subset of the ContainerAPIClient interface
	ContainerInspect(ctx context.Context, container string) (types.ContainerJSON, error)
	ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error)
	ContainerTop(ctx context.Context, container string, arguments []string) (types.ContainerProcessList, error)

	// subset of the SystemAPIClient interface
	Events(ctx context.Context, options types.EventsOptions) (<-chan events.Message, <-chan error)
	Info(ctx context.Context) (types.Info, error)
}

// DockerService satisfies the srvc.Service interface and is the struct for
// coordinating the Docker related services. DockerService communicates with
// the Daemon (main process) by means of a channel of msg.Message, which is
// its 'out' field. This channel is also the 'Daemon.In' field where other
// services and user input communicate with the Daemon.
type DockerService struct {
	Client DockerAPIClient

	ctx *Context
	in chan msg.Message
	out chan msg.Message
	err chan msg.Message
}

// NewDockerService takes a chan of msg.Message and a pointer to a Context, both
// of which are associated with the Daemon. The out channel passed corresponds
// to the Daemon.In channel and is used to send information back to the daemon
// to routing.
func NewDockerService(out, in, err chan msg.Message) (*DockerService, error) {
	dockerClient, e := client.NewEnvClient()
	if e != nil {
		return nil, e
	}

	return &DockerService{
		Client: dockerClient,
		ctx: NewDefaultContext(),
		in: in,
		out: out,
		err: err,
	}, nil
}

// Type satisfies the Service.Type interface and returns a ServiceType of Docker
func (ds *DockerService) Type() ServiceType {
	return Docker
}

// Start is the DockerService implementation of the Service interface's Start
// function. It calls the private method listen() which will start a loop to
// listen for signals from the daemon as well as spin off a go routine to begin
// listening on the Docker API for Events. Start will never return an error for
// DockerService as errors from this service are reported to the daemon by way
// of a channel of msg.Message which the Err() implementation returns a non-nil
// value.
func (ds *DockerService) Start() error {
	log.Daemon.Tracef("Starting %s service", ds.Type())
	ds.listen()
	return nil
}

// Stop is the DockerService implementation of the Service interface's Stop
// function. It signals to the loop running from a call to listen() that it
// should exit. This signal is sent by means of a msg.Message of type
// DockerMessage being sent over a channel of msg.Message located on the 'in'
// field of the DockerService. The shutdown signal will ultimately call the
// Cancel() method on the context associated with the DockerService, which is
// independent of the context associated with the Daemon or any other Service.
func (ds *DockerService) Stop() error {
	log.Daemon.Tracef("Stopping %s service", ds.Type())
	ds.in <- msg.DockerMessage{Action: "shutdown"}
	return nil
}

// listen is called by the Start() method and will begin a loop to listen for
// any msg.Message passed over it's 'in' channel. This channel is the means by
// which the daemon can send signals recieved from user input to the service.
func (ds *DockerService) listen() {
	signal := make(chan struct{})
	go ds.listenDockerAPI(signal)
	defer close(signal)
	for {
		select {
		case m := <-ds.in:
			log.Service.WithField("service", Docker).Tracef("docker listen got %+v", m)
			switch m.(msg.DockerMessage).Action {
			case "stop":
				signal <- struct{}{}
				ds.ctx = NewDefaultContext()
			case "start":
				go ds.listenDockerAPI(signal)
			case "shutdown":
				ds.ctx.Cancel()
				return
			}
		}
	}
}

// listenDockerAPI takes a channel of struct and begins listening for Events
// from the Docker API. The signal channel is used only to indicate to the API
// listening loop that there was signal from the daemon to stop this part of
// the service but not to shutdown, i.e. the service can be started again.
func (ds *DockerService) listenDockerAPI(signal chan struct{}) {
	dockerEvents, errChan := ds.Client.Events(ds.ctx.Ctx, types.EventsOptions{})
	for {
		select {
		case <-signal:
			// NOTE: we may also need/want to do something with the
			//       ds.ctx.Ctx passed to ds.Client.Events above but
			//       this is also the service level context so we must
			//       not cancel it. Simply returning for now.
			log.Service.WithField("service", Docker).Trace("stopping container listener...")
			return
		case err := <- errChan:
			log.Service.WithField("service", Docker).Tracef("err in listen %s\n", err)
			ds.err <- msg.NewDockerMessage(events.Message{}, err)
		case dEvent := <-dockerEvents:
			switch dEvent.Type {
			// Container Events from Docker API
			case events.ContainerEventType:
				switch dEvent.Action {
				case "start":
					// Add a container to the list being monitored should the label
					// indicating it should be monitored exist
					log.Service.WithField("service", Docker).Infof("docker action %s\n", dEvent.Action)
					ds.out <- msg.NewDockerMessage(dEvent, nil)
				case "restart":
					log.Service.WithField("service", Docker).Infof("docker action %s\n", dEvent.Action)
					ds.out <- msg.NewDockerMessage(dEvent, nil)
				case "stop":
					log.Service.WithField("service", Docker).Infof("docker action %s\n", dEvent.Action)
					ds.out <- msg.NewDockerMessage(dEvent, nil)
				case "die":
					log.Service.WithField("service", Docker).Infof("docker action %s\n", dEvent.Action)
					ds.out <- msg.NewDockerMessage(dEvent, nil)
				// introduced in docker v1.12 (2016)
				case "health_status":
					log.Service.WithField("service", Docker).Infof("docker action %s\n", dEvent.Action)
					ds.out <- msg.NewDockerMessage(dEvent, nil)
				}
			}
		}
	}
}

/*
   "State": {
       "Status": "running",
       "Running": true,
       "Paused": false,
       "Restarting": false,
       "OOMKilled": false,
       "Dead": false,
       "Pid": 12700,
       "ExitCode": 0,
       "Error": "",
       "StartedAt": "2020-05-06T16:36:01.074570319Z",
       "FinishedAt": "0001-01-01T00:00:00Z",
       "Health": {
           "Status": "unhealthy",
           "FailingStreak": 9,
           "Log": [
               {
                   "Start": "2020-05-06T16:36:26.960121047Z",
                   "End": "2020-05-06T16:36:27.16199565Z",
                   "ExitCode": 1,
                   "Output": ""
               }
           ]
       }
   },
 */