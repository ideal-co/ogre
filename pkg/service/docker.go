package srvc

import (
	"bytes"
	"context"
	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/lowellmower/ogre/pkg/health"
	"github.com/lowellmower/ogre/pkg/log"
	msg "github.com/lowellmower/ogre/pkg/message"
	internalTypes "github.com/lowellmower/ogre/pkg/types"
	"io/ioutil"
	"os/exec"
	"time"
)

// DockerAPIClient is an interface which wraps a subset of a number of the
// Docker API interfaces. By limiting our client to only the interfaces we
// use, we are able to reduce code footprint and more easily test.
type DockerAPIClient interface {
	// subset of the ContainerAPIClient interface
	ContainerInspect(ctx context.Context, container string) (dockerTypes.ContainerJSON, error)
	ContainerList(ctx context.Context, options dockerTypes.ContainerListOptions) ([]dockerTypes.Container, error)

	// subset of the SystemAPIClient interface
	Events(ctx context.Context, options dockerTypes.EventsOptions) (<-chan events.Message, <-chan error)
	//Info(ctx context.Context) (dockerTypes.Info, error)

	// https://pkg.go.dev/github.com/docker/docker/client@v1.13.1?tab=doc#ContainerAPIClient
	//ContainerAttach(ctx context.Context, container string, options dockerTypes.ContainerAttachOptions) (dockerTypes.HijackedResponse, error)
	ContainerExecAttach(ctx context.Context, execID string, config dockerTypes.ExecConfig) (dockerTypes.HijackedResponse, error)
	ContainerExecCreate(ctx context.Context, container string, config dockerTypes.ExecConfig) (dockerTypes.IDResponse, error)
	ContainerExecInspect(ctx context.Context, execID string) (dockerTypes.ContainerExecInspect, error)
	//ContainerExecResize(ctx context.Context, execID string, options dockerTypes.ResizeOptions) error
	//ContainerExecStart(ctx context.Context, execID string, config dockerTypes.ExecStartCheck) error
}

// DockerService satisfies the srvc.Service interface and is the struct for
// coordinating the Docker related services. DockerService communicates with
// the Daemon (main process) by means of a channel of msg.Message, which is
// its 'out' field. This channel is also the 'Daemon.In' field where other
// services and user input communicate with the Daemon.
type DockerService struct {
	Client        DockerAPIClient
	Containers    []*Container
	RunningChecks map[string]context.CancelFunc

	ctx *Context
	in  chan msg.Message
	out chan msg.Message
	err chan msg.Message
}

// Container is the struct representation of a Docker container on the host the
// docker daemon ogre is communicating with is running. It encapsulates the list
// of HealthChecks to execute.
type Container struct {
	Name string
	ID   string
	ctx  *Context

	Info         dockerTypes.ContainerJSON
	HealthChecks []*health.DockerHealthCheck
}

// Type satisfies the Service.Type interface and returns a ServiceType of Docker
func (ds *DockerService) Type() internalTypes.ServiceType {
	return internalTypes.DockerService
}

// Start is the DockerService implementation of the Service interface's Start
// function. It calls the private method listen() which will start a loop to
// listen for signals from the daemon as well as spin off a go routine to begin
// listening on the Docker API for Events. Start will never return an error for
// DockerService as errors from this service are reported to the daemon by way
// of a channel of msg.Message which the Err() implementation returns a non-nil
// value.
func (ds *DockerService) Start() error {
	log.Service.Infof("starting %s service", ds.Type())
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
	log.Service.Infof("stopping %s service", ds.Type())
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
		Client:        dockerClient,
		RunningChecks: make(map[string]context.CancelFunc),
		ctx:           NewDefaultContext(),
		in:            in,
		out:           out,
		err:           errChan,
	}

	ds.Containers, err = ds.collectContainers()
	if err != nil {
		return nil, err
	}

	return ds, nil
}

