package eventstream

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{}
}

type Stream interface {
	Publish(tag string, data []byte)
}

type Broadcaster struct {
	Streams []Stream
}

func (broadcast *Broadcaster) Publish(tag string, data []byte) {
	for _, stream := range broadcast.Streams {
		go stream.Publish(tag, data)
	}
}

func (broadcaster *Broadcaster) AddStream(stream Stream) {
	broadcaster.Streams = append(broadcaster.Streams, stream)
}
