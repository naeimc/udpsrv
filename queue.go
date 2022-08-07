package udpsrv

import (
	"sync"
	"time"
)

// A Queue is responsible for handling the transfer of bundles between listeners and servers
// as well as managing the workers used by the server.
//
// Queues can be used to hand rate and worker limiting.
type Queue interface {
	Ready() bool     // Returns if the queue is ready to receive a bundle.
	Enqueue(Bundle)  // Add a bundle to the queue.
	Dequeue() Bundle // Remove a bundle from the queue and add a worker to the wait group.
	Notify()         // Remove a worker from the wait group.
	Wait()           // Wait for the wait group.
}

// A StdQueue provides a bufferable queue as well as an optional rate limiter and an optional worker limiter.
type StdQueue struct {
	C         chan Bundle
	WaitGroup *sync.WaitGroup
	Ticker    *time.Ticker
	Limit     int
	Current   int
}

func NewStdQueue(capacity, workerLimit int, rateLimit time.Duration) *StdQueue {

	var ticker *time.Ticker
	if rateLimit != 0 {
		ticker = time.NewTicker(rateLimit)
	}

	return &StdQueue{
		C:         make(chan Bundle, capacity),
		WaitGroup: new(sync.WaitGroup),
		Ticker:    ticker,
		Limit:     workerLimit,
	}
}

func (q StdQueue) Ready() bool {
	if len(q.C) == 0 {
		return false
	}

	if (q.Limit != 0) && (q.Current >= q.Limit) {
		return false
	}

	if (q.Ticker != nil) && (len(q.Ticker.C) == 0) {
		return false
	}

	return true
}

func (q StdQueue) Enqueue(b Bundle) {
	q.C <- b
}

func (q *StdQueue) Dequeue() Bundle {
	q.WaitGroup.Add(1)
	q.Current++

	if q.Ticker != nil {
		<-q.Ticker.C
	}

	return <-q.C
}

func (q *StdQueue) Notify() {
	q.WaitGroup.Done()
	q.Current--
}

func (q StdQueue) Wait() {
	q.WaitGroup.Wait()
}