// NewContainer takes the ContainerJSON result from a call to the Docker API
// for inspect and returns a pointer to a container with the any applicable
// labels from the info param parsed into DockerHealthChecks. If no applicable
// checks were found, i.e. there were no labels prefixed with 'ogre.health'
// than an empty list is returned from NewDockerHealthCheck and a nil value
// will be returned from NewContainer.
func NewContainer(info dockerTypes.ContainerJSON) *Container {
	// parse label info
	heathChecks := health.NewDockerHealthCheck(info.Config.Labels)
	if len(heathChecks) == 0 {
		return nil
	}

	// instantiate container
	c := &Container{Info: info, HealthChecks: heathChecks}
	c.ID = info.ID
	c.Name = info.Name
	c.ctx = NewDefaultContext()

	return c
}

// collectContainers will return a slice of pointers of type Container and an
// error which the latter will be nil upon success. The slice will be empty
// (of len 0) if there were no running containers or there were no containers
// which had at least one label prefixed with 'ogre.health'. Only containers
// which are running and have a configured health check will be returned from
// collect containers
func (ds *DockerService) collectContainers() ([]*Container, error) {
	var cList []*Container
	arg, _ := filters.FromParam("status=running")
	opts := dockerTypes.ContainerListOptions{Filters: arg}

	// gather running containers
	containers, err := ds.Client.ContainerList(ds.ctx.Ctx, opts)
	if err != nil {
		return nil, err
	}

	// gather info on containers
	for _, c := range containers {
		newCont, err := ds.getContainerFromInfo(c.ID)
		if err != nil {
			log.Service.WithField("service", internalTypes.DockerService).Errorf("could not get info for %s: %s", c.ID, err)
			continue
		}

		cList = append(cList, newCont)
	}

	return cList, nil
}

