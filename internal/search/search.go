package search

import (
	"context"
	"crypto/md5"
	"encoding/hex"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/elipatov/web-crawler/pkg/errs"
)

type Store struct {
	client *elasticsearch.TypedClient
}

func New(addresses []string) (*Store, error) {
	cfg := elasticsearch.Config{
		Addresses: addresses,
	}

	clientTyped, err := elasticsearch.NewTypedClient(cfg)
	if err != nil {
		return nil, errs.WrapError(err)
	}

	return &Store{
		client: clientTyped,
	}, nil
}

// Document represents the body of an indexed document in Elasticsearch.
type Document struct {
	URL  string `json:"url"`
	Text string `json:"text"`
}

// urlToKey hashes a URL to an Elasticsearch-safe document ID.
func urlToKey(url string) string {
	hash := md5.Sum([]byte(url))
	return hex.EncodeToString(hash[:])
}

func (s *Store) Set(ctx context.Context, url, text string) error {
	doc := Document{
		URL:  url,
		Text: text,
	}

	_, err := s.client.Index("documents").
		Id(urlToKey(url)).
		Document(doc).
		Do(ctx)
	if err != nil {
		return errs.WrapError(err)
	}

	return nil
}

func (s *Store) Close(ctx context.Context) error {
	err := s.client.Close(ctx)
	if err != nil {
		return errs.WrapError(err)
	}

	err = s.client.Close(ctx)
	if err != nil {
		return errs.WrapError(err)
	}

	return nil
}
