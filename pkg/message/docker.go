package msg

import (
	"encoding/json"
	"github.com/docker/docker/api/types/events"
	"github.com/lowellmower/ogre/pkg/health"
	"github.com/lowellmower/ogre/pkg/types"
)

type DockerMessage struct {
	Event  events.Message     `json:"event,omitempty"`
	Actor  events.Actor       `json:"actor,omitempty"`
	Health health.HealthCheck `json:"health,omitempty"`

	Action string `json:"action"`
	Err    error  `json:"err,omitempty"`
}

func (dm DockerMessage) Type() types.MessageType {
	return types.DockerMessage
}

func (dm DockerMessage) Error() error {
	if dm.Err != nil {
		return dm.Err
	}

	return nil
}

func (dm DockerMessage) Serialize() ([]byte, error) {
	return json.Marshal(dm)
}

func (dm DockerMessage) Deserialize(data []byte) (Message, error) {
	err := json.Unmarshal(data, &dm)
	if err != nil {
		return nil, err
	}
	return dm, nil
}

func NewDockerMessage(m events.Message, e error) Message {
	return DockerMessage{
		Event: m,
		Actor: m.Actor,
		Err:   e,
	}
}
