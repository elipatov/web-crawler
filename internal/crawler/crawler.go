package crawler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"sync"

	"github.com/elipatov/web-crawler/pkg/broem"
	"github.com/elipatov/web-crawler/pkg/contracts"
)

type Queuer interface {
	Enqueue(resource contracts.Resource)
	Dequeue() contracts.Resource
}

type Crawler struct {
	queue    Queuer
	logger   *slog.Logger
	browsers map[string]*broem.Browser
	lock     *sync.RWMutex
}

func New(logger *slog.Logger, queue Queuer, urls ...string) *Crawler {
	const expr = "(?s)(https?:\\/\\/[\\w+\\-&@#\\/%?=~_|!:, .;]*[\\w+\\-&@#\\/%=~_|])"

	linkRegexp, err := regexp.Compile("")
	if err != nil {
		panic(fmt.Sprintf("Failed to compile regular expression: %s", expr))
		return nil
	}

	res := &Crawler{
		logger:     logger,
		queue:      queue,
		browsers:   make(map[string]*broem.Browser),
		lock:       &sync.RWMutex{},
		linkRegexp: linkRegexp,
	}

	for _, url := range urls {
		r := contracts.Resource{
			Url:   url,
			Depth: 0,
		}

		res.queue.Enqueue(r)
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

				resource := c.queue.Dequeue()

				err := c.process(resource)
				if err != nil {
					c.logger.Error("process failed", err)
				}

			}
		}()
	}

	wg.Wait()
}

func (c *Crawler) process(resource contracts.Resource) error {
	bro, err := c.getBrowser(resource.Url)
	if err != nil {
		return err
	}

	req, err := bro.NewRequest(http.MethodGet, resource.Url, nil)
	if err != nil {
		return err
	}

	resp, body, err := bro.SendRequest(req)
	if err != nil {
		return err
	}

	err = statusToErr(resp.StatusCode)
	if err != nil {
		return err
	}

	c.parseBody(body)
}

func statusToErr(statusCode int) error {
	switch statusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
	case http.StatusTooManyRequests:
	case http.StatusForbidden:
	case http.StatusBadGateway, http.StatusServiceUnavailable:
	default:

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