func (ds *DockerService) getContainerFromInfo(cid string) (*Container, error) {
	info, err := ds.Client.ContainerInspect(ds.ctx.Ctx, cid)
	if err != nil {
		return nil, err
	}

	if newCont := NewContainer(info); newCont != nil {
		return newCont, nil
	}

	return nil, internalTypes.ErrNoCheck
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
			log.Service.WithField("service", internalTypes.DockerService).Tracef("docker listen got %+v", m)
			switch dm.Action {
			case "start-health":
				cont, err := ds.getContainerFromInfo(dm.Actor.ID)
				if err != nil {
					log.Service.WithField("service", internalTypes.DockerService).Errorf("could not get check on container start: %s", err)
					continue
				}

				ds.RunningChecks[cont.ID] = cont.ctx.Cancel
				go ds.startChecking(cont)
			case "stop-health":
				ds.stopContainerChecking(dm.Actor.ID)
			case "stop":
				ds.stopAllChecking()
				signal <- struct{}{}
				ds.ctx = NewDefaultContext()
			case "start":
				go ds.listenDockerAPI(signal)
				containers, err := ds.collectContainers()
				if err != nil {
					log.Service.WithField("service", internalTypes.DockerService).Errorf("error getting containers on start: %s", err)
					continue
				}
				ds.Containers = containers
				go ds.listenHealthChecks()
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
	dockerEvents, errChan := ds.Client.Events(ds.ctx.Ctx, dockerTypes.EventsOptions{})
	for {
		select {
		case <-signal:
			// NOTE: we may also need/want to do something with the
			//       ds.ctx.Ctx passed to ds.Client.Events above but
			//       this is also the service level context so we must
			//       not cancel it. Simply returning for now.
			log.Service.WithField("service", internalTypes.DockerService).Trace("stopping container listener...")
			return
		case err := <-errChan:
			log.Service.WithField("service", internalTypes.DockerService).Tracef("err in listen %s\n", err)
			ds.err <- msg.NewDockerMessage(events.Message{}, err.Error())
		case dEvent := <-dockerEvents:
			switch dEvent.Type {
			// Container Events from Docker API
			case events.ContainerEventType:
				switch dEvent.Action {
				case "start":
					ds.out <- msg.NewDockerMessage(dEvent, "start-health")
				case "restart":
					// TODO (lmower): decide if there is action we want to take here, possibly track
					//                flapping containers or restarts over a period of time?
					log.Service.WithField("service", internalTypes.DockerService).Infof("docker action %s\n", dEvent.Action)
				case "stop":
					log.Service.WithField("service", internalTypes.DockerService).Infof("docker action %s\n", dEvent.Action)
					ds.out <- msg.NewDockerMessage(dEvent, "stop-health")
				case "die":
					log.Service.WithField("service", internalTypes.DockerService).Infof("docker action %s\n", dEvent.Action)
					ds.out <- msg.NewDockerMessage(dEvent, "stop-health")
				// introduced in docker v1.12 (2016)
				case "health_status":
					// TODO (lmower): need to decide what to do on any health status event recieved
					//                which will only happen on state changes if the HEATHCHECK stanza
					//                is provided in a Dockerfile. Below comment is structure for
					//                which fields could be available
					//                Issue: https://github.com/ideal-co/ogre/issues/12
					log.Service.WithField("service", internalTypes.DockerService).Infof("docker action %s\n", dEvent.Action)
					ds.out <- msg.NewDockerMessage(dEvent, dEvent.Action)
				}
			}
		}
	}
}

// listenHealthChecks iterates over the DockerService's containers and kicks
// off the listening loop for all health checks for each container.
func (ds *DockerService) listenHealthChecks() {
	for _, c := range ds.Containers {
		if _, ok := ds.RunningChecks[c.ID]; !ok {
			ds.RunningChecks[c.ID] = c.ctx.Cancel
			ds.startChecking(c)
		}
	}
}

// startChecking takes a pointer to a container and kicks of a go routine for
// each health check associated with that container.
func (ds *DockerService) startChecking(c *Container) {
	for _, chk := range c.HealthChecks {
		go ds.startCheckLoop(c, chk)
	}
}

// stopContainerChecking takes a string representing a container ID and stops
// a particular containers health checks by means of the associated context's
// cancel function.
func (ds *DockerService) stopContainerChecking(cid string) {
	if cancel, ok := ds.RunningChecks[cid]; ok {
		cancel()
	}
}

// stopAllChecking stops all running health checks by means of the associated
// context's cancel function.
func (ds *DockerService) stopAllChecking() {
	for cid, cancel := range ds.RunningChecks {
		log.Service.WithField("service", internalTypes.DockerService).Infof("stopping check for %s", cid)
		cancel()
		delete(ds.RunningChecks, cid)
	}
}

// startCheckLoop takes a pointer to a Container and a DockerHealthCheck and
// begins an infinite loop where the check is executed within or against the
// container at an interval configured for the check. This loop can only be
// interrupted by the DockerService's context being canceled, the Container's
// context being canceled, or the ogre daemon process being killed or issued
// an interrupt. When a health check is executed on the interval, the completed
// check is then sent as a msg.Message to the ogre daemon to be routed to the
// appropriate reporting backend.
func (ds *DockerService) startCheckLoop(c *Container, chk *health.DockerHealthCheck) {
	tick := time.NewTicker(chk.Interval)
	defer tick.Stop()

	for {
		select {
		case <-ds.ctx.Done():
			log.Service.WithField("service", internalTypes.DockerService).Tracef("service stopped, stopping checks for container %s", c.Name)
			return
		case <-c.ctx.Done():
			log.Service.WithField("service", internalTypes.DockerService).Tracef("stopping container checks for %s", c.Name)
			tick.Stop()
			return
		case <-tick.C:
			if chk.Destination == "ex" {
				result, err := ds.execExternalCheck(chk)
				if err != nil {
					log.Service.WithField("service", internalTypes.DockerService).Errorf("check %s could not be run: %s", chk.Name, err)
					continue
				}
				log.Service.WithField("service", internalTypes.DockerService).Tracef("EXTERN CHECK: %+v", chk)
				chk.Result = result
				ds.out <- msg.NewBackendMessage(chk, chk.Formatter.Platform.Target)
			} else {
				result, err := ds.execInternalCheck(c.ctx.Ctx, c.ID, chk.Cmd.Args)
				if err != nil {
					log.Service.WithField("service", internalTypes.DockerService).Errorf("check %s could not be run: %s", chk.Name, err)
					continue
				}
				chk.Result = result
				ds.out <- msg.NewBackendMessage(chk, chk.Formatter.Platform.Target)
			}
		}
	}
}

func (ds *DockerService) execExternalCheck(chk *health.DockerHealthCheck) (health.ExecResult, error) {
	var result health.ExecResult
	// make a copy of the command to reset after exec
	var copyCmd exec.Cmd
	copyCmd = *chk.Cmd
	defer func() {
		chk.Cmd = &copyCmd
	}()

	var outBuf, errBuf bytes.Buffer
	chk.Cmd.Stdout = &outBuf
	chk.Cmd.Stderr = &errBuf

	err := chk.Cmd.Start()
	if err != nil {
		return result, err
	}

	err = chk.Cmd.Wait()
	if err != nil {
		return result, err
	}

	stdout, err := ioutil.ReadAll(&outBuf)
	if err != nil {
		return result, err
	}

	stderr, err := ioutil.ReadAll(&errBuf)
	if err != nil {
		return result, err
	}

	result.Exit = chk.Cmd.ProcessState.ExitCode()
	result.StdOut = string(stdout)
	result.StdErr = string(stderr)

	return result, nil
}

// execInternalCheck takes a context.Context associated with a Container, a
// string representing the container's ID, and a slice of string which is the
// command associated with the DockerHealthCheck for that container. The method
// returns an ExecResult and an error, the latter of which will be nil upon
// successful execution of the health check. The command associated with the
// DockerHealthCheck will be executed inside of the container and the result of
// that command (exit code, stdout, stderr) will be passed to the ExecResult
// and stored on the DockerHealthCheck struct for reporting to the backend.
func (ds *DockerService) execInternalCheck(ctx context.Context, cid string, cmd []string) (health.ExecResult, error) {
	var result health.ExecResult
	execConf := dockerTypes.ExecConfig{
		User:         "root",
		Privileged:   true,
		AttachStderr: true,
		AttachStdout: true,
		Env:          nil,
		Cmd:          cmd,
	}

	// create the exec instance
	exec, err := ds.Client.ContainerExecCreate(ctx, cid, execConf)
	if err != nil {
		log.Service.WithField("service", internalTypes.DockerService).Tracef("error creating check on container %s: %s", cid, err)
		return result, err
	}
	// execute the command and get hijacked response
	hijack, err := ds.Client.ContainerExecAttach(ctx, exec.ID, execConf)
	if err != nil {
		log.Service.WithField("service", internalTypes.DockerService).Tracef("error attaching exec: %s", err)
		return result, err
	}
	defer hijack.Close()

	// use docker api lib to trim prepending bytes from message
	var outBuf, errBuf bytes.Buffer
	if _, e := stdcopy.StdCopy(&outBuf, &errBuf, hijack.Reader); e != nil {
		log.Service.Errorf("error copying check output: %s", e)
	}

	stdout, err := ioutil.ReadAll(&outBuf)
	if err != nil {
		return result, err
	}
	stderr, err := ioutil.ReadAll(&errBuf)
	if err != nil {
		return result, err
	}

	// get the exit code from the exec
	res, err := ds.Client.ContainerExecInspect(ctx, exec.ID)
	if err != nil {
		log.Service.WithField("service", internalTypes.DockerService).Tracef("error inspecting exec: %s", err)
		return result, err
	}

	result.Exit = res.ExitCode
	result.StdOut = string(stdout)
	result.StdErr = string(stderr)

	return result, nil
}
