package msg

import (
    "encoding/json"
    "github.com/docker/docker/api/types/events"
    "github.com/lowellmower/ogre/pkg/log"
)

type DockerMessage struct {
    Event events.Message `json:"event,omitempty"`
    Actor events.Actor `json:"actor,omitempty"`

    Action string `json:"action"`
    Err error `json:"err,omitempty"`
}

func (dm DockerMessage) Type() MessageType {
    return Docker
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
    log.Daemon.Tracef("DATA: %s", string(data))
    err := json.Unmarshal(data, &dm)
    if err != nil {
        return nil, err
    }
    log.Daemon.Tracef("DM: %+v", dm)
    return dm, nil
}

func NewDockerMessage(m events.Message, e error) Message {
    return DockerMessage{
        Event: m,
        Actor: m.Actor,
        Err: e,
    }
}