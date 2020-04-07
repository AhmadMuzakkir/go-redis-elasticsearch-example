package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ahmadmuzakkir/redis-elasticsearch-go-example/api"
	"github.com/ahmadmuzakkir/redis-elasticsearch-go-example/queue"
	"github.com/ahmadmuzakkir/redis-elasticsearch-go-example/store/elastic"
	"github.com/ahmadmuzakkir/redis-elasticsearch-go-example/store/redis"
	"github.com/ahmadmuzakkir/redis-elasticsearch-go-example/version"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/namsral/flag"
)

// Server provides HTTP APIs to track and retrieve views.
// It acts as the producer of the message queue, where it'll insert track event messages into the queue.

var (
	logger = log.New(os.Stdout, "server", log.LstdFlags|log.LUTC)
)

func main() {
	logger.Printf("version.BuildTime: %v, version.Commit: %v\n", version.BuildTime, version.Commit)

	portFlag := flag.Int("port", 8001, "API server port, default is 8001")
	elasticURLFlag := flag.String("elastic_url", "http://127.0.0.1:9200", "Elastic server URL, must include protocol, default is http://127.0.0.1:9200")
	redisAddrFlag := flag.String("redis_addr", "127.0.0.1:6379", "Redis host addr,  default is 127.0.0.1:6379")

	flag.Parse()

	port := *portFlag
	elasticURL := *elasticURLFlag
	redisAddr := *redisAddrFlag

	elasticDb, err := elastic.Connect(elasticURL)
	if err != nil {
		panic(err)
	}

	redisTrackerQueue := redis.Connect(redisAddr, 0)

	// If you don't want to redis, you can use the elastic store directly. Refer to below code
	//viewTrackerQueue := queue.NewViewTrackerQueue(redisTrackerQueue, logger, queue.WithBatchSize(256), queue.WithBatchInterval(3*time.Second))

	viewTrackerQueue := queue.NewViewTrackerQueue(redisTrackerQueue, logger, queue.WithBatchSize(256), queue.WithBatchInterval(3*time.Second))

	apiHandler := api.NewHandler(viewTrackerQueue, elasticDb.ViewRetriever(), logger)

	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	router.Mount("/", apiHandler)

	srv := &http.Server{
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		Addr:         ":" + strconv.Itoa(port),
		Handler:      router,
	}

	fmt.Printf("starting server on port %d\n", port)

	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			panic(fmt.Errorf("error starting server: %v", err))
		}
	}()

	// Wait for terminate signal
	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, os.Interrupt, syscall.SIGTERM)
	<-shutdownSignal

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Shut down gracefully, but wait no longer than 60 seconds before stopping.
	if err := srv.Shutdown(ctx); err != nil {
		fmt.Printf("error shutting down server: %v\n", err)
	}

	viewTrackerQueue.Stop(ctx)
}
