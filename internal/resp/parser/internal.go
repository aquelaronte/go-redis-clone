package parser

import (
	"errors"
	"fmt"
	"go-redis-clone/internal/resp"
)

func parse(received []byte) (*resp.Message, []byte, error) {
	firstCharacter := received[0]
	var message *resp.Message
	var remaining []byte

	switch firstCharacter {
	case '+':
		value, r := readUntilNextCRLF(received)

		if len(value) != 0 {
			message = &resp.Message{
				MessageType: resp.SimpleStringMessageType,
				Bytes:       value[1:],
			}
		}

		remaining = r
	case '-':
		value, r := readUntilNextCRLF(received)

		if len(value) != 0 {
			message = &resp.Message{
				MessageType: resp.SimpleErrorMessageType,
				Bytes:       value[1:],
			}
		}

		remaining = r
	case ':':
		value, r := readUntilNextCRLF(received)

		if len(value) != 0 {
			message = &resp.Message{
				MessageType: resp.IntegerMessageType,
				Bytes:       value[1:],
			}
		}

		remaining = r
	case '$':
		length, r, err := retrieveLength(received)

		if err != nil {
			return nil, nil, fmt.Errorf("invalid message %v", err)
		}

		// the message is complete if the remaining has the length bytes + CRLF or more
		isComplete := len(r) >= length+2

		if !isComplete {
			remaining = received
			break
		}

		// if message is complete, then should end with CLRF
		hasCRLF := r[length] == '\r' && r[length+1] == '\n'

		if !hasCRLF {
			return nil, nil, fmt.Errorf("invalid message, %q %q is missing. Found %q, %q instead", '\r', '\n', string(r[length]), string(r[length+1]))
		}

		message = &resp.Message{
			MessageType: resp.BulkStringMessageType,
			Bytes:       r[:length],
		}

		// If message has more characters, then is a remaining (TCP fragmented)
		if len(r) > length+2 {
			remaining = r[length+2:]
		}
	case '*':
		length, r, err := retrieveLength(received)

		if err != nil {
			return nil, nil, fmt.Errorf("invalid message %v", err)
		}

		// There should be always a remaining
		if len(r) == 0 {
			remaining = received
			break
		}

		values := make([]resp.Message, 0, length)
		lastValue := r

		for range length {
			parsedMsg, parsedRem, err := parse(lastValue)

			if err != nil {
				return nil, nil, fmt.Errorf("invalid message %v", err)
			}

			// If there are no message but no error, then
			// the received data will be the remaining
			if parsedMsg == nil {
				remaining = received
				break
			}

			lastValue = parsedRem // Walk to the next remaining
			values = append(values, *parsedMsg)
		}

		message = &resp.Message{
			MessageType: resp.ArrayMessageType,
			Values:      values,
		}

		remaining = lastValue

	default:
		return nil, nil, errors.New("invalid message")
	}

	return message, remaining, nil
}
