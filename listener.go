package udpsrv

import (
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

	// The PacketHandler is responsible for handling the bundles of data from the client.
	// The bundles are usually passed to the server's queue.
	// If nil a default PacketHandler is used that calls the Queue's Enqueue function.
	PacketHandler func(Bundle)

	// The RequestHandler processes requests and provides an writer for responses.
	RequestHandler func(ResponseWriter, *Request)

	// The ErrorHandler processes any errors returned by net.PacketConn.ReadFrom().
	// Errors returned by the ErrorHandler cause the server to panic.
	// If nil all errors are passed to the server.
	ErrorHandler func(error) error

	// The size of the data buffer used by the client.
	// If <= 0 the BufferSize is set to the the maximum size of a udp payload (65527).
	// Data not captured in the buffer is discarded.
	BufferSize int

	// The connection used by the listener.
	Connection net.PacketConn

	// The buffer used by the Listener.
	Buffer []byte
}

func (l *Listener) setup(queue Queue) error {

	connection, err := net.ListenPacket(network, l.Address)
	if err != nil {
		return err
	}
	l.Connection = connection

	if l.PacketHandler == nil {
		l.PacketHandler = func(b Bundle) { queue.Enqueue(b) }
	}

	if l.ErrorHandler == nil {
		l.ErrorHandler = func(err error) error { return err }
	}

	if l.BufferSize <= 0 {
		l.BufferSize = maxUDPPayloadSize
	}
	l.Buffer = make([]byte, l.BufferSize)

	return nil
}

func (l *Listener) run() {
	for {
		length, address, err := l.Connection.ReadFrom(l.Buffer)
		l.PacketHandler(Bundle{
			Timestamp:      time.Now().UTC(),
			Connection:     l.Connection,
			RequestHandler: l.RequestHandler,
			ErrorHandler:   l.ErrorHandler,
			RemoteAddress:  address,
			LocalAddress:   l.Connection.LocalAddr(),
			Length:         length,
			Data:           l.Buffer,
			Error:          err,
		})
	}
}
