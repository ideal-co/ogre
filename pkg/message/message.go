package msg

type MessageType string

const (
	Daemon MessageType = "daemon"
	Docker MessageType = "docker"
	Host MessageType = "host"
)

type Message interface {
	Deserializer
	Serializer
	Type() MessageType
	Error() error
}

type Serializer interface {
	Serialize() ([]byte, error)
}

type Deserializer interface {
	Deserialize([]byte) (Message, error)
}
