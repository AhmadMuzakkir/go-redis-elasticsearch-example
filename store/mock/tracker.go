package mock

import (
	"context"

	"github.com/ahmadmuzakkir/redis-elasticsearch-go-example/store"
)

var _ store.ViewTracker = (*ViewTracker)(nil)

type ViewTracker struct {
	OnTrack      func(ctx context.Context, v store.ViewTrack) error
	OnBatchTrack func(ctx context.Context, vs []store.ViewTrack) error
}

func (t *ViewTracker) Track(ctx context.Context, v store.ViewTrack) error {
	return t.OnTrack(ctx, v)
}

func (t *ViewTracker) BatchTrack(ctx context.Context, vs []store.ViewTrack) error {
	return t.OnBatchTrack(ctx, vs)
}
