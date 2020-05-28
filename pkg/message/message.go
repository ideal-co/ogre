package msg

import "github.com/lowellmower/ogre/pkg/types"

// Message is the interface by which all processes communicate within the ogre
// application
type Message interface {
	Deserializer
	Serializer
	Type() types.MessageType
	Error() error
}

// Serializer is the interface which any Message must also implement
type Serializer interface {
	Serialize() ([]byte, error)
}

// Deserializer is the interface which any Message must also implement
type Deserializer interface {
	Deserialize([]byte) (Message, error)
}
