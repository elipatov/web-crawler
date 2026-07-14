package app

import (
	"context"
	"errors"

	"github.com/elipatov/web-crawler/internal/crawler"
	"github.com/elipatov/web-crawler/internal/search"
	"github.com/elipatov/web-crawler/pkg/contracts"
	"github.com/elipatov/web-crawler/pkg/errs"
	"github.com/elipatov/web-crawler/pkg/kvstore"
	"github.com/elipatov/web-crawler/pkg/logger"
	"github.com/elipatov/web-crawler/pkg/queue"
	"github.com/nats-io/nats.go"
	"golang.org/x/sync/errgroup"
)

type App struct {
	queue       *queue.Queue[contracts.Resource]
	crawler     *crawler.Crawler
	cfg         Config
	logger      *logger.Logger
	conn        *nats.Conn
	searchStore *search.Store
}

func New(ctx context.Context, logger *logger.Logger, cfg Config) (*App, error) {
	conn, err := nats.Connect(cfg.NATS.URL)
	if err != nil {
		return nil, err
	}

	qCfg := queue.Config{
		ConsumerID: cfg.NATS.ConsumerID,
		Stream: queue.StreamConfig{
			Name: cfg.NATS.Stream,
		},
		Subject: cfg.NATS.Subject,
	}

	q, err := queue.New[contracts.Resource](ctx, logger, qCfg, conn)
	if err != nil {
		return nil, err
	}

	cCfg := crawler.Config{
		ReprocessDelay: cfg.ReprocessDelay,
		TTL:            cfg.TTL,
	}

	resourceStore, err := kvstore.New[crawler.ResourceInfo](ctx, cfg.NATS.Bucket, conn)
	if err != nil {
		return nil, err
	}

	searchStore, err := search.New(cfg.Elasticsearch.Addresses)
	if err != nil {
		return nil, err
	}

	app := &App{
		cfg:         cfg,
		logger:      logger,
		queue:       q,
		crawler:     crawler.New(cCfg, logger, q, resourceStore, searchStore),
		conn:        conn,
		searchStore: searchStore,
	}

	return app, nil

}

func (a *App) Run(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		return a.runHTTPServer(ctx)
	})

	group.Go(func() error {
		return a.process(ctx)
	})

	err := group.Wait()
	if err != nil && !errors.Is(err, context.Canceled) {
		return errs.WrapError(err)
	}

	a.logger.Info("application stopped")

	return nil
}

func (a *App) Close() {
	a.conn.Close()

	err := a.searchStore.Close(context.Background())
	if err != nil {
		a.logger.WithError(err).Error("failed to close search store")
	}
}

func (a *App) seed(ctx context.Context, urls ...string) error {
	a.logger.Info("seed")

	for _, url := range urls {
		select {
		case <-ctx.Done():
			return errs.WrapError(ctx.Err())
		default:
		}

		err := a.queue.Enqueue(ctx, contracts.Resource{Url: url})
		if err != nil {
			return errs.WrapError(err)
		}
	}

	return nil
}

func (a *App) process(ctx context.Context) error {
	err := a.queue.Run(ctx)
	if err != nil {
		return err
	}

	a.crawler.Run(ctx, a.cfg.Concurrency)

	<-a.queue.Done()

	return nil
}
