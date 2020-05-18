package srvc

import (
	"context"
	"github.com/lowellmower/ogre/pkg/backend"
	"github.com/lowellmower/ogre/pkg/health"
	msg "github.com/lowellmower/ogre/pkg/message"
	"github.com/lowellmower/ogre/pkg/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

type MockCompletedHC struct {
	Result string
	Exit   int
	Pass   bool
}

func (mchc MockCompletedHC) String() string {
	return mchc.Result
}
func (mchc MockCompletedHC) ExitCode() int {
	return mchc.Exit
}
func (mchc MockCompletedHC) Passed() bool {
	return mchc.Pass
}

type MockPlatform struct {
	Check    health.HealthCheck
	Canceler context.CancelFunc
}

func (mp *MockPlatform) Type() types.PlatformType {
	mock := "mock"
	return types.PlatformType(mock)
}

func (mp *MockPlatform) Send(m msg.Message) error {
	defer mp.Canceler()
	mp.Check = m.(msg.BackendMessage).CompletedCheck
	return nil
}

func TestBackendService_listen(t *testing.T) {
	testIO := []struct {
		name string
		ch   chan msg.Message
		hc   MockCompletedHC
		inp  backend.Platform
		test func(ch chan msg.Message, args backend.Platform)
	}{
		{
			name: "should have a default backend listening",
			hc: MockCompletedHC{
				Result: "foo",
				Exit:   127,
				Pass:   true,
			},
			ch:  make(chan msg.Message),
			inp: &MockPlatform{},
			test: func(ch chan msg.Message, args backend.Platform) {
				bes, _ := NewBackendService(nil, ch, nil)
				args.(*MockPlatform).Canceler = bes.ctx.Cancel
				pMap := map[types.PlatformType]backend.Platform{
					args.Type(): args,
				}
				bes.Platforms = pMap
				go bes.listen()
			},
		},
	}
	for _, io := range testIO {
		t.Run(io.name, func(t *testing.T) {
			io.test(io.ch, io.inp)
			m := msg.NewBackendMessage(io.hc, io.inp.Type())
			io.ch <- m
			completed := io.inp.(*MockPlatform).Check
			assert.Equal(t, completed.String(), io.hc.Result)
			assert.Equal(t, completed.ExitCode(), io.hc.Exit)
			assert.Equal(t, completed.Passed(), io.hc.Pass)
		})
	}
}
