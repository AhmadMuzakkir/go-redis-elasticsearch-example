// +build integration

package elastic

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	testElasticServerURL = "http://127.0.0.1:9200/"
)

func TestConnect(t *testing.T) {
	db, cleanup, err := connect(t)
	if !assert.NoError(t, err) {
		return
	}
	defer cleanup()

	exists, err := db.client.IndexExists(indexName).Do(context.Background())
	if !assert.NoError(t, err) {
		return
	}

	assert.True(t, exists, "index does not exist")
}

func connect(t *testing.T) (*Store, func(), error) {
	_, err := wait(testElasticServerURL)
	if err != nil {
		return nil, func() {}, err
	}

	db, err := Connect(testElasticServerURL)
	if err != nil {
		return nil, func() {}, err
	}

	cleanup := func() {
		_, err = db.client.DeleteIndex(indexName).Do(context.Background())
		if !assert.NoError(t, err) {
			return
		}
	}

	return db, cleanup, nil
}

// Keep trying to reach elastic server every one second, for 30 seconds.
func wait(elasticURL string) (bool, error) {
	timeout := 30 * time.Second

	start := time.Now()

	client := &http.Client{
		Timeout: 1 * time.Second,
	}

	req, err := http.NewRequest("GET", elasticURL, nil)
	if err != nil {
		return false, err
	}

	for {
		resp, _ := client.Do(req)
		if resp != nil && resp.StatusCode == http.StatusOK {
			return true, nil
		}

		if time.Since(start) > timeout {
			return false, nil
		}

		time.Sleep(1 * time.Second)
	}
}
