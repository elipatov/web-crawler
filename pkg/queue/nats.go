package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type Queue[T any] struct {
	conn *nats.Conn
	js   jetstream.JetStream
	cfg  *Config
}

// New creates new queue instance.
func New[T any](ctx context.Context, cfg *Config, natsConn *nats.Conn) (*Queue[T], error) {
	js, err := jetstream.New(natsConn)
	if err != nil {
		return nil, err
	}

	res := &Queue[T]{
		conn: natsConn,
		js:   js,
		cfg:  cfg,
	}

	err = res.initStreams(ctx)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// Add publishes the provided payload to the configured NATS subject.
func (q *Queue[T]) Add(ctx context.Context, value T) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}

	_, err = q.js.Publish(ctx, q.cfg.Subject, payload)
	if err != nil {
		return err
	}

	return nil
}

func (q *Queue[T]) Consume(ctx context.Context) error {
	consConfig := jetstream.ConsumerConfig{
		Durable:       q.cfg.ConsumerID,
		AckPolicy:     jetstream.AckExplicitPolicy,
		FilterSubject: q.cfg.Subject,
	}

	cons, err := q.js.CreateOrUpdateConsumer(ctx, q.cfg.Stream.Name, consConfig)
	if err != nil {
		return err
	}

	for i := 0; i < q.cfg.Concurrency; i++ {
		consCtx, err := cons.Consume(q.handle)
		if err != nil {
			return err
		}

		defer consCtx.Stop()
	}

	<-ctx.Done()

	return nil
}

func (q *Queue[T]) handle(msg jetstream.Msg) {
	defer func() {
		err := msg.Ack()
		if err != nil {
			fmt.Println(err)
		}
	}()

	fmt.Printf("%s: %s\n", msg.Subject(), string(msg.Data()))
}

func (q *Queue[T]) initStreams(ctx context.Context) error {
	streamCfg := jetstream.StreamConfig{
		Name:      q.cfg.Stream.Name,
		Subjects:  []string{q.cfg.Subject},
		MaxAge:    q.cfg.Stream.RetentionTime,
		Retention: jetstream.LimitsPolicy,
		Discard:   jetstream.DiscardNew,
	}

	_, err := q.js.CreateOrUpdateStream(ctx, streamCfg)
	if err != nil {
		return err
	}

	return nil
}
