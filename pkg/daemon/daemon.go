package daemon

import (
	"bytes"
	"fmt"
	"github.com/ideal-co/ogre/pkg/backend"
	"github.com/ideal-co/ogre/pkg/config"
	"github.com/ideal-co/ogre/pkg/log"
	msg "github.com/ideal-co/ogre/pkg/message"
	srvc "github.com/ideal-co/ogre/pkg/service"
	"github.com/ideal-co/ogre/pkg/types"
	"io"
	"net"
	"os"
)

const (
	OgredSocket  = "ogred_socket"
	OgredPIDFile = "ogred_pid"
)

// Daemon this the top level process for communications into the information
// gathering mechanisms in Ogre. The daemon will handle user input from the
// CLI and disseminate the message out to the appropriate sub-process. Each
// sub-process will maintain an in-box and out-box for the purposes of sending
// and receiving information to and from the daemon.
type Daemon struct {
	In  chan msg.Message
	Out map[types.MessageType]chan msg.Message
	Err chan msg.Message

	ctx      *srvc.Context
	services map[types.ServiceType]srvc.Service
	listener net.Listener
}

// Run is the main entry point for the ogre daemon. It will establish the
// necessary configuration for the listening process based on the environment
// in which it is running as well as any user applied or default configuration
func Run() {
	d := New()
	d.collectServices()
	d.establishClients()
	go d.runServices()
	go d.listenChannel()
	d.ListenSocket()
}

// New returns a pointer to a new instance of Daemon struct
func New() *Daemon {
	return &Daemon{
		In:       make(chan msg.Message),
		Out:      make(map[types.MessageType]chan msg.Message),
		Err:      make(chan msg.Message),
		ctx:      srvc.NewDefaultContext(),
		services: make(map[types.ServiceType]srvc.Service),
	}

}

// ListenSocket starts the main daemon listener on the unix socket. When the
// listener accepts a connection, that connection is passed in a go routine to
// the handleMessage method.
func (d *Daemon) ListenSocket() {
	ogredSock := config.Daemon.GetString(OgredSocket)
	ogredPID := config.Daemon.GetString(OgredPIDFile)

	// ensure FD doesn't already exist for socket
	if err := os.RemoveAll(ogredSock); err != nil {
		log.Daemon.Infof("no socket file collision detected for %s: %s", ogredSock, err)
	}

	daemon, err := net.Listen("unix", ogredSock)
	if err != nil {
		log.Daemon.Fatalf("could not start listener on %s: %s", ogredSock, err)
	}

	// clean up files
	defer daemon.Close()
	defer os.RemoveAll(ogredSock)
	defer os.RemoveAll(ogredPID)

	d.listener = daemon
	for {
		select {
		case <-d.ctx.Done():
			return
		default:
			conn, err := d.listener.Accept()
			if err != nil {
				log.Fatal(err)
			}

			go d.handleMessage(conn)
		}
	}
}

// Deserialize takes a slice of bytes and returns a msg.Message interface and
// and error. The msg.Message returned will be of the GenericMessage type which
// is the message type responsible for interfacing with user input from the CLI.
func (d *Daemon) Deserialize(data []byte) (msg.Message, error) {
	var gm msg.GenericMessage
	return gm.Deserialize(data)
}

// runServices calls the Start method for all the services configured on the
// daemon and stored in the services field.
func (d *Daemon) runServices() {
	for _, srv := range d.services {
		go func(s srvc.Service) {
			if err := s.Start(); err != nil {
				log.Daemon.Errorf("could not start service %s", s.Type())
			}
		}(srv)
	}
}

// stopServices calls the Stop method for all the services configured on the
// daemon and stored in the services field.
func (d *Daemon) stopServices() {
	for _, s := range d.services {
		if err := s.Stop(); err != nil {
			log.Daemon.Errorf("encountered error stopping service %s: %s", s.Type(), err)
		}
	}
}

// handleMessage takes a net.Conn which is passed from the unix socket accepting
// a new connection and copies the data into a buffer. The bytes in that buffer
// are then deserialized into a msg.Message of type GenericMessage and send over
// the daemon's 'In' field which is a channel of msg.Message. From there, the
// daemon will properly route the message to the appropriate service, or if it
// was a message destined for the daemon, induce some desired behavior.
func (d *Daemon) handleMessage(c net.Conn) {
	defer c.Close()
	var buf bytes.Buffer

	if _, err := io.Copy(&buf, c); err != nil {
		log.Daemon.Errorf("error copying message: %s", err)
	}

	m, err := d.Deserialize(buf.Bytes())
	if err != nil {
		log.Daemon.Errorf("encountered error deserializing message: %s", err)
		return
	}

	d.In <- m
}

