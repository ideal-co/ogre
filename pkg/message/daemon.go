package msg

import (
    "encoding/json"
    "github.com/lowellmower/ogre/pkg/types"
)

type DaemonMessage struct {
    Action string `json:"action"`
    Err error `json:"err:omitempty"`
}

func (dm DaemonMessage) Type() types.MessageType {
    return types.DaemonMessage
}

func (dm DaemonMessage) Error() error {
    if dm.Err != nil {
        return dm.Err
    }

    return nil
}

func (dm DaemonMessage) Serialize() ([]byte, error) {
    return json.Marshal(dm)
}

func (dm DaemonMessage) Deserialize(data []byte) (Message, error) {
    err := json.Unmarshal(data, &dm)
    if err != nil {
        return nil, err
    }
    return dm, nil
}
