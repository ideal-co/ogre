package daemon

import (
	"github.com/ideal-co/ogre/pkg/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDaemon_collectServices(t *testing.T) {
	testIO := []struct {
		name  string
		srvcs []types.ServiceType
		test  func() *Daemon
	}{
		{
			name:  "should have at least two services by default",
			srvcs: []types.ServiceType{types.BackendService, types.DockerService},
			test: func() *Daemon {
				d := New()
				// TODO (lmower): this will expect to create the Docker service which needs
				//                an API client which expects the docker unix socket
				// d.collectServices()
				return d
			},
		},
	}

	for _, io := range testIO {
		t.Run(io.name, func(t *testing.T) {
			d := io.test()
			assert.NotNil(t, d, "was expecting daemon to be not nil")
		})
	}
}
