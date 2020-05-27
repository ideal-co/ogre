package msg

import (
	"encoding/json"
	"github.com/lowellmower/ogre/pkg/health"
	"github.com/lowellmower/ogre/pkg/types"
)

type BackendMessage struct {
	CompletedCheck health.HealthCheck
	Destination    types.PlatformType
	Data           *health.ExecResult
	Err            error
}

func (bm BackendMessage) Type() types.MessageType {
	return types.BackendMessage
}

func (bm BackendMessage) Error() error {
	if bm.Err != nil {
		return bm.Err
	}

	return nil
}

func (bm BackendMessage) Serialize() ([]byte, error) {
	m := BackendMessage{
		Data: bm.CompletedCheck.(*health.DockerHealthCheck).Result,
	}
	return json.Marshal(m)
}

func (bm BackendMessage) Deserialize(data []byte) (Message, error) {
	err := json.Unmarshal(data, &bm)
	if err != nil {
		return nil, err
	}
	return bm, nil
}

func NewBackendMessage(hc health.HealthCheck, dest types.PlatformType) Message {
	return BackendMessage{
		CompletedCheck: hc,
		Destination:    dest,
		Data:           hc.(*health.DockerHealthCheck).Result,
		Err:            nil,
	}
}
