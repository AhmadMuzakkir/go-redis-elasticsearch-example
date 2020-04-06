// +build integration

package elastic

import (
	"context"
	"testing"
	"time"

	"github.com/ahmadmuzakkir/redis-elasticsearch-go-example/store"

	"github.com/stretchr/testify/assert"
)

// Test each range constant should have unit.
func TestRangeUnit(t *testing.T) {
	for i := 0; i < int(store.NumRange); i++ {
		if rangeUnit(store.Range(i)) == "" {
			t.Fatalf("unimplemented range unit for range %d", i)
		}
	}
}

func TestRetrieve(t *testing.T) {
	db, cleanup, err := connect(t)
	if !assert.NoError(t, err) {
		return
	}
	defer cleanup()

	err = db.viewTracker.Track(context.Background(), store.ViewTrack{
		ID:        "1",
		Timestamp: time.Now(),
	})
	if !assert.NoError(t, err) {
		return
	}

	// Wait for consistency, since elastic is eventual consistency
	time.Sleep(1 * time.Second)

	res, err := db.viewRetriever.Retrieve(context.Background(), "1", store.OneMinute)
	if !assert.NoError(t, err) {
		return
	}

	assert.Len(t, res, 1)

	assert.Equal(t, store.RangeDescription(store.OneMinute), res[0].Description)
	assert.Equal(t, int64(1), res[0].Count)

	res, err = db.viewRetriever.Retrieve(context.Background(), "2", store.OneMinute)
	if !assert.NoError(t, err) {
		return
	}

	assert.Len(t, res, 1)

	assert.Equal(t, store.RangeDescription(store.OneMinute), res[0].Description)
	assert.Equal(t, int64(0), res[0].Count)
}
