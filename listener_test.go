package udpsrv_test

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/naeimc/udpsrv"
)

func TestListener(t *testing.T) {
	queue := make(chan *udpsrv.Packet, 1)
	address := "127.0.0.1:32100"

	var listener *udpsrv.Listener

	success := t.Run("NewListener", func(t *testing.T) {
		var err error
		listener, err = udpsrv.NewListener(queue, address, 1024)
		if err != nil {
			t.Fatalf("error received from udpsrv.NewListener(): %s", err)
		}
	})

	if !success {
		t.FailNow()
	}

	success = t.Run("Start", func(t *testing.T) {
		go listener.Start()
		timeout := time.NewTimer(5 * time.Second)
		for !listener.Running() {
			select {
			case <-timeout.C:
				t.Fatalf("timeout: listener did not start")
			default:
			}
		}
	})

	if !success {
		t.FailNow()
	}

	t.Run("Packet", func(t *testing.T) {
		message := "Hello, World!"
		client, _ := net.Dial(udpsrv.Network, address)
		t.Cleanup(func() { client.Close() })
		client.Write([]byte(message))

		timeout := time.NewTimer(5 * time.Second)
		select {
		case <-timeout.C:
			t.Fatalf("timeout: listener did not send packet")
		case <-queue:
		}
	})

	t.Run("Stop", func(t *testing.T) {
		expectedReason := fmt.Errorf("reason")
		if err := listener.Stop(expectedReason); err != nil {
			t.Fatalf("error: %s", err)
		}

		t.Run("Report", func(t *testing.T) {
			reason, _ := listener.Report()
			if expectedReason.Error() != reason.Error() {
				t.Errorf("incorrect reason received: expected %s, got %s", expectedReason, reason)
			}
		})
	})
}
