package udpsrvq

import (
	"github.com/naeimc/udpsrv"
)

type BasicQueue struct {
	PacketHandler func(udpsrv.Response, *udpsrv.Packet)
	ErrorHandler  func(error)

	running bool
	input   chan *udpsrv.Packet
	output  chan *udpsrv.Packet
}

func NewBasicQueue() *BasicQueue {
	return &BasicQueue{
		input:  make(chan *udpsrv.Packet, 1),
		output: make(chan *udpsrv.Packet, 1),
	}
}

func (q *BasicQueue) Start() {
	for q.running = true; q.running; {
		select {
		case p := <-q.input:
			if q.PacketHandler != nil {
				p.PacketHandler = q.PacketHandler
			}
			if q.ErrorHandler != nil {
				p.ErrorHandler = q.ErrorHandler
			}
			q.output <- p
		default:
		}
	}
}

func (q *BasicQueue) Stop(err error) error {
	q.running = false
	return nil
}

func (q *BasicQueue) Input() chan<- *udpsrv.Packet {
	return q.input
}

func (q *BasicQueue) Output() <-chan *udpsrv.Packet {
	return q.output
}
