package udpsrv_test

import (
	"fmt"
	"net"
	"time"

	"github.com/naeimc/udpsrv"
	"github.com/naeimc/udpsrv/udpsrvq"
)

func Example() {

	var (
		address        = "127.0.0.1:32000"
		bufferSize     = 1024
		waitForWorkers = true
		stopTimeout    = 30 * time.Second
	)

	queue := udpsrvq.NewBasicQueue()
	queue.PacketHandler = func(r udpsrv.Response, p *udpsrv.Packet) {
		fmt.Printf("(Server) %d \"%s\"\n", p.Length, string(p.Data))
		r.Write(p.Data)
	}
	queue.ErrorHandler = func(err error) {
		fmt.Println(err)
	}

	fmt.Println("Starting Queue")
	go queue.Start()

	listener, err := udpsrv.NewListener(queue.Input(), address, bufferSize)
	if err != nil {
		panic(err)
	}

	fmt.Println("Starting Listener on", address)
	go listener.Start()

	server := udpsrv.NewServer(queue.Output(), waitForWorkers, stopTimeout)
	server.Register(queue.Stop)
	server.Register(listener.Stop)

	go func() {
		time.Sleep(5 * time.Second)
		server.Stop(fmt.Errorf("Stop Requested by User"))
	}()

	go func() {
		time.Sleep(time.Second)

		client, err := net.Dial(udpsrv.Network, address)
		if err != nil {
			panic(err)
		}
		defer client.Close()

		msg := "Hello, World!"
		fmt.Printf("(Client) Sending \"%s\"\n", msg)
		client.Write([]byte(msg))

		data := make([]byte, len(msg))
		client.Read(data)
		fmt.Printf("(Client) Received \"%s\"\n", msg)
	}()

	fmt.Println("Starting Server")
	server.Start()

	reason, _ := server.Report()
	fmt.Println("Stopping Server:", reason)

	// Output:
	// Starting Queue
	// Starting Listener on 127.0.0.1:32000
	// Starting Server
	// (Client) Sending "Hello, World!"
	// (Server) 13 "Hello, World!"
	// (Client) Received "Hello, World!"
	// Stopping Server: Stop Requested by User

}
