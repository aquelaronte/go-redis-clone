package resp

import "fmt"

type Message struct {
	MessageType MessageType
	Bytes       []byte
	Values      []Message
}

func (m *Message) ToRaw() string {
	switch m.MessageType {
	case SimpleStringMessageType:
		return fmt.Sprintf("+%s\r\n", string(m.Bytes))

	case SimpleErrorMessageType:
		return fmt.Sprintf("-%s\r\n", string(m.Bytes))

	case IntegerMessageType:
		return fmt.Sprintf(":%s\r\n", string(m.Bytes))

	case BulkStringMessageType:
		return fmt.Sprintf("$%d\r\n%s\r\n", len(m.Bytes), string(m.Bytes))

	case ArrayMessageType:
		var rawMsgs string

		for _, msg := range m.Values {
			rawMsgs = fmt.Sprintf("%s%s", rawMsgs, msg.ToRaw())
		}

		return fmt.Sprintf("*%d\r\n%s", len(m.Values), rawMsgs)

	default:
		return ""
	}
}
