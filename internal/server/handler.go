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
	comparer := resp.Comparer(command)
	send := resp.Sender(conn)

	if comparer("get") {
		key := msg.Values[1].Bytes
		value := core.GET(key)

		if value == nil {
			send([]byte("$-1\r\n"))
			return
		}

		send(value)
	} else if comparer("set") {
		key := msg.Values[1].Bytes
		value := msg.Values[2].ToRaw()

		core.SET(key, []byte(value))

		send([]byte("+OK\r\n"))
	} else if comparer("del") {
		key := msg.Values[1].Bytes

		core.DEL(key)
		send([]byte("+OK\r\n"))
	} else if comparer("ping") {
		send([]byte("+PONG\r\n"))
	} else if comparer("command") {
		send([]byte("+OK\r\n"))
	}
}
