package queue

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/ahmadmuzakkir/redis-elasticsearch-go-example/store"

	"github.com/pkg/errors"
)

// ViewTrackerQueue wraps ViewTracker, to enable batching.
// The ViewTrack's will be sent using BatchTrack().
//
// The batch is flushed when one the below parameters is fulfilled:
// - batchSize: the maximum amount of documents in a batch.
// - batchInterval: the interval between indexes.
//
// The maximum delay before a ViewTrack is indexed is equal to batchInterval.
//
// BatchTrack will be called in a separate Goroutine, so the queue channel is never blocked.
type ViewTrackerQueue struct {
	batchSize     int
	batchInterval time.Duration

	logger *log.Logger

	queue       chan store.ViewTrack
	stopWg      sync.WaitGroup
	stopped     bool
	viewTracker store.ViewTracker
}

type Option func(*ViewTrackerQueue)

// Default batch size is 256
func WithBatchSize(batchSize int) func(*ViewTrackerQueue) {
	return func(queue *ViewTrackerQueue) {
		queue.batchSize = batchSize
	}
}

// Default batch interval is 3 seconds.
func WithBatchInterval(batchInterval time.Duration) func(*ViewTrackerQueue) {
	return func(queue *ViewTrackerQueue) {
		queue.batchInterval = batchInterval
	}
}

func NewViewTrackerQueue(viewTracker store.ViewTracker, logger *log.Logger, opts ...Option) *ViewTrackerQueue {
	q := &ViewTrackerQueue{
		batchSize:     256,
		batchInterval: 3 * time.Second,
		queue:         make(chan store.ViewTrack, 128),
		viewTracker:   viewTracker,
		logger:        logger,
	}

	for _, opt := range opts {
		opt(q)
	}

	q.stopWg.Add(1)
	go q.run()

	return q
}

func (q *ViewTrackerQueue) run() {
	defer q.stopWg.Done()

	buf := make([]store.ViewTrack, 0, q.batchSize)

Outer:
	for {
		select {
		case item, ok := <-q.queue:
			// The queue has been closed
			if !ok {
				break Outer
			}

			buf = append(buf, item)

			// If buf is full, send the buf.
			if len(buf) == cap(buf) {
				q.send(buf)

				buf = make([]store.ViewTrack, 0, q.batchSize)
			}

		case <-time.After(q.batchInterval):
			if len(buf) > 0 {
				q.send(buf)

				buf = make([]store.ViewTrack, 0, q.batchSize)
			}
		}
	}

	if len(buf) > 0 {
		q.send(buf)
	}
}

func (q *ViewTrackerQueue) send(buf []store.ViewTrack) {
	q.stopWg.Add(1)

	go func() {
		defer q.stopWg.Done()

		// Maximum retry 3 times
		for retry := 1; retry <= 4; retry++ {
			err := q.viewTracker.BatchTrack(context.Background(), buf)
			if err == nil {
				return
			}

			q.logger.Printf("ERROR queue batch track (retry: %d): %v", retry, err)

			time.Sleep(time.Duration(retry) * time.Second)
		}
	}()
}

// Track send the view into the queue.
func (q *ViewTrackerQueue) Track(ctx context.Context, view store.ViewTrack) error {
	if q.stopped {
		return errors.New("queue has stopped")
	}

	q.queue <- view

	return nil
}

// BatchTrack send the the batch of views directly to the storage, instead of using queue.
func (q *ViewTrackerQueue) BatchTrack(ctx context.Context, views []store.ViewTrack) error {
	if q.stopped {
		return errors.New("queue has stopped")
	}

	return q.viewTracker.BatchTrack(ctx, views)
}

func (q *ViewTrackerQueue) Stop(ctx context.Context) {
	if q.stopped {
		return
	}

	close(q.queue)

	done := make(chan struct{})

	go func() {
		q.stopWg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
	case <-done:
	}

	q.stopped = true
}
