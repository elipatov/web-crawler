package main

type config struct {
	NATS natsConfig
}

type natsConfig struct {
	URL         string `env:"NATS_URL" envDefault:"nats://127.0.0.1:4222"`
	Stream      string `env:"NATS_URL" envDefault:"webcrawler"`
	Subject     string `env:"NATS_SUBJECT" envDefault:"resources"`
	ConsumerID  string `env:"NATS_CONSUMER_ID" envDefault:"webcrawler"`
	Concurrency int    `env:"NATS_CONCURRENCY" envDefault:"1"`
}
