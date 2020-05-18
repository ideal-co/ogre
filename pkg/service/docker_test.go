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

var runningID = "09cc8f08b9397f1175058661a16becf417f140da001c738bd44617f42e631f78"
var stoppedID = "fooBarBaz9397f1175058661a16becf417f140da001c738bd44617fooBarBaz"

func getRunningJSON(id string) *types.ContainerJSONBase {
	return &types.ContainerJSONBase{
		ID:   id,
		Name: "ogre-test-container",
		State: &types.ContainerState{
			Status:  "running",
			Running: true,
		},
	}
}

var stoppedContJSONBase = &types.ContainerJSONBase{
	ID:   stoppedID,
	Name: "ogre-test-container",
	State: &types.ContainerState{
		Status:  "down",
		Running: false,
	},
}

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
	var returnVal types.ContainerJSON
	inList := mc.TypeMap["inspect"].([]types.ContainerJSON)
	for _, insp := range inList {
		if insp.ID == container {
			returnVal = insp
		}
	}
	mc.On("ContainerInspect", ctx, container).Return(returnVal, nil)
	args := mc.Mock.Called(ctx, container)
	return args.Get(0).(types.ContainerJSON), args.Error(1)
}

func (mc *MockClient) ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
	var returnList []types.Container
	conts := mc.TypeMap["list"].([]types.Container)
	for _, c := range conts {
		if c.Status == "running" {
			returnList = append(returnList, c)
		}
	}
	mc.On("ContainerList", ctx, options).Return(returnList, nil)
	args := mc.Mock.Called(ctx, options)
	return args.Get(0).([]types.Container), args.Error(1)
}

