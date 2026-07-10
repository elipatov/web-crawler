package crawler

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/elipatov/web-crawler/internal/parser"
	"github.com/elipatov/web-crawler/pkg/broem"
	"github.com/elipatov/web-crawler/pkg/contracts"
	"github.com/elipatov/web-crawler/pkg/errs"
	"github.com/elipatov/web-crawler/pkg/logger"
)

const codeRetryable = errs.ErrorCode("RETRYABLE")

var errRetryable = errs.New(codeRetryable, "")

type Crawler struct {
	cfg           Config
	logger        *logger.Logger
	queue         Queuer
	resourceStore ResourceStorer
	textStore     TextStorer
	browsers      map[string]*broem.Browser
	parser        *parser.Parser
	lock          *sync.RWMutex
}

func New(
	ctx context.Context,
	cfg Config,
	logger *logger.Logger,
	queue Queuer,
	resourceStore ResourceStorer,
	textStore TextStorer,
	urls ...string,
) *Crawler {
	res := &Crawler{
		cfg:           cfg,
		logger:        logger,
		queue:         queue,
		resourceStore: resourceStore,
		textStore:     textStore,
		browsers:      make(map[string]*broem.Browser),
		lock:          &sync.RWMutex{},
		parser:        parser.New(logger),
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
					c.logger.WithError(err).With("url", msg.Item.Url).Error("process failed")

					if errors.Is(err, errRetryable) {
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

	contentType := resp.Header.Get("Content-Type")
	if !isHTMLContent(contentType) {
		return nil
	}

	parseRes := c.parser.ParseBody(body)

	rInfo := ResourceInfo{
		Url:       resource.Url,
		Timestamp: time.Now().UTC(),
	}

	err = c.textStore.Set(ctx, resource.Url, parseRes.Text)
	if err != nil {
		return err
	}

	err = c.resourceStore.Set(ctx, urlToKey(resource.Url), rInfo)
	if err != nil {
		return err
	}

	for _, link := range parseRes.Links {
		key := urlToKey(link)
		r := contracts.Resource{
			Url:   link,
			Depth: resource.Depth + 1,
		}

		existing, err := c.resourceStore.Get(ctx, key)
		if err != nil && !errors.Is(err, errs.ErrNotFound) {
			return err
		}

		if errors.Is(err, errs.ErrNotFound) || time.Since(existing.Timestamp) > c.cfg.TTL {
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
		return errRetryable.WithMessage("Too Many Requests")
	case http.StatusBadGateway, http.StatusServiceUnavailable:
		return errRetryable.WithMessagef("Unavailable (status %d)", statusCode)
	default:
		return errs.ErrUnexpected.WithMessagef("Unexpected status %d", statusCode)
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

var htmlTypes = []string{
	"text/html",
	"application/xhtml+xml",
	"text/xml",
	"application/xml",
}

func isHTMLContent(contentType string) bool {
	if contentType == "" {
		return false
	}

	ct := strings.ToLower(strings.Split(contentType, ";")[0])
	for _, htmlType := range htmlTypes {
		if ct == htmlType {
			return true
		}
	}

	return false
}

func urlToKey(url string) string {
	hash := md5.Sum([]byte(url))
	return hex.EncodeToString(hash[:])
}
