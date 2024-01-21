# Websocket
This project serves as a simple websocket server library following the websocket protocol as defined in [RFC 6455](https://datatracker.ietf.org/doc/html/rfc6455).

## Installation
```
go get github.com/tofl/websocket
```

## API reference
The API reference can be viewed at [pkg.go.dev/github.com/tofl/websocket](https://pkg.go.dev/github.com/tofl/websocket).

## Usage
To start a new websocket server, just use the `NewServer` function. It takes as parameters the address, the path and the `onRead` function.

The `onRead` function, defined by the library user, is where frames received by the server are managed.

A new frame can be created using the `NewFrame` function. The library exposes the `Connection.Write` method to send a frame to a given client and `ConnectionPool.Broadcast` to send a frame to all connected clients.

## Example

```go
package main

import "github.com/tofl/websocket"

func onMessage(c *websocket.Connection, cp *websocket.ConnectionPool, f websocket.Frame) {
	ack := websocket.NewFrame(true, 1, false, []byte("Message received"))
	c.Write(ack)

	frame1 := websocket.NewFrame(false, 1, false, []byte("Hello"))
	continuation := websocket.NewFrame(true, 0, false, []byte(" everyone"))

	cp.Broadcast(frame1)
	cp.Broadcast(continuation)
}

func main() {
	websocket.NewServer(":8080", "/ws", onMessage)
}

```