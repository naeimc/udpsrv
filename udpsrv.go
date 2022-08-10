package udpsrv

import (
	"net"
	"time"
)

const (
	network           = "udp"
	maxUDPPayloadSize = 65527
	defaultQueueSize  = 256
)

// Data received from the client.
type Request struct {
	Timestamp     time.Time
	LocalAddress  net.Addr
	RemoteAddress net.Addr
	Length        int
	Data          []byte
}

// A Writer that can be used to respond to the client.
type ResponseWriter struct {
	connection net.PacketConn
	address    net.Addr
}

// Writes data to the client.
func (w ResponseWriter) Write(p []byte) (int, error) {
	return w.connection.WriteTo(p, w.address)
}

// Contains all the data received from a client.
// As well as a timestamp, the connection, and the RequestHandler and ErrorHandler of the listener.
// This data is passed to the server to construct the Request and ResponseWriter.
type Bundle struct {
	Timestamp     time.Time
	Connection    net.PacketConn
	PacketHandler func(ResponseWriter, *Request)
	ErrorHandler  func(error)
	Length        int
	LocalAddress  net.Addr
	RemoteAddress net.Addr
	Data          []byte
	Error         error
}
