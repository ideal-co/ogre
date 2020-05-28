package msg

import (
	"encoding/json"
	"github.com/lowellmower/ogre/pkg/types"
)

// DaemonMessage implements the msg.Message interface and is responsible
// for interfacing with the daemon process, indicating to it some specific
// actions should take place.
type DaemonMessage struct {
	// the string indicating some action needs to be taken by the daemon
	Action string `json:"action"`

	// retaining for future use of indicating process ending failures
	// or service level errors severe enough to indicate a lower level
	// action be taken.
	Err error `json:"err:omitempty"`
}

// Type is the DaemonMessage type implementation of the Message interface's
// Type method and will always return a types.DaemonMessage
func (dm DaemonMessage) Type() types.MessageType {
	return types.DaemonMessage
}

// Error is currently unused see comment on corresponding struct field
func (dm DaemonMessage) Error() error {
	if dm.Err != nil {
		return dm.Err
	}

	return nil
}

// Serialize is the DaemonMessage type implementation of the Message interface's
// Serialize method. It returns a slice of bytes and an error, the latter will
// be nil on successful serialization of the calling DaemonMessage struct.
func (dm DaemonMessage) Serialize() ([]byte, error) {
	return json.Marshal(dm)
}

// Deserialize is the DaemonMessage type implementation of the Message interface's
// Deserialize method. It returns a Message of type DaemonMessage and an error of
// which the latter will be nil upon successful unmarshalling of the byte slice
// passed into a DaemonMessage.
func (dm DaemonMessage) Deserialize(data []byte) (Message, error) {
	err := json.Unmarshal(data, &dm)
	if err != nil {
		return nil, err
	}
	return dm, nil
}
