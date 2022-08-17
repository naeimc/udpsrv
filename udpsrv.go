package udpsrv

import (
	"net"
	"time"
)

const (
	MaxUPDPayloadSize = 65527
	Network           = "udp"
)

// A Packet represents the data received from the client and
// the behaviour to be performed by the server.
type Packet struct {
	Timestamp     time.Time               // The time when the data was received
	LocalAddress  net.Addr                // The address of the listener
	RemoteAddress net.Addr                // The address of the client
	Length        int                     // The size of the data
	Data          []byte                  // The data sent by the client
	Error         error                   // An error that may be raised
	PacketHandler func(Response, *Packet) // How the packet should be handled by the server
	ErrorHandler  func(error)             // How the error should be handled by the server

	connection net.PacketConn
}

// A writer that can be used to respond to the client.
type Response struct {
	connection net.PacketConn
	address    net.Addr
}

// Writes data to the client.
func (r Response) Write(p []byte) (n int, err error) {
	return r.connection.WriteTo(p, r.address)
}
