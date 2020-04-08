package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/namsral/flag"
)

var (
	logger = log.New(os.Stdout, "indexer", log.LstdFlags|log.LUTC)
)

// Benchmark will repeatedly call /counter/v1/statistics endpoint as fast as possible using a number of Goroutine, and
// output the number of request per second.
func main() {
	numOfWorkersFlag := flag.Uint("w", 1, "Number of workers, default is 1")
	hitIDFlag := flag.String("i", "1", "The hit ID to spam, default is '1'")
	flag.Parse()

	numOfWorkers := int(*numOfWorkersFlag)
	hitID := *hitIDFlag

	fmt.Printf("using %d worker(s)", numOfWorkers)

	stopCh := make(chan struct{})
	var stopWg sync.WaitGroup

	var count int64

	for i := 0; i < numOfWorkers; i++ {
		stopWg.Add(1)

		go func() {
			defer stopWg.Done()

			flood := floodTrack(hitID)

			for {
				select {
				case <-stopCh:
					return
				default:
				}

				res, err := flood()
				if err != nil {
					logger.Printf("ERROR: %v\n", err)
					continue
				}

				if res.StatusCode != 200 {
					logger.Printf("STATUS: %v\n", res.StatusCode)
					continue
				}

				atomic.AddInt64(&count, 1)
			}
		}()
	}

	stopWg.Add(1)
	// Display the number of request per second.
	go func() {
		defer stopWg.Done()

		for {
			select {
			case <-stopCh:
				return
			default:
			}

			time.Sleep(1 * time.Second)

			v := atomic.SwapInt64(&count, 0)

			fmt.Printf("Requests per second: %d\n", v)
		}
	}()

	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, os.Interrupt, syscall.SIGTERM)
	<-shutdownSignal

	close(stopCh)

	stopWg.Wait()
}

func floodTrack(id string) func() (*http.Response, error) {
	client := http.Client{
		Timeout: 3 * time.Second,
	}

	type body struct {
		ID string `json:"id"`
	}

	var b body
	b.ID = id

	j, err := json.Marshal(b)
	if err != nil {
		panic(err)
	}

	return func() (*http.Response, error) {
		return client.Post("http://localhost:8001/analytics", "application/json", bytes.NewBuffer(j))
	}
}
