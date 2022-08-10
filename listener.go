package udpsrv

import (
	"errors"
	"net"
	"time"
)

// A Listener binds to a port and listens for data,
// the data is then passed on to server along with the RequestHandler and ErrorHandler
// to be processed by a worker.
type Listener struct {

	// The host and port of the server.
	// See net.ListenPacket for suitable address strings.
	Address string

	// The size of the data buffer used by the client.
	// If <= 0 the BufferSize is set to the the maximum size of a udp payload (65527).
	// Data not captured in the buffer is discarded.
	BufferSize int

	// The InitialHandler is responsible for handling the initial connection
	// by the client and passing it on the the queue.
	// The bundles are usually passed to the server's queue.
	// If nil a default PacketHandler is used that calls the Queue's Enqueue function.
	InitialHandler func(Bundle)

	// The PacketHandler processes requests and provides an writer for responses.
	PacketHandler func(Responder, *Packet)

	// The ErrorHandler processes any errors returned by net.PacketConn.ReadFrom().
	ErrorHandler func(error)

	// The connection used by the listener.
	Connection net.PacketConn

	// The buffer used by the Listener.
	Buffer []byte

	halted bool
}

func (l *Listener) setup(queue Queue) error {

	connection, err := net.ListenPacket(network, l.Address)
	if err != nil {
		return err
	}
	l.Connection = connection

	if l.PacketHandler == nil {
		l.InitialHandler = func(b Bundle) { queue.Enqueue(b) }
	}

	if l.ErrorHandler == nil {
		l.ErrorHandler = func(err error) {}
	}

	if l.BufferSize <= 0 {
		l.BufferSize = maxUDPPayloadSize
	}
	l.Buffer = make([]byte, l.BufferSize)

	return nil
}

func (l *Listener) run() {
	for !l.halted {
		length, address, err := l.Connection.ReadFrom(l.Buffer)

		if l.halted && errors.Is(err, net.ErrClosed) {
			return
		}

		l.InitialHandler(Bundle{
			Timestamp:     time.Now().UTC(),
			Connection:    l.Connection,
			PacketHandler: l.PacketHandler,
			ErrorHandler:  l.ErrorHandler,
			RemoteAddress: address,
			LocalAddress:  l.Connection.LocalAddr(),
			Length:        length,
			Data:          l.Buffer,
			Error:         err,
		})
	}
}

func (l *Listener) halt() error {
	l.halted = true
	return l.Connection.Close()
}
