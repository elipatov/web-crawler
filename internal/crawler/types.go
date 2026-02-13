package crawler

import (
	"context"
	"time"

	"github.com/elipatov/web-crawler/pkg/contracts"
	"github.com/elipatov/web-crawler/pkg/queue"
)

type Config struct {
	ReprocessDelay time.Duration
	TTL            time.Duration
}

type ResourceInfo struct {
	Url       string
	Timestamp time.Time
}

type Queuer interface {
	Enqueue(context.Context, contracts.Resource) error
	Dequeue() queue.Item[contracts.Resource]
}

type ResourceStorer interface {
	Get(context.Context, string) (ResourceInfo, error)
	Set(context.Context, string, ResourceInfo) error
}

type TextStorer interface {
	Set(ctx context.Context, url, text string) error
}
