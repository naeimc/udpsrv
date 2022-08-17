package udpsrv

import (
	"errors"
	"net"
	"time"
)

type Listener struct {
	running    bool
	buffer     []byte
	address    net.Addr
	connection net.PacketConn
	queue      chan<- *Packet
	report     chan error
}

func NewListener(queue chan<- *Packet, addr string, bufferSize int) (*Listener, error) {
	if queue == nil {
		return nil, ErrNoQueue{}
	}

	address, err := net.ResolveUDPAddr(Network, addr)
	if err != nil {
		return nil, err
	}

	connection, err := net.ListenPacket(Network, address.String())
	if err != nil {
		return nil, err
	}

	if bufferSize <= 0 {
		bufferSize = MaxUPDPayloadSize
	}

	return &Listener{
		address:    address,
		connection: connection,
		queue:      queue,
		buffer:     make([]byte, bufferSize),
		report:     make(chan error, 2),
	}, nil
}

// Start the Listener
func (l *Listener) Start() {
	for l.running = true; l.running; {
		length, address, err := l.connection.ReadFrom(l.buffer)
		buffer := make([]byte, length)
		copy(buffer, l.buffer)

		if errors.Is(err, net.ErrClosed) {
			if l.running {
				l.running = false
				l.report <- err
				l.report <- nil
				close(l.report)
			}
			break
		}

		packet := &Packet{
			Timestamp:     time.Now().UTC(),
			LocalAddress:  l.address,
			RemoteAddress: address,
			Length:        length,
			Data:          buffer,
			Error:         err,
			connection:    l.connection,
		}

		l.queue <- packet
	}
}

// Stop the Listener
func (l *Listener) Stop(reason error) error {
	if !l.running {
		return ErrListenerNotRunning{l.address.String()}
	}

	l.running = false
	l.report <- reason
	l.report <- l.connection.Close()
	close(l.report)
	return nil
}

// Returns the stop reason and if there was an error while closing the connection.
func (l Listener) Report() (reason, err error) {
	return <-l.report, <-l.report
}

func (l Listener) Running() bool {
	return l.running
}

func (l Listener) Address() net.Addr {
	return l.address
}
