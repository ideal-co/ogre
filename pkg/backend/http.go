package backend

import (
	"bytes"
	"fmt"
	"github.com/lowellmower/ogre/pkg/log"
	msg "github.com/lowellmower/ogre/pkg/message"
	"github.com/lowellmower/ogre/pkg/types"
	"net/http"
	"net/url"
)

// HTTPBackend satisfies the Platform interface and is responsible for sending
// health check results to an arbitrary HTTP endpoint capable of handling POST
// requests.
type HTTPBackend struct {
	Client *http.Client
	URL    *url.URL
	Format string
}

// NewHTTPBackend takes three strings, a server which is the address of either
// a local or remote HTTP server. A path, representing the resource path by
// which to address requests. And format, which can be used to set the content
// type of the request. At the moment, the only supported content type is in
// the form of application/json and is hard coded in the request. And error
func NewHTTPBackend(server, path, format string) (Platform, error) {
	hb := &HTTPBackend{
		Client: http.DefaultClient,
		Format: format,
	}

	// ensure the url is acceptable and can be parsed
	addr, err := url.Parse("http://" + server + path)
	if err != nil {
		return nil, err
	}
	hb.URL = addr

	return hb, nil
}

// Send HTTPBackend's implementation of the Platform interface Send method. Send
// takes a Message, serializes it and makes a POST request to the configured HTTP
// backend passed in the ogred config file. An error is returned if the Message
// cannot be serialized or sent.
func (hb *HTTPBackend) Send(m msg.Message) error {
	bem := m.(msg.BackendMessage)
	data, err := bem.Serialize()
	if err != nil {
		return fmt.Errorf("could not serialize message for http: %s", err)
	}
	resp, err := hb.Client.Post(hb.URL.String(), "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("could not send message for http: %s", err)
	}

	log.Daemon.Tracef("HTTP backend response %v", resp)
	return nil
}

// Type is the HTTPBackend implementation of the Platform interface Type
// and returns a PlatformType of type HTTPBackend.
func (hb *HTTPBackend) Type() types.PlatformType {
	return types.HTTPBackend
}
