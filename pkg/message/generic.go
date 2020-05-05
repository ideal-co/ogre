package msg

import (
    "encoding/json"
    "fmt"
)

type GenericMessage struct {
    MType string `json:"service"`
    Err error `json:"err:omitempty"`
}

// GenericMessage's Type implementation of the Message interface will return
// the MessageType for the desired target service it has a message for and is
// provided in the MType field and will default to a MessageType of msg.Daemon.
func (gm GenericMessage) Type() MessageType {
    switch gm.MType {
    case "docker":
        return Docker
    default:
        return Daemon
    }
}

func (gm GenericMessage) Error() error {
    if gm.Err != nil {
        return gm.Err
    }
    return nil
}

func (gm GenericMessage) Serialize() ([]byte, error) {
    return json.Marshal(gm)
}

func (gm GenericMessage) Deserialize(data []byte) (Message, error) {
    err := json.Unmarshal(data, &gm)
    if err != nil {
        return nil, err
    }

    switch gm.Type() {
    case Docker:
        var dockerMsg DockerMessage
        return dockerMsg.Deserialize(data)
    case Daemon:
        var daemonMsg DaemonMessage
        return daemonMsg.Deserialize(data)
    default:
        return nil, fmt.Errorf("could not handle message %v", gm)
    }
}
