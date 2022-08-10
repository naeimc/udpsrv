# udpsrv

UDP srv is a small UDP server written in Go (Golang).

## Contents
- [udpsrv](#udpsrv)
    - [Contents](#contents)
    - [Getting Started](#getting-started)
        - [Installation](#installation)
        - [Server](#server)
        - [Command Line](#command-line)

## Getting Started
### Installation
```sh
$ go get -u github.com/naeimc/udpsrv
```

### Server
`server.go`
```go
package main

import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "runtime"
    "time"
    
    "github.com/naeimc/udpsrv"
)

func main() {
    server := udpsrv.NewServer(udpsrv.NewBasicQueue(16))

    listener := udpsrv.Listener{
        Address:        "127.0.0.1:49000",
        BufferSize:     1024,
        InitialHandler: func(b udpsrv.Bundle) { server.Queue.Enqueue(b) },
        ErrorHandler:   func(err error) { log.Printf("%s", err) },
        PacketHandler:  func(r udpsrv.Responder, p *udpsrv.Packet) {
            length, err := r.Write(r.Data)
            if err != nil {
                log.Printf("<%s> %d %d '%s' %s", p.RemoteAddress, p.Length, length, string(p.Data), err)
            } else {
                log.Printf("<%s> %d %d '%s'", p.RemoteAddress, p.Length, length, string(p.Data))
            }
        },
    }
    server.Listeners = append(server.Listeners, listener)
    log.Printf("listener setup on: %s", listener.Address)

    signals := make(chan os.Signal, 1)
    signal.Notify(signals, os.Interrupt, os.Kill)
    go func() {
        server.Halt(fmt.Errorf("os signal received: %s", <-signals), true, 10 * time.Second)
    }()

    log.Printf("starting server")
    if err := server.Listen(); err != nil {
        log.Printf("error during setup: %s", err)
        os.Exit(1)
    }

    log.Printf("stopping server: %s", <-server.Done)
    if err := <-server.Done; err != nil {
        log.Printf("error during stop: %s", err)
        os.Exit(1)
    }
}
```

### Command Line
Server Side:
```sh
$ go build server.go
$ ./server
2022/08/07 16:05:41 listener setup on: 127.0.0.1:49000
2022/08/07 16:05:41 starting server
2022/08/07 16:05:52 <127.0.0.1:53761> 11 11 'Hello, One!'
2022/08/07 16:06:00 <127.0.0.1:53762> 11 11 'Hello, Two!'
2022/08/07 16:06:06 <127.0.0.1:53763> 13 13 'Hello, Three!'
2022/08/07 16:06:12 <127.0.0.1:53767> 12 12 'Hello, Four!'
2022/08/07 16:06:18 <127.0.0.1:58945> 12 12 'Hello, Five!'
2022/08/07 16:06:22 stopping server: os signal received: interrupt
```

Client Side:
```sh
$ ./client "Hello, One!"
2022/08/07 16:05:52 (11) Hello, One!

$ ./client  "Hello, Two!" 
2022/08/07 16:06:00 (11) Hello, Two!

$ ./client  "Hello, Three!" 
2022/08/07 16:06:06 (13) Hello, Three!

$ ./client  "Hello, Four!"  
2022/08/07 16:06:12 (12) Hello, Four!

$ ./client  "Hello, Five!" 
2022/08/07 16:06:18 (12) Hello, Five!
```

## LICENSE
[MIT](./LICENSE)