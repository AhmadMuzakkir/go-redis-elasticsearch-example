package queue

import (
	"context"
	"testing"
	"time"

	"github.com/ahmadmuzakkir/redis-elasticsearch-go-example/store"
	"github.com/ahmadmuzakkir/redis-elasticsearch-go-example/store/mock"

	"github.com/stretchr/testify/assert"
)

func TestQueue(t *testing.T) {
	tracksCh := make(chan []store.ViewTrack, 1)

	viewTracker := &mock.ViewTracker{
		OnBatchTrack: func(ctx context.Context, vs []store.ViewTrack) error {
			tracksCh <- vs
			return nil
		},
	}

	queue := NewViewTrackerQueue(viewTracker, nil, WithBatchInterval(500*time.Millisecond), WithBatchSize(3))

	// Test batch interval

	for i := 0; i < 2; i++ {
		_ = queue.Track(context.Background(), store.ViewTrack{
			ID:        "1",
			Timestamp: time.Now(),
		})
	}

	tracks := wait(t, tracksCh, 600*time.Millisecond)
	assert.Len(t, tracks, 2)

	// Test batch size

	for i := 0; i < 3; i++ {
		_ = queue.Track(context.Background(), store.ViewTrack{
			ID:        "1",
			Timestamp: time.Now(),
		})
	}

	tracks = wait(t, tracksCh, 100*time.Millisecond)
	assert.Len(t, tracks, 3)

	// Test BatchTrack

	_ = queue.BatchTrack(context.Background(), []store.ViewTrack{
		{
			ID:        "1",
			Timestamp: time.Now(),
		},
		{
			ID:        "1",
			Timestamp: time.Now(),
		},
		{
			ID:        "1",
			Timestamp: time.Now(),
		},
	})

	tracks = wait(t, tracksCh, 100*time.Millisecond)
	assert.Len(t, tracks, 3)

	// Test Stop while there's an item in the queue

	_ = queue.Track(context.Background(), store.ViewTrack{
		ID:        "1",
		Timestamp: time.Now(),
	})

	queue.Stop(context.Background())
	// Second call should have not effect
	queue.Stop(context.Background())

	tracks = wait(t, tracksCh, 100*time.Millisecond)
	assert.Len(t, tracks, 1)

	// Test after the queue has stopped
	assert.Error(t, queue.Track(context.Background(), store.ViewTrack{}))
	assert.Error(t, queue.BatchTrack(context.Background(), nil))
}

func wait(t *testing.T, ch <-chan []store.ViewTrack, timeout time.Duration) []store.ViewTrack {
	select {
	case tracks := <-ch:
		return tracks

	case <-time.After(timeout):
		t.Fatal("timeout reading from tracks")
		return nil
	}
}
