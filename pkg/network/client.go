package network

import msg "github.com/lowellmower/ogre/pkg/message"

type Client interface {
    msg.Serializer
    msg.Deserializer
}
