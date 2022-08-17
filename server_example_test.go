package udpsrv_test

import (
	"fmt"
	"time"

	"github.com/naeimc/udpsrv"
)

func ExampleServer() {

	queue := make(chan *udpsrv.Packet, 2)
	waitForWorker := true
	stopTimeout := 30 * time.Second

	server := udpsrv.NewServer(queue, waitForWorker, stopTimeout)

	go func() {
		time.Sleep(1 * time.Second)

		queue <- &udpsrv.Packet{
			Length:        len("This is a normal packet."),
			Data:          []byte("This is a normal packet."),
			PacketHandler: func(r udpsrv.Response, p *udpsrv.Packet) { fmt.Println(string(p.Data)) },
		}

		time.Sleep(1 * time.Second)

		queue <- &udpsrv.Packet{
			Error:        fmt.Errorf("This packet has an error."),
			ErrorHandler: func(err error) { fmt.Println(err) },
		}

		time.Sleep(1 * time.Second)

		server.Stop(fmt.Errorf("Stop requested by user."))
	}()

	fmt.Println("Starting Server.")
	server.Start()

	reason, _ := server.Report()
	fmt.Printf("Stopping Server: %s", reason)

	// Output:
	// Starting Server.
	// This is a normal packet.
	// This packet has an error.
	// Stopping Server: Stop requested by user.

}
