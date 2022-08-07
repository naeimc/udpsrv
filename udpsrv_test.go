package udpsrv

import (
	"fmt"
	"net"
	"runtime"
	"time"
)

func Example() {

	queue := NewStdQueue(16, runtime.NumCPU(), time.Second)

	listener := &Listener{
		Address:        "127.0.0.1:49000",
		PacketHandler:  func(b Bundle) { queue.Enqueue(b) },
		RequestHandler: func(w ResponseWriter, r *Request) { fmt.Printf("(%d) %s\n", r.Length, string(r.Data)) },
		ErrorHandler: func(err error) error {
			fmt.Printf("%s", err)
			return nil
		},
		BufferSize: 1024,
	}

	server := Server{
		Listeners:       []*Listener{listener},
		Queue:           queue,
		ErrorHandler:    func(err error) { panic(err) },
		ShutdownTimeout: 10 * time.Second,
	}

	go func() {
		time.Sleep(8 * time.Second)
		server.Shutdown(fmt.Errorf("requested by user"))
	}()

	// simulate client
	go func() {
		time.Sleep(2 * time.Second)
		connection, err := net.Dial("udp", "127.0.0.1:49000")
		if err != nil {
			panic(err)
		}

		time.Sleep(2 * time.Second)
		connection.Write([]byte("Hello, One!"))

		time.Sleep(2 * time.Second)
		connection.Write([]byte("Hello, Two!"))
	}()

	fmt.Printf("listener setup on %s\n", server.Listeners[0].Address)
	fmt.Printf("starting server\n")

	if err := server.Listen(); err != nil {
		panic(err)
	}
	fmt.Printf("stopping server: %s\n", <-server.Done)

	if err := <-server.Done; err != nil {
		fmt.Printf("timeout error: %s\n", <-server.Done)
	} else {
		fmt.Print("timeout error: none\n")
	}

	// Output:
	// listener setup on 127.0.0.1:49000
	// starting server
	// (11) Hello, One!
	// (11) Hello, Two!
	// stopping server: requested by user
	// timeout error: none

}
