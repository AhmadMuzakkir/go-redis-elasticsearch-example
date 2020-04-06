package redis

import (
	"context"
	"time"

	"github.com/ahmadmuzakkir/redis-elasticsearch-go-example/proto"
	"github.com/ahmadmuzakkir/redis-elasticsearch-go-example/store"

	"github.com/adjust/rmq"
)

var _ store.ViewTracker = (*ViewTracker)(nil)

type ViewTracker struct {
	conn rmq.Connection

	singleQueue rmq.Queue
	batchQueue  rmq.Queue
}

func (t *ViewTracker) Track(ctx context.Context, v store.ViewTrack) error {
	req := &proto.ViewTrackRequest{
		Id: []byte(v.ID),
		Timestamp: v.Timestamp.UnixNano(),
	}

	msg, err := req.Marshal()
	if err != nil {
		return err
	}

	t.singleQueue.PublishBytes(msg)

	return nil
}

// BatchTrack sends all the tracks in a single request.
func (t *ViewTracker) BatchTrack(ctx context.Context, vs []store.ViewTrack) error {
	batch := &proto.ViewTrackBatchRequest{
		Requests:      make([]*proto.ViewTrackRequest, 0, len(vs)),
		SentTimestamp: time.Now().UnixNano(),
	}

	for _, v := range vs {
		req := &proto.ViewTrackRequest{
			Id: []byte(v.ID),
			Timestamp: v.Timestamp.UnixNano(),
		}

		batch.Requests = append(batch.Requests, req)
	}

	msg, err := batch.Marshal()
	if err != nil {
		return err
	}

	t.batchQueue.PublishBytes(msg)

	return nil
}
