package srvc

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

//var contJSON = types.ContainerJSON{
//	Config: &container.Config{
//		Hostname:  "09cc8f08b939",
//		Tty:       true,
//		OpenStdin: true,
//		Env: []string{
//			"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
//		},
//		Image:      "test:0.0.1",
//		Entrypoint: []string{"nc", "-lke", "127.0.0.1", "8000"},
//		OnBuild:    nil,
//		Labels: map[string]string{
//			"ogre.health.ex.test.check.script": "./usr/bin/healthcheck.sh",
//			"ogre.health.in.test.check.ping":   "ping -c 1 -W 1 172.17.0.3",
//			"ogre.health.test.check.default":   "echo foo",
//		},
//	},
//}

type MockClient struct {
	mock.Mock
	TypeMap map[string]interface{}
}

func NewMockClient(args map[string]interface{}) *MockClient {
	return &MockClient{
		TypeMap: args,
	}
}

func (mc *MockClient) Events(ctx context.Context, options types.EventsOptions) (<-chan events.Message, <-chan error) {
	args := mc.Mock.Called(ctx, options)
	return args.Get(0).(chan events.Message), args.Get(1).(chan error)
}

func (mc *MockClient) ContainerInspect(ctx context.Context, container string) (types.ContainerJSON, error) {
	mc.On("ContainerInspect", ctx, container).Return(mc.TypeMap["inspect"], nil)
	args := mc.Mock.Called(ctx, container)
	return args.Get(0).(types.ContainerJSON), args.Error(1)
}

func (mc *MockClient) ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
	mc.On("ContainerList", ctx, options).Return(mc.TypeMap["list"], nil)
	args := mc.Mock.Called(ctx, options)
	return args.Get(0).([]types.Container), args.Error(1)
}

func (mc *MockClient) ContainerExecAttach(ctx context.Context, execID string, config types.ExecConfig) (types.HijackedResponse, error) {
	mc.On("ContainerExecAttach", ctx, execID, config).Return(mc.TypeMap["exec_attach"], nil)
	args := mc.Mock.Called(ctx, execID, config)
	return args.Get(0).(types.HijackedResponse), args.Error(1)
}

func (mc *MockClient) ContainerExecCreate(ctx context.Context, container string, config types.ExecConfig) (types.IDResponse, error) {
	mc.On("ContainerExecCreate", ctx, container, config).Return(mc.TypeMap["exec_create"], nil)
	args := mc.Mock.Called(ctx, container, config)
	return args.Get(0).(types.IDResponse), args.Error(1)
}

func (mc *MockClient) ContainerExecInspect(ctx context.Context, execID string) (types.ContainerExecInspect, error) {
	mc.On("ContainerExecInspect", ctx, execID).Return(mc.TypeMap["exec_inspect"], nil)
	args := mc.Mock.Called(ctx, execID)
	return args.Get(0).(types.ContainerExecInspect), args.Error(1)
}

func TestCollectContainers(t *testing.T) {
	testIO := []struct {
		name string
		dsrv *DockerService
		inp  map[string]interface{}
		test func(ds *DockerService, args map[string]interface{}) ([]*Container, error)
		cont int
	}{
		{
			name: "should return a single running container with a default health check",
			dsrv: &DockerService{
				ctx: NewDefaultContext(),
			},
			inp: map[string]interface{}{
				"list": []types.Container{
					{
						ID: "09cc8f08b9397f1175058661a16becf417f140da001c738bd44617f42e631f78",
					},
				},
				"inspect": types.ContainerJSON{
					ContainerJSONBase: &types.ContainerJSONBase{
						ID:    "09cc8f08b9397f1175058661a16becf417f140da001c738bd44617f42e631f78",
						Image: "ogre-test:latest",
						Name:  "ogre-test-container",
						State: &types.ContainerState{
							Status:     "running",
							Running:    true,
						},
					},
					Config: &container.Config{
						Hostname: "09cc8f08b939",
						Labels: map[string]string{
							"ogre.health.test.check.default": "echo foo",
						},
					},
				},
			},
			test: func(ds *DockerService, args map[string]interface{}) ([]*Container, error) {
				ds.Client = NewMockClient(args)
				return ds.collectContainers()
			},
			cont: 1,
		},
	}
	for _, io := range testIO {
		t.Run(io.name, func(t *testing.T) {
			containers, err := io.test(io.dsrv, io.inp)
			assert.Nil(t, err, "error was not nil %s", err)
			assert.Len(t, containers, io.cont,"had %d checks, expected %d", len(containers), io.cont)

			for _, c := range containers {
				labels := io.inp["inspect"].(types.ContainerJSON).Config.Labels
				name := io.inp["inspect"].(types.ContainerJSON).Name
				id := io.inp["inspect"].(types.ContainerJSON).ID

				assert.NotEmptyf(t, c.HealthChecks, "container health checks empty %v", *c)
				assert.Len(t, c.HealthChecks, len(labels),"had %d checks, expected %d", len(c.HealthChecks), len(labels))
				assert.Equal(t, c.Name, name,"names did not match - had: %s, expect %s", c.Name, name)
				assert.Equal(t, c.ID, id,"ids did not match - had: %s, expect %s", c.ID, id)
			}
		})
	}
}
