package udpsrv

import (
	"sync"
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

// A BasicQueue provides a bufferable queue and a waitgroup.
type BasicQueue struct {
	C         chan Bundle
	WaitGroup *sync.WaitGroup
}

func NewBasicQueue(capacity int) *BasicQueue {
	return &BasicQueue{
		C:         make(chan Bundle, capacity),
		WaitGroup: new(sync.WaitGroup),
	}
}

func (q BasicQueue) Ready() bool {
	return len(q.C) > 0
}

func (q BasicQueue) Enqueue(b Bundle) {
	q.C <- b
}

func (q BasicQueue) Dequeue() Bundle {
	q.WaitGroup.Add(1)
	return <-q.C
}

func (q BasicQueue) Notify() {
	q.WaitGroup.Done()
}

func (q BasicQueue) Wait() {
	q.WaitGroup.Wait()
}
