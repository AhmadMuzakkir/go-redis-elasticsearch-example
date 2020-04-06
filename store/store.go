package store

import (
	"context"
	"time"
)

type ViewTrack struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
}

type ViewCount struct {
	Description string `json:"reference"`
	Count       int64  `json:"count"`
}

type ViewTracker interface {
	Track(ctx context.Context, v ViewTrack) error

	BatchTrack(ctx context.Context, vs []ViewTrack) error
}

type ViewRetriever interface {
	Retrieve(ctx context.Context, id string, ranges ...Range) ([]ViewCount, error)
}

type Store interface {
	ViewTracker() ViewTracker

	ViewRetriever() ViewRetriever
}