// listenChannel is an infinite loop where the daemon waits for signals over the
// error channel, or the  in channel. An associated context is also listening
// for the execution of the context's CancelFunc.
func (d *Daemon) listenChannel() {
	for {
		select {
		case <-d.ctx.Done():
			d.stopServices()
			// TODO (lmower): determine if we want to do anything with the ctx.Err
			//                and ctx.Callback or if these are unnecessary
			//d.Err <- d.ctx.Err
			//d.Err <- d.ctx.Callback
			return
		case eMsg := <-d.Err:
			// this error channel should be used to signal terminal errors which result in
			// more drastic action from the daemon like attempted restarts or shutdowns
			// other less impacting errors should be sent over the Daemon.In channel and
			// the Message interface Err() should be checked for a nil value
			log.Daemon.Fatalf("daemon received fatal error from %s: %s", eMsg.Type(), eMsg.Error())
		case m := <-d.In:
			// check if a non-fatal error was sent
			if m.Error() != nil {
				log.Daemon.Errorf("daemon received error from %s service: %s", m.Type(), m.Error())
				continue
			}

			d.directIncomingMsg(m)
		}
	}
}

// collectServices sets the services field on the daemon with respect to the
// configured services which should be run. Should there be an error in setting
// this field, the process should exit. At the moment, the only service to be
// configured is the Docker service, others will be added as the project grows
func (d *Daemon) collectServices() {
	// our default services which will always be made
	srvMap := map[types.ServiceType]types.MessageType{
		types.DockerService:  types.DockerMessage,
		types.BackendService: types.BackendMessage,
	}

	for s, m := range srvMap {
		out := make(chan msg.Message)
		service, err := srvc.NewService(s, d.In, out, d.Err)
		if err != nil {
			log.Daemon.Fatalf("could not establish services: %s", err)
		}
		d.services[s] = service
		d.Out[m] = out
	}
}

// establishClients looks to the configuration parsed at start to see if there
// are any backend platforms to configure clients for. Should no configuration
// be passed, a default backend is always established where the output is in
// JSON to the log file of the daemon process.
func (d *Daemon) establishClients() {
	bes := d.services[types.BackendService].(*srvc.BackendService)

	// check to see if there were user provided backends
	if config.DaemonConf.Backends != nil {
		for _, bEnd := range config.DaemonConf.Backends {
			platform, err := backend.NewBackendClient(types.PlatformType(bEnd.Type), bEnd)
			if err != nil {
				fmt.Printf("ogred not started check daemon log at %s\n", config.DaemonConf.Log.File)
				log.Daemon.Fatalf("could not get backend %s: %s", bEnd.Type, err)
			}
			bes.Platforms[platform.Type()] = platform
		}
	}

	// always set up the default backend (log)
	platform, err := backend.NewBackendClient(types.DefaultBackend, config.BackendConfig{})
	if err != nil {
		log.Daemon.Fatalf("cannot start daemon %s", err)
	}
	bes.Platforms[types.DefaultBackend] = platform
}

// directIncomingMsg takes a message and pushes it over the corresponding
// channel for the MessageType. This is used by the daemon to direct messages
// to services and backends. If there is no channel for that message type it
// is presumed that the message is meant for the daemon itself.
func (d *Daemon) directIncomingMsg(m msg.Message) {
	log.Daemon.Tracef("in directIncoming %+v", m)

	// if it is a message destined for a service, send it over the
	// corresponding channel
	if ch, ok := d.Out[m.Type()]; ok {
		ch <- m
		return
	}

	// otherwise, it is a message meant for the daemon
	d.handleDaemonMsg(m)
}

// handleDaemonMsg takes a msg.Message expecting it to be of type DaemonMessage
// where it will process the action desired accordingly.
func (d *Daemon) handleDaemonMsg(m msg.Message) {
	if m.(msg.DaemonMessage).Action == "stop" {
		log.Daemon.Trace("stopping ogre daemon...")
		d.ctx.Cancel()
	}
}
