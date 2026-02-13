package search

import (
	"context"
	"fmt"

	"github.com/elastic/go-elasticsearch/v9"
)

type Store struct {
	client *elasticsearch.Client
}

func New(addresses []string) (*Store, error) {
	cfg := elasticsearch.Config{
		Addresses: addresses,
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("elasticsearch client: %w", err)
	}

	return &Store{
		client: client,
	}, nil
}

func (s *Store) Set(ctx context.Context, url, text string) error {
	return nil
}
