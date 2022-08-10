package udpsrv

import (
	"time"
)

type Server struct {
	// All the listeners used by the server.
	Listeners []*Listener

	// The Queue on which server receives bundles of data from the listeners.
	Queue Queue

	// The Done channel is channel (of min size 2) that returns the reason for the server shutdown
	// an error if the server times out during shutdown and any errors when closing the connection.
	Done chan error

	stopped     bool
	haltChannel chan haltMessage
}

// Starts the server.
func (s *Server) Listen() error {
	if err := s.setup(); err != nil {
		return err
	}
	s.run()
	return nil
}

func (s *Server) Halt(reason error, wait bool, timeout time.Duration) {
	s.haltChannel <- haltMessage{reason, wait, timeout}
}

func (s *Server) setup() error {
	if s.Queue == nil {
		return ErrNoQueueAllocated{}
	}

	var listenerCount int
	for _, listener := range s.Listeners {
		if err := listener.setup(s.Queue); err != nil {
			return err
		}
		go listener.run()
		listenerCount++
	}

	s.haltChannel = make(chan haltMessage, 1)
	s.Done = make(chan error, 2)

	return nil
}

func (s *Server) run() {
	for !s.stopped {
		select {
		case h := <-s.haltChannel:
			s.halt(h)
		default:
			if s.Queue.Ready() {
				go s.work(s.Queue.Dequeue())
			}
		}
	}
}

func (s *Server) work(b Bundle) {
	defer s.Queue.Notify()

	if b.Length > 0 {
		b.PacketHandler(
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

func (s *Server) halt(h haltMessage) {
	s.stopped = true

	listenerErrors := make([]error, 0)
	for _, listener := range s.Listeners {
		listenerErrors = append(listenerErrors, listener.halt())
	}

	var haltError error
	if h.wait {
		wait := make(chan bool, 1)
		go func() {
			s.Queue.Wait()
			wait <- true
		}()

		timeout := make(chan time.Time, 1)
		if h.timeout > 0 {
			go func() {
				timeout <- <-time.NewTicker(h.timeout).C
			}()
		}

		select {
		case <-timeout:
			haltError = ErrShutdownTimedOut{}
		case <-wait:
		}
	}

	s.Done <- h.reason
	s.Done <- haltError
	for _, listenerError := range listenerErrors {
		s.Done <- listenerError
	}
	close(s.Done)
}

type haltMessage struct {
	reason  error
	wait    bool
	timeout time.Duration
}

type ErrShutdownTimedOut struct{}

func (e ErrShutdownTimedOut) Error() string {
	return "shutdown timed out"
}

type ErrNoQueueAllocated struct{}

func (e ErrNoQueueAllocated) Error() string {
	return "no queue allocated"
}
