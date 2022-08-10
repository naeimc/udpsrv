package udpsrv

import (
	"time"
)

type Server struct {
	// All the listeners used by the server.
	Listeners []*Listener

	// The Queue on which server receives bundles of data from the listeners.
	Queue Queue

	// How long to wait for workers to complete before forcing a shutdown.
	// If 0 the server waits forever.
	ShutdownTimeout time.Duration

	// The Done channel is channel (of size 2) that returns both the reason for the server shutdown
	// and an error if the server times out during shutdown.
	Done chan error

	stopped bool
}

// Starts the server.
func (s *Server) Listen() error {
	if err := s.setup(); err != nil {
		return err
	}
	s.run()
	return nil
}

// Closes the server while waiting for the workers to complete
// or a timeout to occur.
func (s *Server) Shutdown(reason error) {
	s.halt(reason, true)
}

// Closes the server wihout waiting for the workers to complete.
func (s *Server) Close(reason error) {
	s.halt(reason, false)
}

func (s *Server) setup() error {
	if s.Queue == nil {
		return ErrNoQueueAllocated{}
	}

	for _, listener := range s.Listeners {
		if err := listener.setup(s.Queue); err != nil {
			return err
		}
		go listener.run()
	}

	s.Done = make(chan error, 2)

	return nil
}

func (s *Server) run() {
	for !s.stopped {
		if s.Queue.Ready() {
			go s.work(s.Queue.Dequeue())
		}
	}
}

func (s *Server) work(b Bundle) {

	defer s.Queue.Notify()

	if b.Length > 0 {
		b.RequestHandler(
			ResponseWriter{
				connection: b.Connection,
				address:    b.RemoteAddress,
			},
			&Request{
				Timestamp:     b.Timestamp,
				LocalAddress:  b.LocalAddress,
				RemoteAddress: b.RemoteAddress,
				Length:        b.Length,
				Data:          b.Data[:b.Length],
			},
		)
	}

	if b.Error != nil {
		b.ErrorHandler(b.Error)
	}
}

func (s *Server) halt(reason error, waitingForWorkers bool) {
	s.stopped = true

	for _, listener := range s.Listeners {
		listener.Connection.Close()
	}

	var err error

	if waitingForWorkers {
		wait := make(chan bool, 1)
		go func() {
			s.Queue.Wait()
			wait <- true
		}()

		timeout := make(chan time.Time, 1)
		if s.ShutdownTimeout > 0 {
			go func() {
				timeout <- <-time.NewTicker(s.ShutdownTimeout).C
			}()
		}

		select {
		case <-timeout:
			err = ErrShutdownTimedOut{}
		case <-wait:
		}
	}

	s.Done <- reason
	s.Done <- err
	close(s.Done)
}

type ErrShutdownTimedOut struct{}

func (e ErrShutdownTimedOut) Error() string {
	return "shutdown timed out"
}

type ErrNoQueueAllocated struct{}

func (e ErrNoQueueAllocated) Error() string {
	return "no queue allocated"
}
