package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/elipatov/web-crawler/internal/crawler"
	"github.com/elipatov/web-crawler/pkg/queue"
	"github.com/nats-io/nats.go"
)

func main() {
	var mode string

	if len(os.Args) > 1 {
		mode = os.Args[1]
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	conf, err := newConfig()
	if err != nil {
		log.Fatalf("failed to apply configuration: %v", err)
	}

	fmt.Sprintf("Mode: %s\n", mode)

	cfg := &queue.Config{
		NatsURL:     conf.NATS.URL,
		Subject:     conf.NATS.Subject,
		ConsumerID:  conf.NATS.ConsumerID,
		Concurrency: conf.NATS.Concurrency,
		Stream: queue.StreamConfig{
			Name:          "webcrawler",
			RetentionTime: 15 * time.Second,
		},
	}

	ctx := appContext()

	conn, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		panic(err)
	}

	q, err := queue.New[string](ctx, cfg, conn)
	if err != nil {
		panic(err)
	}

	crawler.New(logger, q)

	switch mode {
	case "pub":
		fmt.Println("generate")
		generate(ctx, q)
	case "sub":
		fmt.Println("consume")
		q.Consume(ctx)
	default:
		fmt.Println("default")
		q.Consume(ctx)
		//generate(ctx, q)
	}

	//NewStream("Below", "are", "doc", "comments").Map(Enqueue(" ")).Apply(Drain())
}

func newConfig() (*config, error) {
	conf := new(config)

	err := env.Parse(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration: %w", err)
	}

	return conf, nil
}

func generate(ctx context.Context, queue *queue.Queue[string]) {
	for i := 0; ; i++ {
		select {
		case <-ctx.Done():
			return
		default:
		}

		err := queue.Add(ctx, fmt.Sprintf("val%d", i))
		if err != nil {
			fmt.Println(err)
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func appContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		c := make(chan os.Signal, 1)

		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		cancel()
	}()

	return ctx
}

//type Stream[T any] chan T
//
//func NewStream[T any](values ...T) Stream[T] {
//	res := make(Stream[T])
//
//	go func() {
//		defer close(res)
//
//		for _, val := range values {
//			res <- val
//		}
//	}()
//
//	return res
//}
//
//func (s Stream[T]) Map(fn func(T) T) Stream[T] {
//	res := make(Stream[T])
//
//	go func() {
//		defer close(res)
//
//		for val := range s {
//			res <- fn(val)
//		}
//	}()
//
//	return res
//}
//
//func (s Stream[T]) Apply(fn func(T)) {
//	for val := range s {
//		fn(val)
//	}
//}
//
//func Enqueue(additive string) func(string) string {
//	return func(val string) string {
//		return val + additive
//	}
//}
//
//func Drain() func(string) {
//	return func(val string) {
//		fmt.Print(val)
//	}
//}
//
//func Iterate[T any](done chan struct{}, values ...int) {
//	res := make(Stream[T])
//
//	go func() {
//		defer close(res)
//
//		for _, val := range values {
//			res <- val
//		}
//	}()
//
//	return res
//}
