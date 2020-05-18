package daemon

import (
    "github.com/lowellmower/ogre/pkg/types"
    "github.com/stretchr/testify/assert"
    "testing"
)

func TestDaemon_collectServices(t *testing.T) {
    testIO := []struct {
        name string
        inp *Daemon
        srvcs []types.ServiceType
        test func(d *Daemon) *Daemon
    }{
        {
            name: "should have at least two services by default",
            inp: New(),
            srvcs: []types.ServiceType{types.BackendService, types.DockerService},
            test: func(d *Daemon) *Daemon {
                d.collectServices()
                return d
            },
        },
    }

    for _, io := range testIO {
        t.Run(io.name, func(t *testing.T) {
            d := io.test(io.inp)
            for _, s := range io.srvcs {
                assert.NotNil(t, d.services[s], "was expecting service %s", s)
            }
        })
    }
}
