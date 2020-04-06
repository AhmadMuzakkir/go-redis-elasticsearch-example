package elastic

import (
	"context"

	"github.com/ahmadmuzakkir/redis-elasticsearch-go-example/store"

	"github.com/olivere/elastic/v7"
)

var _ store.ViewTracker = (*viewTracker)(nil)

type viewTracker struct {
	client *elastic.Client
}

func (t *viewTracker) Track(ctx context.Context, v store.ViewTrack) error {
	_, err := t.client.Index().Index(indexName).BodyJson(v).Do(ctx)
	return err
}

func (t *viewTracker) BatchTrack(ctx context.Context, vs []store.ViewTrack) error {
	bulk := t.client.Bulk()

	for i := range vs {
		bulk.Add(elastic.NewBulkIndexRequest().Index(indexName).UseEasyJSON(true).Doc(vs[i]))
	}

	_, err := bulk.Do(ctx)

	return err
}
