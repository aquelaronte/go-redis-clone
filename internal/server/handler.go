package server

import (
	"go-redis-clone/internal/core"
	"go-redis-clone/internal/resp"
	"net"
)

func handleCommand(msg resp.Message, conn net.Conn) {
	sender := resp.NewSender(conn)

	if msg.MessageType != resp.ArrayMessageType || len(msg.Values) == 0 {
		sender.SendError("invalid command")
		return
	}

	command := msg.Values[0]
	comparer := resp.NewComparer(command)
	valuesLength := len(msg.Values)

	switch comparer.RetrieveCommand(SupportedCommands) {
	case "get":
		if valuesLength != 2 {
			sender.SendWrongNumberOfArguments("get")
			return
		}

		key := msg.Values[1].Bytes
		value := core.GET(key)

		if value == nil {
			sender.SendNil()
			return
		}

		sender.Send(value)
	case "set":
		if valuesLength != 3 {
			sender.SendWrongNumberOfArguments("set")
			return
		}

		key := msg.Values[1].Bytes
		value := msg.Values[2].ToRaw()

		core.SET(key, []byte(value))

		sender.SendInteger(1)
	case "del":
		if valuesLength != 2 {
			sender.SendWrongNumberOfArguments("del")
			return
		}

		key := msg.Values[1].Bytes

		core.DEL(key)
		sender.SendInteger(1)
	case "ping":
		switch valuesLength {
		case 1:
			sender.SendSimpleString("PONG")
		case 2:
			value := msg.Values[1]

			sender.SendMsg(string(value.Bytes))
		default:
			sender.SendWrongNumberOfArguments("ping")
		}
	case "command":
		sender.SendInteger(1)
	default:
		sender.SendUnknownCommand(string(command.Bytes))
	}
}
