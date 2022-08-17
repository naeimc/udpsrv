package udpsrv_test

import (
	"fmt"
	"net"

	"github.com/naeimc/udpsrv"
)

func ExampleListener() {
	address := "127.0.0.1:32200"
	bufferSize := 1024

	queue := make(chan *udpsrv.Packet, 1)

	listener, err := udpsrv.NewListener(queue, address, bufferSize)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Starting Listener on %s.\n", address)
	go listener.Start()

	client, _ := net.Dial(udpsrv.Network, address)
	defer client.Close()
	client.Write([]byte("Hello, World!"))

	packet := <-queue
	fmt.Printf("Received \"%s\".\n", string(packet.Data))

	listener.Stop(fmt.Errorf("Stop Requested by User"))
	reason, _ := listener.Report()
	fmt.Printf("Stopping Listener: %s.\n", reason)

	// Output:
	// Starting Listener on 127.0.0.1:32200.
	// Received "Hello, World!".
	// Stopping Listener: Stop Requested by User.

}
