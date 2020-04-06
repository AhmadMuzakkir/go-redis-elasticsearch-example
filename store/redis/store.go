package redis

import (
	"github.com/ahmadmuzakkir/redis-elasticsearch-go-example/store"

	"github.com/adjust/rmq"
)

var _ store.ViewTracker = (*ViewTracker)(nil)

// addr must include port. e.g. 127.0.0.1:6379
func Connect(addr string, db int) *ViewTracker {
	conn := rmq.OpenConnection("producer", "tcp", addr, db)

	viewTracker := &ViewTracker{
		conn:        conn,
		singleQueue: conn.OpenQueue("view_single"),
		batchQueue:  conn.OpenQueue("view_batch"),
	}

	return viewTracker
}
