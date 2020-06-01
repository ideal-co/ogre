package msg

import (
	"encoding/json"
	"fmt"
	"github.com/ideal-co/ogre/pkg/types"
)

type GenericMessage struct {
	MType string `json:"service"`
	Err   error  `json:"err:omitempty"`
}

// GenericMessage's Type implementation of the Message interface will return
// the MessageType for the desired target service it has a message for and is
// provided in the MType field and will default to a MessageType of msg.Daemon.
func (gm GenericMessage) Type() types.MessageType {
	switch gm.MType {
	case "docker":
		return types.DockerMessage
	case "backend":
		return types.BackendMessage
	default:
		return types.DaemonMessage
	}
}

// Error is not currently used, reserved for future implementation. See other
// message types for detail.
func (gm GenericMessage) Error() error {
	if gm.Err != nil {
		return gm.Err
	}
	return nil
}

// Serialize is the GenericMessage type implementation of the Message interface's
// Serialize method. It returns a slice of bytes and an error, the latter will
// be nil on successful serialization of the calling GenericMessage struct.
func (gm GenericMessage) Serialize() ([]byte, error) {
	return json.Marshal(gm)
}

// Deserialize is the GenericMessage type implementation of the Message interface's
// Deserialize method. It returns a Message of the type which is indicated by the
// MType field and can return the types found in the switch statement in the Type
// method. An error is non-nil if the underlying message type's Deserialize method
// results in an error, the data could not be unmarshaled into a GenericMessage, or
// the MType passed was not supported/recognized.
func (gm GenericMessage) Deserialize(data []byte) (Message, error) {
	err := json.Unmarshal(data, &gm)
	if err != nil {
		return nil, err
	}

	switch gm.Type() {
	case types.DockerMessage:
		var dockerMsg DockerMessage
		return dockerMsg.Deserialize(data)
	case types.BackendMessage:
		var backendMsg BackendMessage
		return backendMsg.Deserialize(data)
	case types.DaemonMessage:
		var daemonMsg DaemonMessage
		return daemonMsg.Deserialize(data)
	default:
		return nil, fmt.Errorf("could not handle message %v", gm)
	}
}
