package crawler

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/elipatov/web-crawler/internal/parser"
	"github.com/elipatov/web-crawler/pkg/broem"
	"github.com/elipatov/web-crawler/pkg/contracts"
	"github.com/elipatov/web-crawler/pkg/errs"
	"github.com/elipatov/web-crawler/pkg/logger"
)

const (
	codeRetryable = errs.ErrorCode("RETRYABLE")
)

type Crawler struct {
	cfg          Config
	logger       *logger.Logger
	queue        Queuer
	store        Storer
	browsers     map[string]*broem.Browser
	parser       *parser.Parser
	lock         *sync.RWMutex
	errRetryable *errs.Error
}

func New(
	ctx context.Context,
	cfg Config,
	logger *logger.Logger,
	queue Queuer,
	store Storer,
	urls ...string,
) *Crawler {
	res := &Crawler{
		cfg:          cfg,
		logger:       logger,
		queue:        queue,
		store:        store,
		browsers:     make(map[string]*broem.Browser),
		lock:         &sync.RWMutex{},
		parser:       parser.New(logger),
		errRetryable: errs.New(codeRetryable, ""),
	}

	for _, url := range urls {
		r := contracts.Resource{
			Url:   url,
			Depth: 0,
		}

		res.queue.Enqueue(ctx, r)
	}

	return res
}

func (c *Crawler) Run(ctx context.Context, concurrency int) {
	wg := new(sync.WaitGroup)

	wg.Add(concurrency)

	for range concurrency {
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				default:
				}

				msg := c.queue.Dequeue()

				err := c.process(ctx, msg.Item)
				if err != nil {
					c.logger.WithError(err).Error("process failed")

					tErr, ok := err.(*errs.Error)
					if ok && tErr.ErrorCode() == codeRetryable {
						err = msg.NakWithDelay(c.cfg.ReprocessDelay)
						if err != nil {
							c.logger.WithError(err).Error("nak failed")
						}

						continue
					}
				}

				err = msg.Ack()
				if err != nil {
					c.logger.WithError(err).Error("ack failed")
				}
			}
		}()
	}

	wg.Wait()
}

func (c *Crawler) process(ctx context.Context, resource contracts.Resource) error {
	bro, err := c.getBrowser(resource.Url)
	if err != nil {
		return errs.WrapError(err)
	}

	req, err := bro.NewRequest(http.MethodGet, resource.Url, nil)
	if err != nil {
		return errs.WrapError(err)
	}

	resp, body, err := bro.SendRequest(req)
	if err != nil {
		return errs.WrapError(err)
	}

	err = c.statusToErr(resp.StatusCode)
	if err != nil {
		return err
	}

	parseRes := c.parser.ParseBody(body)

	for _, link := range parseRes.Links {
		r := contracts.Resource{
			Url:   link,
			Depth: resource.Depth + 1,
		}

		key := urlToKey(link)

		existing, err := c.store.Get(ctx, key)
		if err != nil && !errors.Is(err, errs.ErrNotFound) {
			return err
		}

		if errors.Is(err, errs.ErrNotFound) || time.Since(existing.timestamp) > c.cfg.TTL {
			err = c.queue.Enqueue(ctx, r)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Crawler) statusToErr(statusCode int) error {
	switch statusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return errs.ErrNotFound
	case http.StatusUnauthorized:
		return errs.ErrUnauthorized
	case http.StatusForbidden:
		return errs.ErrForbidden
	case http.StatusUnavailableForLegalReasons:
		return errs.ErrUnexpected
	case http.StatusTooManyRequests:
		return c.errRetryable.WithMessage("Too Many Requests")
	case http.StatusBadGateway, http.StatusServiceUnavailable:
		return c.errRetryable.WithMessagef("Unavailable (status %s)", statusCode)
	default:
		return errs.ErrUnexpected.WithMessagef("Unexpected status %s", statusCode)
	}
}

func (c *Crawler) getBrowser(address string) (*broem.Browser, error) {
	uri, err := url.Parse(address)
	if err != nil {
		return nil, err
	}

	c.lock.RLock()
	br, ok := c.browsers[uri.Host]
	c.lock.RUnlock()

	if !ok {
		origin := url.URL{
			Scheme: uri.Scheme,
			Host:   uri.Host,
		}

		br = broem.New(origin.String(), "", nil)

		c.lock.Lock()
		c.browsers[uri.Host] = br
		c.lock.Unlock()
	}

	return br, nil
}

func urlToKey(url string) string {
	return url
}