func (mc *MockClient) ContainerExecAttach(ctx context.Context, execID string, config types.ExecConfig) (types.HijackedResponse, error) {
	resp := mc.TypeMap["exec_attach"].(types.HijackedResponse)
	mc.On("ContainerExecAttach", ctx, execID, config).Return(resp, nil)
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
	tesio := []struct {
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
						ID:     runningID,
						Status: "running",
					},
				},
				"inspect": []types.ContainerJSON{
					{
						ContainerJSONBase: getRunningJSON(runningID),
						Config: &container.Config{
							Hostname: "09cc8f08b939",
							Labels: map[string]string{
								"ogre.health.test.check.default": "echo foo",
							},
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
		{
			name: "should only return running containers with a default health check",
			dsrv: &DockerService{
				ctx: NewDefaultContext(),
			},
			inp: map[string]interface{}{
				"list": []types.Container{
					{
						ID:     runningID,
						Status: "running",
					},
					{
						ID:     stoppedID,
						Status: "stopped",
					},
				},
				"inspect": []types.ContainerJSON{
					{
						ContainerJSONBase: getRunningJSON(runningID),
						Config: &container.Config{
							Hostname: "09cc8f08b939",
							Labels: map[string]string{
								"ogre.health.test.check.default": "echo foo",
							},
						},
					},
					{
						ContainerJSONBase: stoppedContJSONBase,
						Config: &container.Config{
							Hostname: "fooBarBaz939",
							Labels: map[string]string{
								"ogre.health.test.check.default": "echo foo",
							},
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
		{
			name: "should return multiple running containers with a default health check",
			dsrv: &DockerService{
				ctx: NewDefaultContext(),
			},
			inp: map[string]interface{}{
				"list": []types.Container{
					{
						ID:     runningID,
						Status: "running",
					},
					{
						ID:     "fooBarBazID",
						Status: "running",
					},
				},
				"inspect": []types.ContainerJSON{
					{
						ContainerJSONBase: getRunningJSON(runningID),
						Config: &container.Config{
							Hostname: "09cc8f08b939",
							Labels: map[string]string{
								"ogre.health.test.check.default": "echo foo",
							},
						},
					},
					{
						ContainerJSONBase: getRunningJSON("fooBarBazID"),
						Config: &container.Config{
							Hostname: "fooBar",
							Labels: map[string]string{
								"ogre.health.test.check.two.default": "echo foo",
							},
						},
					},
				},
			},
			test: func(ds *DockerService, args map[string]interface{}) ([]*Container, error) {
				ds.Client = NewMockClient(args)
				return ds.collectContainers()
			},
			cont: 2,
		},
		{
			name: "should return multiple running containers with different backends for health checks",
			dsrv: &DockerService{
				ctx: NewDefaultContext(),
			},
			inp: map[string]interface{}{
				"list": []types.Container{
					{
						ID:     runningID,
						Status: "running",
					},
					{
						ID:     "fooBarBazID",
						Status: "running",
					},
				},
				"inspect": []types.ContainerJSON{
					{
						ContainerJSONBase: getRunningJSON(runningID),
						Config: &container.Config{
							Hostname: "09cc8f08b939",
							Labels: map[string]string{
								"ogre.health.test.check.default": "echo foo",
							},
						},
					},
					{
						ContainerJSONBase: getRunningJSON("fooBarBazID"),
						Config: &container.Config{
							Hostname: "fooBar",
							Labels: map[string]string{
								"ogre.health.in.test.check.two": "echo foo",
								"ogre.format.backend.statsd":    "true",
							},
						},
					},
				},
			},
			test: func(ds *DockerService, args map[string]interface{}) ([]*Container, error) {
				ds.Client = NewMockClient(args)
				return ds.collectContainers()
			},
			cont: 2,
		},
	}
	for _, io := range tesio {
		t.Run(io.name, func(t *testing.T) {
			containers, err := io.test(io.dsrv, io.inp)
			assert.Nil(t, err, "error was not nil %s", err)
			assert.Len(t, containers, io.cont, "had %d checks, expected %d", len(containers), io.cont)

			for idx, c := range containers {
				name := io.inp["inspect"].([]types.ContainerJSON)[idx].Name
				id := io.inp["inspect"].([]types.ContainerJSON)[idx].ID

				assert.NotEmptyf(t, c.HealthChecks, "container health checks empty %v", *c)
				assert.Equal(t, c.Name, name, "names did not match - had: %s, expect %s", c.Name, name)
				assert.Equal(t, c.ID, id, "ids did not match - had: %s, expect %s", c.ID, id)
			}
		})
	}
}

/*
TODO (lmower): there are some issues with testing the ContainerExecAttach interface
               and returning a HijackedResponse. This needs a bit more tooling to
               be truly useful and the Docker API lib sadly doesn't even test this
               particular interface so we lack some direction. See the linked issue
               for tracking this work: https://github.com/ideal-co/ogre/issues/14

var pingResp = `PING 127.0.0.1 (127.0.0.1) 56(84) bytes of data.
64 bytes from 127.0.0.1: icmp_seq=1 ttl=64 time=0.024 ms

--- 127.0.0.1 ping statistics ---
1 packets transmitted, 1 received, 0% packet loss, time 0ms
rtt min/avg/max/mdev = 0.024/0.024/0.024/0.000 ms`

func TestExecInternalCheck(t *testing.T) {
	tesio := []struct {
		name string
		dsrv *DockerService
		inp  map[string]interface{}
		test func(ds *DockerService, args map[string]interface{}, cmd []string) (health.ExecResult, error)
		cmd  []string
		exp  health.ExecResult
	}{
		{
			name: "should return a non nil exec result",
			dsrv: &DockerService{
				ctx: NewDefaultContext(),
			},
			inp: map[string]interface{}{
				"exec_create": types.IDResponse{ID: "fooBarExecID"},
				"exec_attach": types.HijackedResponse{

				},
				"exec_inspect": types.ContainerExecInspect{
					ExecID:      "fooBarExecID",
					ContainerID: runningID,
					ExitCode:    0,
				},
			},
			test: func(ds *DockerService, args map[string]interface{}, cmd []string) (health.ExecResult, error) {
				ds.Client = NewMockClient(args)
				return ds.execInternalCheck(context.Background(), runningID, cmd)
			},
			exp: health.ExecResult{
				Exit:   0,
				StdOut: pingResp,
				StdErr: "",
			},
		},
	}

	for _, io := range tesio {
		t.Run(io.name, func(t *testing.T) {
			result, err := io.test(io.dsrv, io.inp, io.cmd)
			assert.Nil(t, err, "err from execInternalCheck was not nil")
			//assert.Equal(t, result.StdOut, io.exp.StdOut, "result out did not match had: %s expected %s", result.StdOut, io.exp.StdOut)
			assert.Equal(t, result.Exit, io.exp.Exit)
		})
	}
}
*/
