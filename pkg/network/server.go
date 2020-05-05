package network

import (
	msg "github.com/lowellmower/ogre/pkg/message"
)

type Server interface {
	msg.Serializer
	msg.Deserializer
}


