package udpsrv_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/naeimc/udpsrv"
)

func TestServer(t *testing.T) {

	queue := make(chan *udpsrv.Packet, 1)

	server := udpsrv.NewServer(queue, false, 0)

	running := t.Run("Start", func(t *testing.T) {
		go server.Start()

		timeout := time.NewTimer(5 * time.Second)
		for !server.Running() {
			select {
			case <-timeout.C:
				t.Fatalf("timeout: server did not start")
			default:
			}
		}
	})

	if !running {
		t.FailNow()
	}

	var (
		activatedPacketHandler bool
		activatedErrorHandler  bool
		packetPointer          *udpsrv.Packet
		packetError            error
	)

	packet := &udpsrv.Packet{
		Length: 1,
		Data:   []byte{0},
		Error:  fmt.Errorf("error"),
		PacketHandler: func(r udpsrv.Response, p *udpsrv.Packet) {
			activatedPacketHandler = true
			packetPointer = p
		},
		ErrorHandler: func(err error) {
			activatedErrorHandler = true
			packetError = err
		},
	}

	queue <- packet

	t.Run("PacketHandler", func(t *testing.T) {
		timeout := time.NewTimer(5 * time.Second)
		for !activatedPacketHandler {
			select {
			case <-timeout.C:
				t.Fatalf("timeout: packet handler did not run")
			default:
			}
		}
		if packetPointer != packet {
			t.Fatalf("incorrect packet passed")
		}
	})

	t.Run("ErrorHandler", func(t *testing.T) {
		timeout := time.NewTimer(5 * time.Second)
		for !activatedErrorHandler {
			select {
			case <-timeout.C:
				t.Fatalf("timeout: error handler did not run")
			default:
			}
		}
		if packetError.Error() != "error" {
			t.Fatalf("incorrect error passed")
		}
	})

	t.Run("Stop", func(t *testing.T) {
		var activatedStopFunc bool
		server.Register(func(err error) error {
			activatedStopFunc = true
			return nil
		})

		expectedReason := fmt.Errorf("stop requsted by user")
		server.Stop(expectedReason)
		timeout := time.NewTimer(5 * time.Second)
		for server.Running() {
			select {
			case <-timeout.C:
				t.Fatalf("timeout: server did not stop")
			default:
			}
		}

		reason, _ := server.Report()
		if expectedReason != reason {
			t.Fatalf("incorrect stop reason received: expected %s, got %s", expectedReason, reason)
		}

		t.Run("StopFunc", func(t *testing.T) {
			if !activatedStopFunc {
				t.Fatalf("stop func not activated")
			}
		})
	})

}
