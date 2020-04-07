package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ahmadmuzakkir/redis-elasticsearch-go-example/internal/worker"
	"github.com/ahmadmuzakkir/redis-elasticsearch-go-example/proto"
	"github.com/ahmadmuzakkir/redis-elasticsearch-go-example/store"
	"github.com/ahmadmuzakkir/redis-elasticsearch-go-example/store/elastic"
	"github.com/ahmadmuzakkir/redis-elasticsearch-go-example/version"

	"github.com/adjust/rmq"
	"github.com/namsral/flag"
)

// Indexer is the consumer of the redis message queue. It takes messages from the queue and index them into ElasticSearch.
//
// The messages will be indexed into ElasticSearch in batch ,with batch size of 256, and poll duration of 1 seconds.
// ElasticSearch likes it when documents are indexed in batch (bulk).
// Refer to https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-bulk.html

const (
	prefetchLimit = 512
)

var (
	logger = log.New(os.Stdout, "indexer", log.LstdFlags|log.LUTC)
)

func main() {
	logger.Printf("version.BuildTime: %v, version.Commit: %v\n", version.BuildTime, version.Commit)

	elasticURLFlag := flag.String("elastic_url", "http://127.0.0.1:9200", "ElasticSearch server URL, must include protocol, default is http://127.0.0.1:9200")
	redisAddrFlag := flag.String("redis_addr", "127.0.0.1:6379", "Redis host addr,  default is 127.0.0.1:6379")
	numWorkersFlag := flag.Uint("workers", 10, "Number of workers used to index to ElasticSearch, default is 10")
	flag.Parse()

	elasticURL := *elasticURLFlag
	redisAddr := *redisAddrFlag
	numWorkers := *numWorkersFlag

	db, err := elastic.Connect(elasticURL)
	if err != nil {
		panic(err)
	}

	connection := rmq.OpenConnection("consumer", "tcp", redisAddr, 0)

	workers := worker.NewWorkerPool()
	workers.Start(int(numWorkers))

	singleQueue := connection.OpenQueue("view_single")
	singleQueue.StartConsuming(prefetchLimit, 400*time.Millisecond)
	singleQueue.AddConsumer("queue_1", singleConsumer(db.ViewTracker(), workers))

	batchQueue := connection.OpenQueue("view_batch")
	batchQueue.StartConsuming(prefetchLimit, 400*time.Millisecond)
	batchQueue.AddConsumer("queue_1", batchConsumer(db.ViewTracker(), workers))

	// Wait for terminate signal
	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, os.Interrupt, syscall.SIGTERM)
	<-shutdownSignal

	singleQueue.Close()

	batchQueue.Close()

	workers.Stop()
}

// A helper type to implement Consume interface
type ConsumerFunc func(delivery rmq.Delivery)

func (c ConsumerFunc) Consume(delivery rmq.Delivery) {
	c(delivery)
}

func batchConsumer(viewTracker store.ViewTracker, workers *worker.Pool) ConsumerFunc {
	return func(delivery rmq.Delivery) {
		data := delivery.Payload()
		batch := &proto.ViewTrackBatchRequest{}

		if err := batch.Unmarshal([]byte(data)); err != nil {
			// For now, we ignore messages with Unmarshal error.

			logger.Printf("ERROR unmarshal message: %v", data)

			delivery.Ack()

			return
		}

		tracks := make([]store.ViewTrack, 0, len(batch.Requests))

		for _, req := range batch.Requests {
			tracks = append(tracks, store.ViewTrack{
				ID:        string(req.Id),
				Timestamp: time.Unix(0, req.Timestamp),
			})
		}

		workers.Queue(func() {
			if err := viewTracker.BatchTrack(context.Background(), tracks); err != nil {
				logger.Printf("ERROR batch track: %v\n", err)
			}

			delivery.Ack()
		})
	}
}

func singleConsumer(viewTracker store.ViewTracker, workers *worker.Pool) ConsumerFunc {
	return func(delivery rmq.Delivery) {
		data := delivery.Payload()
		req := &proto.ViewTrackRequest{}

		if err := req.Unmarshal([]byte(data)); err != nil {
			// For now, we ignore messages with Unmarshal error.

			logger.Printf("ERROR unmarshal message: %v", data)

			delivery.Ack()

			return
		}

		track := store.ViewTrack{
			ID:        string(req.Id),
			Timestamp: time.Unix(0, req.Timestamp),
		}

		workers.Queue(func() {
			if err := viewTracker.Track(context.Background(), track); err != nil {
				logger.Printf("ERROR track: %v\n", err)

			}

			delivery.Ack()
		})
	}
}
