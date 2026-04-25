package resp

import (
	"errors"
	"fmt"
	"strconv"
)

func Parse(received []byte) (*Message, []byte, error) {
	if len(received) == 0 {
		return nil, nil, errors.New("invalid message")
	}

	firstCharacter := received[0]

	switch firstCharacter {
	case '+':
		value, remaining := readUntilNextCRLF(received[1:])

		return &Message{
			MessageType: SimpleStringMessageType,
			Bytes:       value,
		}, remaining, nil

	case '-':
		value, remaining := readUntilNextCRLF(received[1:])

		return &Message{
			MessageType: SimpleErrorMessageType,
			Bytes:       value,
		}, remaining, nil

	case ':':
		value, remaining := readUntilNextCRLF(received[1:])

		return &Message{
			MessageType: IntegerMessageType,
			Bytes:       value,
		}, remaining, nil

	case '$':
		length, remaining, err := retrieveLength(received[1:])

		if err != nil {
			return nil, nil, fmt.Errorf("invalid message %v", err)
		}

		value := remaining[:length]
		valueRemaining := remaining[length+2:]

		return &Message{
			MessageType: BulkStringMessageType,
			Bytes:       value,
		}, valueRemaining, nil

	case '*':
		length, remaining, err := retrieveLength(received[1:])

		if err != nil {
			return nil, nil, fmt.Errorf("invalid message %v", err)
		}

		values := make([]Message, 0, length)
		lastValue := remaining

		for range length {
			message, remaining, err := Parse(lastValue)

			if err != nil {
				return nil, nil, fmt.Errorf("invalid message %v", err)
			}

			values = append(values, *message)
			lastValue = remaining
		}

		return &Message{
			MessageType: ArrayMessageType,
			Values:      values,
		}, lastValue, nil

	default:
		return nil, nil, errors.New("invalid message")
	}
}

func readUntilNextCRLF(received []byte) ([]byte, []byte) {
	for i := range received {
		if received[i] == '\r' && received[i+1] == '\n' {
			value := received[:i]
			remaining := received[i+2:] // skip next \r\n bytes

			return value, remaining
		}
	}

	return nil, nil
}

func retrieveLength(received []byte) (int, []byte, error) {
	firstPart, remaining := readUntilNextCRLF(received)
	strLength := string(firstPart)

	parsedLength, err := strconv.Atoi(strLength)

	return parsedLength, remaining, err
}
