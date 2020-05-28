package msg

import (
	"encoding/json"
	"github.com/lowellmower/ogre/pkg/health"
	"github.com/lowellmower/ogre/pkg/types"
)

// BackendMessage implements the Message interface and is the type responsible
// for interfacing between the services (i.e. Docker) and the backend.Platform
// interface. The backend is itself, a Service and can route the BackendMessage
// type to the appropriate Platform by way of the Destination field
type BackendMessage struct {
	// the executed health check (currently only DockerHealthCheck type)
	CompletedCheck health.HealthCheck

	// the backend platform the result is destined for
	Destination types.PlatformType

	// the exit code and result of stdout/stderr from running the health
	// check command, used to (De)Serialize JSON
	Data *health.ExecResult

	// retaining for future use of indicating process ending failures
	// or service level errors severe enough to indicate a lower level
	// action be taken.
	Err error
}

// NewBackendMessage takes a health.HealthCheck, a types.PlatformType, and a
// pointer to a health.ExecResult and returns a Message of type BackendMessage.
func NewBackendMessage(hc health.HealthCheck, dest types.PlatformType, er *health.ExecResult) Message {
	return BackendMessage{
		CompletedCheck: hc,
		Destination:    dest,
		Data:           er,
		Err:            nil,
	}
}

// Type is the BackendMessage type implementation of the Message interface's
// Type method and will always return a types.BackendMessage
func (bm BackendMessage) Type() types.MessageType {
	return types.BackendMessage
}

// Error is currently unused see comment on corresponding struct field
func (bm BackendMessage) Error() error {
	if bm.Err != nil {
		return bm.Err
	}
	return nil
}

// Serialize is the BackendMessage type implementation of the Message interface's
// Serialize method. Serialize will take the results of a DockerHealthCheck and
// return a slice of bytes and an error which will be nil upon success.
func (bm BackendMessage) Serialize() ([]byte, error) {
	m := BackendMessage{
		Data: bm.CompletedCheck.(*health.DockerHealthCheck).Result,
	}
	return json.Marshal(m)
}

// Deserialize is the BackendMessage type implementation of the Message interface's
// Deserialize method. Deserialize will take a slice of bytes and unmarshal that
// data into a Message of type BackendMessage, returning that and an error, the
// latter of which will be nil upon success.
func (bm BackendMessage) Deserialize(data []byte) (Message, error) {
	err := json.Unmarshal(data, &bm)
	if err != nil {
		return nil, err
	}
	return bm, nil
}
