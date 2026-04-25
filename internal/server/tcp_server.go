package server

import (
	"fmt"
	"net"
)

func Start() {
	listener, err := net.Listen("tcp", ":6379")

	if err != nil {
		panic("error connecting tcp server")
	}

	defer listener.Close()
	fmt.Println("listening on port 6379")

	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Println("error connecting")
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, 4096)

	for {
		n, err := conn.Read(buffer)

		if err != nil {
			fmt.Println("connection closed")
			break
		}

		HandleCommand(string(buffer[:n]), conn)
	}
}
