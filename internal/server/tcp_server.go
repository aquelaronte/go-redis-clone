package server

import (
	"fmt"
	"go-redis-clone/internal/resp/parser"
	"net"
)

// Run a redis clone server in port 6379
func Start() {
	listener, err := net.Listen("tcp", ":6379")

	if err != nil {
		panic(fmt.Errorf("error connecting tcp server %v", err))
	}

	defer listener.Close()
	fmt.Println("redis clone listening on port 6379")

	Serve(listener)
}

// Serve accepts connections from listener and dispatches them to the
// connection handler. It returns when the listener is closed.
func Serve(listener net.Listener) {
	for {
		conn, err := listener.Accept()

		if err != nil {
			return
		}

		go handleConnection(conn)
	}
}

// handle connections by parsing RESP messages, performing changes to the database in
// RAM, and responding to the connected client
func handleConnection(conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, 4096)

	var remaining []byte

	for {
		n, err := conn.Read(buffer)

		if err != nil {
			// connection closed
			break
		}

		// remaining buffer from the before call + new buffer
		receivedData := append(remaining, buffer[:n]...)
		remaining = nil // reset remaining data

		messages, r, err := parser.Parse(receivedData)

		if len(r) != 0 {
			remaining = r
		}

		for i := range messages {
			handleCommand(messages[i], conn)
		}
	}
}
