# udpsrv

UDP srv is a small UDP server written in Go.

## TODO
  - [ ] Improve Documentation

## Getting Started

### Installation
```
$ go get -u github.com/naeimc/udpsrv
```

### Server
The following is a simple echo server.
`server.go`
```go
package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/naeimc/udpsrv"
	"github.com/naeimc/udpsrv/udpsrvq"
)

const (
	timeFormat     = "2006-01-02 15:04:05"
	address        = "127.0.0.1:32000"
	bufferSize     = 1024
	waitForWorkers = true
	timeout        = 30 * time.Second
)

func main() {
	queue := udpsrvq.NewBasicQueue()
	queue.PacketHandler = packetHandler
	queue.ErrorHandler = errorHandler
	log("starting queue")
	go queue.Start()

	listener, err := udpsrv.NewListener(queue.Input(), address, bufferSize)
	if err != nil {
		log("cannot setup listener: %s", err)
	}
	log("starting listener on %s", listener.Address())
	go listener.Start()

	server := udpsrv.NewServer(queue.Output(), waitForWorkers, timeout)
	server.Register(queue.Stop)
	server.Register(listener.Stop)

	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt)
		<-signals
		server.Stop(fmt.Errorf("received interrupt"))
	}()

	log("starting server")
	server.Start()

	reason, err := server.Report()
	log("stopping server: %s", reason)
	if err != nil {
		log("error during shutdown: %s", err)
	}
}

func packetHandler(r udpsrv.Response, p *udpsrv.Packet) {
	length, err := r.Write(p.Data)
	log("<%s|%s> %d %d", p.LocalAddress, p.RemoteAddress, p.Length, length)
	if err != nil {
		log("<%s|%s> %s", p.LocalAddress, p.RemoteAddress, err)
	}
}

func errorHandler(err error) {
	log("%s", err)
}

func log(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "[%s] %s\n", timestamp(), fmt.Sprintf(format, a...))
}

func timestamp() string {
	return time.Now().UTC().Format(timeFormat)
}
```

Server Side Logs
```
[2022-08-17 15:08:12] starting queue
[2022-08-17 15:08:12] starting listener on 127.0.0.1:32000
[2022-08-17 15:08:12] starting server
[2022-08-17 15:08:52] <127.0.0.1:32000|127.0.0.1:60514> 13 13
[2022-08-17 15:09:03] stopping server: received interrupt
```

Client Side Logs
```
Sent: Hello, World!
Received: Hello, World!
```


## LICENSE
[MIT](./LICENSE)