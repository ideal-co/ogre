package srvc

import (
	"bytes"
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/lowellmower/ogre/pkg/backend"
	"github.com/lowellmower/ogre/pkg/health"
	"github.com/lowellmower/ogre/pkg/log"
	msg "github.com/lowellmower/ogre/pkg/message"
	"io/ioutil"
	"time"
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

	// https://pkg.go.dev/github.com/docker/docker/client@v1.13.1?tab=doc#ContainerAPIClient
	ContainerAttach(ctx context.Context, container string, options types.ContainerAttachOptions) (types.HijackedResponse, error)
	ContainerExecAttach(ctx context.Context, execID string, config types.ExecConfig) (types.HijackedResponse, error)
	ContainerExecCreate(ctx context.Context, container string, config types.ExecConfig) (types.IDResponse, error)
	ContainerExecInspect(ctx context.Context, execID string) (types.ContainerExecInspect, error)
	ContainerExecResize(ctx context.Context, execID string, options types.ResizeOptions) error
	ContainerExecStart(ctx context.Context, execID string, config types.ExecStartCheck) error
}

// DockerService satisfies the srvc.Service interface and is the struct for
// coordinating the Docker related services. DockerService communicates with
// the Daemon (main process) by means of a channel of msg.Message, which is
// its 'out' field. This channel is also the 'Daemon.In' field where other
// services and user input communicate with the Daemon.
type DockerService struct {
	Client DockerAPIClient
	Containers []*Container
	Backend *backend.Platform

	ctx *Context
	in chan msg.Message
	out chan msg.Message
	err chan msg.Message
}

// Container is the struct representation of a Docker container on the host the
// docker daemon ogre is communicating with is running. It encapsulates the list
// of HealthChecks to execute.
type Container struct {
	Name string
	ID string
	Info types.ContainerJSON
	HealthChecks []*health.DockerHealthCheck
}

// ExecResult is the encapsulating struct used to capture the output from a command
// run internal or external to a container.
type ExecResult struct {
	Exit int
	StdOut string
	StdErr string
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

// NewDockerService takes a chan of msg.Message and a pointer to a Context, both
// of which are associated with the Daemon. The out channel passed corresponds
// to the Daemon.In channel and is used to send information back to the daemon
// to routing.
func NewDockerService(out, in, errChan chan msg.Message) (*DockerService, error) {
	dockerClient, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	ds := &DockerService{
		Client: dockerClient,
		ctx: NewDefaultContext(),
		in: in,
		out: out,
		err: errChan,
	}

	ds.Containers, err = ds.CollectContainers()
	if err != nil {
		return nil, err
	}

	// TODO (lmower): need to go through the checks collected and ensure that
	//                all backend.PlatformTypes have platforms to accept output

	return ds, nil
}

func NewContainer(info types.ContainerJSON) *Container {
	// parse label info
	heathChecks := health.NewDockerHealthCheck(info.Config.Labels)

	// instantiate container
	c := &Container{Info: info, HealthChecks: heathChecks}
	c.ID = info.ID
	c.Name = info.Name

	return c
}

func (ds *DockerService) CollectContainers()([]*Container, error){
	var cList []*Container
	arg, _ := filters.FromParam("status=running")
	opts := types.ContainerListOptions{Filters: arg}

	// gather running containers
	containers, err := ds.Client.ContainerList(ds.ctx.Ctx, opts)
	if err != nil {
		return nil, err
	}

	// gather info on containers
	for _, c := range containers {
		info, err := ds.Client.ContainerInspect(ds.ctx.Ctx, c.ID)
		if err != nil {
			log.Service.WithField("service", Docker).Errorf("could not get info for %s: %s", c.ID, err)
			continue
		}

		cList = append(cList, NewContainer(info))
	}

	return cList, nil
}

// listen is called by the Start() method and will begin a loop to listen for
// any msg.Message passed over it's 'in' channel. This channel is the means by
// which the daemon can send signals recieved from user input to the service.
func (ds *DockerService) listen() {
	signal := make(chan struct{})
	go ds.listenDockerAPI(signal)
	go ds.listenHealthChecks()
	defer close(signal)
	for {
		select {
		case m := <-ds.in:
			dm := m.(msg.DockerMessage)
			log.Service.WithField("service", Docker).Tracef("docker listen got %+v", m)
			switch dm.Action {
			case "health":
				log.Service.WithField("service", Docker).Tracef("HEALTH CHECK %+v", )
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
					// TODO (lmower): need to decide what to do on any health status event recieved
					//                which will only happen on state changes if the HEATHCHECK stanza
					//                is provided in a Dockerfile. Below comment is structure for
					//                which fields could be available
					/*
					   "State": {
							... truncated ...
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
					   }
					*/
					log.Service.WithField("service", Docker).Infof("docker action %s\n", dEvent.Action)
					ds.out <- msg.NewDockerMessage(dEvent, nil)
				}
			}
		}
	}
}

