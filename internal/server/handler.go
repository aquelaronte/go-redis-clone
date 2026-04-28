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
	valuesLength := len(msg.Values)

	if comparer("get") {
		if valuesLength != 2 {
			send(fmt.Appendf(nil, "-ERR wrong number of arguments for '%s' command\r\n", "get"))
			return
		}

		key := msg.Values[1].Bytes
		value := core.GET(key)

		if value == nil {
			send([]byte("$-1\r\n"))
			return
		}

		send(value)
	} else if comparer("set") {
		if valuesLength != 3 {
			send(fmt.Appendf(nil, "-ERR wrong number of arguments for '%s' command\r\n", "set"))
			return
		}

		key := msg.Values[1].Bytes
		value := msg.Values[2].ToRaw()

		core.SET(key, []byte(value))

		send([]byte(":1\r\n"))
	} else if comparer("del") {
		if valuesLength != 2 {
			send(fmt.Appendf(nil, "-ERR wrong number of arguments for '%s' command\r\n", "del"))
			return
		}

		key := msg.Values[1].Bytes

		core.DEL(key)
		send([]byte(":1\r\n"))
	} else if comparer("ping") {
		switch valuesLength {
		case 1:
			send([]byte("+PONG\r\n"))
		case 2:
			value := msg.Values[1]

			send(fmt.Appendf(nil, "$%d\r\n%s\r\n", len(value.Bytes), string(value.Bytes)))
		default:
			send(fmt.Appendf(nil, "-ERR wrong number of arguments for '%s' command\r\n", "ping"))
		}
	} else if comparer("command") {
		send([]byte(":1\r\n"))
	} else {
		send(fmt.Appendf(nil, "-ERR unknown command '%s'", string(command.Bytes)))
	}
}
