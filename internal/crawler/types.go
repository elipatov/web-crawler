package crawler

import (
	"context"
	"time"

	"github.com/elipatov/web-crawler/pkg/contracts"
	"github.com/elipatov/web-crawler/pkg/queue"
)

type Config struct {
	ReprocessDelay time.Duration
}

type ResourceInfo struct {
	url       string
	timestamp time.Time
}

type Queuer interface {
	Enqueue(context.Context, contracts.Resource) error
	Dequeue() queue.Item[contracts.Resource]
}

type Storer interface {
	Get(context.Context, string) (ResourceInfo, error)
	Set(context.Context, string, ResourceInfo) error
}
