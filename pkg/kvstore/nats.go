package kvstore

import (
	"context"
	"encoding/json"

	"github.com/elipatov/web-crawler/pkg/errs"
	"github.com/elipatov/web-crawler/pkg/logger"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type Store[T any] struct {
	store jetstream.KeyValue
}

// New creates new queue instance.
func New[T any](ctx context.Context, logger *logger.Logger, bucket string, natsConn *nats.Conn) (*Store[T], error) {
	js, err := jetstream.New(natsConn)
	if err != nil {
		return nil, errs.WrapError(err)
	}

	store, err := js.CreateOrUpdateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket: bucket,
	})
	if err != nil {
		return nil, errs.WrapError(err, "failed to init Key-Value store")
	}

	return &Store[T]{
		store: store,
	}, nil
}

func (s *Store[T]) Get(ctx context.Context, key string) (T, error) {
	var res T

	kv, err := s.store.Get(ctx, key)
	if err != nil {
		return res, errs.WrapError(err)
	}

	err = json.Unmarshal(kv.Value(), &res)
	if err != nil {
		return res, errs.WrapError(err)
	}

	return res, nil
}

func (s *Store[T]) Set(ctx context.Context, key string, value T) error {
	buf, err := json.Marshal(value)
	if err != nil {
		return errs.WrapError(err)
	}

	_, err = s.store.Put(ctx, key, buf)
	if err != nil {
		return errs.WrapError(err)
	}

	return nil
}
