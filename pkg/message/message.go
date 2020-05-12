package msg

import "github.com/lowellmower/ogre/pkg/types"

type Message interface {
	Deserializer
	Serializer
	Type() types.MessageType
	Error() error
}

type Serializer interface {
	Serialize() ([]byte, error)
}

type Deserializer interface {
	Deserialize([]byte) (Message, error)
}
