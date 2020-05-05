package srvc

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lowellmower/ogre/pkg/log"
	"io"

	msg "github.com/lowellmower/ogre/pkg/message"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
)

// DockerAPIClient is an interface which wraps a subset of a number of the
// Docker API interfaces. By limiting our client to only the interfaces we
// use, we are able to reduce code footprint and more easily test.
type DockerAPIClient interface {
	// subset of the SystemAPIClient interface
	Events(ctx context.Context, options types.EventsOptions) (<-chan events.Message, <-chan error)
	Info(ctx context.Context) (types.Info, error)

	// subset of the ContainerAPIClient interface
	ContainerInspect(ctx context.Context, container string) (types.ContainerJSON, error)
	ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error)
	ContainerTop(ctx context.Context, container string, arguments []string) (types.ContainerProcessList, error)
}

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

// Read
func (ds *DockerService) Read(reader io.Reader) (msg.Message, error) {
	dm := msg.DockerMessage{}
	dec := json.NewDecoder(reader)
	err := dec.Decode(&dm)
	if err != nil {
		return dm, err
	}

	return dm, nil
}

func (ds *DockerService) Start() error {
	log.Daemon.Tracef("Starting %s service", ds.Type())
	ds.listen()
	return nil
}

func (ds *DockerService) Stop() error {
	log.Daemon.Tracef("Stopping %s service", ds.Type())
	ds.in <- msg.DockerMessage{Action: "shutdown"}
	return nil
}

func (ds *DockerService) listenDockerAPI(signal chan struct{}) {
	dockerEvents, errChan := ds.Client.Events(ds.ctx.Ctx, types.EventsOptions{})
	for {
		select {
		case <-signal:
			log.Daemon.Trace("stopping container listener...")
			return
		case err := <- errChan:
			log.Daemon.Tracef("err in listen %s\n", err)
			ds.err <- msg.NewDockerMessage(events.Message{}, err)
		case dEvent := <-dockerEvents:
			switch dEvent.Type {
			case events.ContainerEventType:
				switch dEvent.Action {
				case "start":
					fmt.Printf("docker action %s\n", dEvent.Action)
					ds.out <- msg.NewDockerMessage(dEvent, nil)
				case "restart":
					fmt.Printf("docker action %s\n", dEvent.Action)
					ds.out <- msg.NewDockerMessage(dEvent, nil)
				case "stop":
					fmt.Printf("docker action %s\n", dEvent.Action)
					ds.out <- msg.NewDockerMessage(dEvent, nil)
				case "die":
					fmt.Printf("docker action %s\n", dEvent.Action)
					ds.out <- msg.NewDockerMessage(dEvent, nil)
				}
			}
		}
	}
}

func (ds *DockerService) listen() {
	signal := make(chan struct{})
	go ds.listenDockerAPI(signal)
	defer close(signal)
	for {
		select {
		case m := <-ds.in:
			log.Daemon.Tracef("docker listen got %+v", m)
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