package queue

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"

	"github.com/elipatov/web-crawler/pkg/errs"
	"github.com/elipatov/web-crawler/pkg/logger"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type Queue[T any] struct {
	conn   *nats.Conn
	js     jetstream.JetStream
	cfg    Config
	logger *logger.Logger
	wg     *sync.WaitGroup
	out    <-chan Item[T]
}

// New creates new queue instance.
func New[T any](ctx context.Context, logger *logger.Logger, cfg Config, natsConn *nats.Conn) (*Queue[T], error) {
	js, err := jetstream.New(natsConn)
	if err != nil {
		return nil, err
	}

	res := &Queue[T]{
		conn:   natsConn,
		js:     js,
		logger: logger,
		cfg:    cfg,

		wg: new(sync.WaitGroup),
	}

	err = res.initStreams(ctx)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (q *Queue[T]) Run(ctx context.Context) error {
	consConfig := jetstream.ConsumerConfig{
		Durable:       q.cfg.ConsumerID,
		AckPolicy:     jetstream.AckExplicitPolicy,
		FilterSubject: q.cfg.Subject,
	}

	cons, err := q.js.CreateOrUpdateConsumer(ctx, q.cfg.Stream.Name, consConfig)
	if err != nil {
		return err
	}

	out := make(chan Item[T])

	q.out = out

	handler := func(msg jetstream.Msg) {
		var tItem T

		q.wg.Add(1)
		defer q.wg.Done()

		hErr := json.Unmarshal(msg.Data(), &tItem)
		if hErr != nil {
			q.logger.WithError(hErr).Error("unmarshal failed")

			hErr = msg.Ack()
			if hErr != nil {
				q.logger.WithError(hErr).Error("ack failed")
			}

			return
		}

		item := Item[T]{
			Item: tItem,
			msg:  msg,
		}

		select {
		case <-ctx.Done():
		case out <- item:
		}
	}

	q.wg.Add(1)
	defer q.wg.Done()

	consCtx, err := cons.Consume(handler)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		consCtx.Stop()
		close(out)
		q.wg.Done()
	}()

	return nil
}

// Enqueue publishes the provided payload to the configured NATS subject.
// If T implements Key(), it is used as the NATS deduplication ID.
// Otherwise the ID is derived from entite T.
func (q *Queue[T]) Enqueue(ctx context.Context, value T) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}

	var msgID string
	if k, ok := any(value).(interface{ Key() string }); ok {
		msgID = k.Key()
	} else {
		hash := md5.Sum(payload)
		msgID = hex.EncodeToString(hash[:])
	}

	jsMsg := nats.NewMsg(q.cfg.Subject)
	jsMsg.Data = payload
	jsMsg.Header.Set(jetstream.MsgIDHeader, msgID)

	_, err = q.js.PublishMsg(ctx, jsMsg)
	if err != nil {
		return err
	}

	return nil
}

func (q *Queue[T]) Dequeue() Item[T] {
	return <-q.out
}

func (q *Queue[T]) Chan() <-chan Item[T] {
	return q.out
}

// Done returns channel that got closed when consumer complete.
func (q *Queue[T]) Done() <-chan struct{} {
	done := make(chan struct{})

	go func() {
		defer close(done)
		q.wg.Wait()
	}()

	return done
}

func (q *Queue[T]) initStreams(ctx context.Context) error {
	streamCfg := jetstream.StreamConfig{
		Name:      q.cfg.Stream.Name,
		Subjects:  []string{q.cfg.Subject},
		Retention: jetstream.LimitsPolicy,
		Discard:   jetstream.DiscardNew,
	}

	_, err := q.js.CreateOrUpdateStream(ctx, streamCfg)
	if err != nil {
		return err
	}

	return nil
}

func (m Item[T]) Ack() error {
	return errs.WrapError(m.msg.Ack())
}

func (m Item[T]) Nack() error {
	return errs.WrapError(m.msg.Nak())
}

func (m Item[T]) NakWithDelay(delay time.Duration) error {
	return errs.WrapError(m.msg.NakWithDelay(delay))
}