func (ds *DockerService) listenHealthChecks() {
	for _, c := range ds.Containers {
		ds.startChecking(c)
	}
}

func (ds *DockerService) startChecking(c *Container)  {
	for _, chk := range c.HealthChecks {
		go ds.startCheckLoop(c, chk)
	}
}

func (ds *DockerService) stopChecking()  {
}

func (ds *DockerService) startCheckLoop(c *Container, chk *health.DockerHealthCheck) {
	tick := time.NewTicker(chk.Formatter.Interval)
	defer tick.Stop()

	for {
		select {
		case <-ds.ctx.Done():
			log.Service.WithField("service", Docker).Tracef("health check context stopped")
			return
		case <-tick.C:
			if chk.Destination == "ex" {
				// TODO - external checks
				log.Service.WithField("service", Docker).Tracef("EXTERN CHECK: %+v", *chk)
			} else {
				log.Service.WithField("service", Docker).Tracef("INTERN CHECK: %+v", *chk)
				res, err := ds.execInternalCheck(c, chk)
				if err != nil {
					log.Service.WithField("service", Docker).Errorf("check %s could not be run: %s", chk.Name, err)
				}
				log.Service.Tracef("RESPONSE: %v", *res)
			}
		}
	}
}

func (ds *DockerService) execInternalCheck(c *Container, chk *health.DockerHealthCheck) (*ExecResult, error) {
	execConf := types.ExecConfig{
		User:         "root",
		Privileged:   true,
		AttachStderr: true,
		AttachStdout: true,
		Env:          nil,
		Cmd:          chk.RawCmd,
	}

	// create the exec instance
	exec, err := ds.Client.ContainerExecCreate(ds.ctx.Ctx, c.ID, execConf)
	if err != nil {
		log.Service.WithField("service", Docker).Tracef("error creating check on container %s: %s", c.ID, err)
		return nil, err
	}

	// execute the command and get hijacked response
	hijack, err := ds.Client.ContainerExecAttach(ds.ctx.Ctx, exec.ID, execConf)
	defer hijack.Close()
	if err != nil {
		log.Service.WithField("service", Docker).Tracef("error attaching exec: %s", err)
		return nil, err
	}

	// use docker api lib to trim prepending bytes from message
	var outBuf, errBuf bytes.Buffer
	stdcopy.StdCopy(&outBuf, &errBuf, hijack.Reader)
	stdout, err := ioutil.ReadAll(&outBuf)
	if err != nil {
		return nil, err
	}
	stderr, err := ioutil.ReadAll(&errBuf)
	if err != nil {
		return nil, err
	}

	// get the exit code from the exec
	res, err := ds.Client.ContainerExecInspect(ds.ctx.Ctx, exec.ID)
	if err != nil {
		log.Service.WithField("service", Docker).Tracef("error inspecting exec: %s", err)
		return nil, err
	}

	return &ExecResult{res.ExitCode, string(stdout), string(stderr)}, nil
}
