package udpsrv

import (
	"sync"
	"time"
)

// A udpsrv Server.
//
// The server is responsible for receiving the Packet from the queue
// and running the behaviour contained within.
type Server struct {
	running   bool
	stopping  bool
	wait      bool
	timeout   time.Duration
	waitGroup *sync.WaitGroup
	queue     <-chan *Packet
	stop      chan error
	report    chan error
	stopFuncs []func(error) error
}

// Make a new Server.
//
// queue is the queue to receive Packets on.
//
// wait decides whether the server should wait for a graceful shutdown.
//
// timeout is how long the server should wait for the shutdown before timing out.
func NewServer(queue <-chan *Packet, wait bool, timeout time.Duration) *Server {
	return &Server{
		wait:      wait,
		timeout:   timeout,
		queue:     queue,
		waitGroup: new(sync.WaitGroup),
		stop:      make(chan error, 1),
		report:    make(chan error, 2),
		stopFuncs: make([]func(error) error, 0),
	}
}

// Starts the Server.
func (s *Server) Start() {
	for s.running, s.stopping = true, false; s.running && !s.stopping; {
		select {
		case reason := <-s.stop:
			s.stopping = true
			s.report <- reason
			s.halt(reason)
			s.running = false
			return
		case p := <-s.queue:
			s.waitGroup.Add(1)
			go s.work(p)
		}
	}
}

// Registers a StopFunc.
//
// A StopFunc is run during shutdown.
func (s *Server) Register(stopFunc func(error) error) {
	s.stopFuncs = append(s.stopFuncs, stopFunc)
}

// Stops the Server.
func (s Server) Stop(reason error) error {
	s.stop <- reason
	return nil
}

// Returns the stop reason and if the server timed out during shutdown.
func (s Server) Report() (reason, err error) {
	return <-s.report, <-s.report
}

func (s Server) Running() bool {
	return s.running
}

func (s *Server) work(p *Packet) {
	defer s.waitGroup.Done()

	if p.Length > 0 && p.PacketHandler != nil {
		p.PacketHandler(Response{p.connection, p.RemoteAddress}, p)
	}

	if p.Error != nil && p.ErrorHandler != nil {
		p.ErrorHandler(p.Error)
	}
}

func (s *Server) halt(reason error) {

	wait := make(chan interface{}, 1)
	go func() {
		for _, stopFunc := range s.stopFuncs {
			stopFunc(reason)
		}
		s.waitGroup.Wait()
		wait <- nil
	}()

	if !s.wait {
		s.report <- nil
	} else if !(s.timeout > 0) {
		<-wait
		s.report <- nil
	} else {
		timeout := time.NewTimer(s.timeout)
		select {
		case <-wait:
			s.report <- nil
		case <-timeout.C:
			s.report <- ErrHaltTimedOut{}
		}
	}

	close(s.report)
}
