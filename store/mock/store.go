package mock

import (
	"github.com/ahmadmuzakkir/redis-elasticsearch-go-example/store"
)

var _ store.Store = (*Store)(nil)

type Store struct {
	ViewTrackerStore   store.ViewTracker
	ViewRetrieverStore store.ViewRetriever
}

func (s *Store) ViewTracker() store.ViewTracker {
	return s.ViewTrackerStore
}

func (s *Store) ViewRetriever() store.ViewRetriever {
	return s.ViewRetrieverStore
}
