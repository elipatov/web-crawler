package queue

import "time"

type Config struct {
	NatsURL     string
	ConsumerID  string
	Concurrency int
	Stream      StreamConfig
	Subject     string
}

// StreamConfig represents the configuration for the NATS stream.
type StreamConfig struct {
	Name          string
	RetentionTime time.Duration
}
