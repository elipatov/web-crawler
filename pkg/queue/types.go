package queue

import (
	"github.com/nats-io/nats.go/jetstream"
)

type Config struct {
	NatsURL    string
	ConsumerID string
	Stream     StreamConfig
	Subject    string
}

// StreamConfig represents the configuration for the NATS stream.
type StreamConfig struct {
	Name string
}

type Item[T any] struct {
	Item T
	msg  jetstream.Msg
}
