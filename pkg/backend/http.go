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

type HTTPBackend struct {
	Client *http.Client
	URL    *url.URL
	Format string
}

func NewHTTPBackend(server, path, format string) (Platform, error) {
	hb := &HTTPBackend{
		Client: http.DefaultClient,
		Format: format,
	}
	if addr, err := url.Parse("http://" + server + path); err != nil {
		return nil, err
		hb.URL = addr
	}

	return hb, nil
}

func (hb *HTTPBackend) Send(m msg.Message) error {
	bem := m.(msg.BackendMessage)
	data, err := bem.Serialize()
	if err != nil {
		return fmt.Errorf("could not serialize message for http: %s", err)
	}
	resp, err := hb.Client.Post("http://127.0.0.1:9009", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("could not send message for http: %s", err)
	}

	log.Daemon.Tracef("response %v", resp)
	return nil
}

func (hb *HTTPBackend) Type() types.PlatformType {
	return types.HTTPBackend
}
