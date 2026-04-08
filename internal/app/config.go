package app

import "time"

type Config struct {
	NATS           natsConfig
	HTTPAddress    string `env:"HTTP_ADDRESS" envDefault:":8080"`
	Elasticsearch  elasticsearchConfig
	Concurrency    int           `env:"CONCURRENCY" envDefault:"1"`
	ReprocessDelay time.Duration `env:"REPROCESS_DELAY" envDefault:"60s"`
	TTL            time.Duration `env:"TTL" envDefault:"720h"`
}

type natsConfig struct {
	URL         string `env:"NATS_URL" envDefault:"nats://nats-1:4222"`
	Stream      string `env:"NATS_STREAM" envDefault:"webcrawler"`
	Subject     string `env:"NATS_SUBJECT" envDefault:"resources"`
	Bucket      string `env:"NATS_BUCKET" envDefault:"webcrawler"`
	ConsumerID  string `env:"NATS_CONSUMER_ID" envDefault:"webcrawler"`
	Concurrency int    `env:"NATS_CONCURRENCY" envDefault:"1"`
}

type elasticsearchConfig struct {
	Addresses []string `env:"ELASTICSEARCH_ADDRESSES" envDefault:"http://127.0.0.1:9200"`
}
