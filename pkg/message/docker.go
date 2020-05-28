package msg

import (
	"encoding/json"
	"github.com/docker/docker/api/types/events"
	"github.com/lowellmower/ogre/pkg/health"
	"github.com/lowellmower/ogre/pkg/types"
)

// DockerMessage implements the Message interface and is used to negotiate
// docker daemon and API issues up to the ogre daemon listener.
type DockerMessage struct {
	Event  events.Message     `json:"event,omitempty"`
	Actor  events.Actor       `json:"actor,omitempty"`
	Health health.HealthCheck `json:"health,omitempty"`

	Action string `json:"action"`
	Err    error  `json:"err,omitempty"`
}

// NewDockerMessage takes an events.Message (Docker API) and a string and returns
// a Message of type DockerMessage.
func NewDockerMessage(m events.Message, msg string) Message {
	return DockerMessage{
		Event:  m,
		Actor:  m.Actor,
		Action: msg,
	}
}

// Type is the DockerMessage type implementation of the Message interface's
// Type method and will always return a types.DockerMessage
func (dm DockerMessage) Type() types.MessageType {
	return types.DockerMessage
}

// Error is not currently used, reserved for future implementation. See other
// message types for detail.
func (dm DockerMessage) Error() error {
	if dm.Err != nil {
		return dm.Err
	}

	return nil
}

// Serialize is the DockerMessage type implementation of the Message interface's
// Serialize method. It returns a slice of bytes and an error, the latter will
// be nil on successful serialization of the calling DockerMessage struct.
func (dm DockerMessage) Serialize() ([]byte, error) {
	return json.Marshal(dm)
}

// Deserialize is the DockerMessage type implementation of the Message interface's
// Deserialize method. It returns a Message of type DockerMessage and an error of
// which the latter will be nil upon successful unmarshalling of the byte slice
// passed into a DockerMessage.
func (dm DockerMessage) Deserialize(data []byte) (Message, error) {
	err := json.Unmarshal(data, &dm)
	if err != nil {
		return nil, err
	}
	return dm, nil
}
