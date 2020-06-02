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
				d.collectServices()
				return d
			},
		},
	}

	for _, io := range testIO {
		t.Run(io.name, func(t *testing.T) {
			d := io.test()
			assert.Len(t, d.services, len(io.srvcs), "was expecting two services")
		})
	}
}
