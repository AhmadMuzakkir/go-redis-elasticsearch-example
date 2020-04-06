package mock

import (
	"context"

	"github.com/ahmadmuzakkir/redis-elasticsearch-go-example/store"
)

var _ store.ViewRetriever = (*ViewRetriever)(nil)

type ViewRetriever struct {
	OnRetrieve func(ctx context.Context, id string, ranges ...store.Range) ([]store.ViewCount, error)
}

func (r *ViewRetriever) Retrieve(ctx context.Context, id string, ranges ...store.Range) ([]store.ViewCount, error) {
	return r.OnRetrieve(ctx, id, ranges...)
}
