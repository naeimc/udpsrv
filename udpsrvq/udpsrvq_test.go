package udpsrvq_test

import (
	"testing"

	"github.com/naeimc/udpsrv"
	"github.com/naeimc/udpsrv/udpsrvq"
)

func TestBasicQueue(t *testing.T) {

	queue := udpsrvq.NewBasicQueue()
	queue.PacketHandler = func(r udpsrv.Response, p *udpsrv.Packet) {}
	queue.ErrorHandler = func(err error) {}
	go queue.Start()

	packetIn := &udpsrv.Packet{}

	queue.Input() <- packetIn
	packetOut := <-queue.Output()

	if packetIn != packetOut {
		t.Fatalf("incorrect packet ptr sent out")
	}

	if packetOut.PacketHandler == nil {
		t.Errorf("packet handler not set")
	}

	if packetOut.ErrorHandler == nil {
		t.Errorf("error handler not set")
	}

}
