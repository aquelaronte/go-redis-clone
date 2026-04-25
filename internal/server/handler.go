package server

import (
	"fmt"
	"go-redis-clone/internal/core"
	"go-redis-clone/internal/resp"
	"net"
)

func HandleCommand(received []byte, conn net.Conn) {
	msg, _, err := resp.Parse(received)

	if err != nil {
		fmt.Fprintf(conn, "-ERR: error parsing the command: %s\r\n", err)
		return
	}

	if msg.MessageType != resp.ArrayMessageType || len(msg.Values) == 0 {
		fmt.Fprintf(conn, "-ERR: invalid command: %s\r\n", err)
		return
	}

	command := msg.Values[0]

	switch command.String {
	case "GET":
		key := msg.Values[1].String

		fmt.Fprintf(conn, "+%s\r\n", core.GET(key))

		return
	case "SET":
		key := msg.Values[1].String
		value := msg.Values[2].String

		core.SET(key, value)

		fmt.Fprintf(conn, "+OK\r\n")

		return
	case "DEL":
		key := msg.Values[1].String

		core.DEL(key)

		fmt.Fprintf(conn, "+OK\r\n")
		return

	case "PING":
		fmt.Fprintf(conn, "+PONG\r\n")
	case "COMMAND":
		fmt.Fprintf(conn, "+OK\r\n")
		return
	default:
		fmt.Fprint(conn, "-ERR: invalid command")
	}

}
