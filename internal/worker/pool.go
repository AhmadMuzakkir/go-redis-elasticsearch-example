package worker

import (
	"sync"
)

type Pool struct {
	bus    chan func()
	stop   chan struct{}
	wg     sync.WaitGroup
	closed bool
}

func NewWorkerPool() *Pool {
	return &Pool{
		stop: make(chan struct{}),
		bus:  make(chan func()),
	}
}

func (p *Pool) Start(workers int) {
	p.wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer p.wg.Done()
			for {
				select {
				case job := <-p.bus:
					if job != nil {
						job()
					}
				case <-p.stop:
					return
				}
			}
		}()
	}
}

func (p *Pool) Queue(job func()) {
	select {
	case p.bus <- job:
	case <-p.stop: // Make sure we're not accepting new job after the queue is stopped.
	}

}

func (p *Pool) Stop() {
	if p.closed {
		return
	}

	close(p.stop)

	close(p.bus)

	p.wg.Wait()

	p.closed = true
}
