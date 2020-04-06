package elastic

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/ahmadmuzakkir/redis-elasticsearch-go-example/store"

	"github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
)

var _ store.Store = (*Store)(nil)

const indexName = "views"

type Store struct {
	client *elastic.Client

	viewTracker   *viewTracker
	viewRetriever *viewRetriever
}

func Connect(serverUrl string) (*Store, error) {
	httpClient := &http.Client{
		Timeout: 3 * time.Second,
	}

	client, err := elastic.NewClient(elastic.SetURL(serverUrl), elastic.SetHttpClient(httpClient))
	if err != nil {
		return nil, errors.Wrap(err, "client")
	}

	_, _, err = client.Ping(serverUrl).Timeout("5").Do(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "ping")
	}

	if exists, err := client.IndexExists(indexName).Do(context.Background()); err != nil {
		return nil, errors.Wrap(err, "index exists")
	} else if !exists {
		res, err := client.CreateIndex(indexName).BodyString(mapping).Do(context.Background())

		// Ignore error index already exists.
		// For some reason, sometimes IndexExists() return false even if the Index already exists.
		if err != nil && !strings.Contains(err.Error(), "resource_already_exists_exception") {
			return nil, errors.Wrap(err, "create index")
		}

		if res != nil && !res.Acknowledged {
			return nil, errors.New("create index acknowledged is false")
		}
	}

	s := &Store{
		client:        client,
		viewTracker:   &viewTracker{client: client},
		viewRetriever: &viewRetriever{client: client},
	}

	return s, err
}

func (s *Store) ViewTracker() store.ViewTracker {
	return s.viewTracker
}

func (s *Store) ViewRetriever() store.ViewRetriever {
	return s.viewRetriever
}
