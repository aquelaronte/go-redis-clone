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
			String:      string(value),
			Integer:     0,
		}, remaining, nil

	case '-':
		value, remaining := readUntilNextCRLF(received[1:])

		return &Message{
			MessageType: SimpleErrorMessageType,
			String:      string(value),
			Integer:     0,
		}, remaining, nil

	case ':':
		value, remaining := readUntilNextCRLF(received[1:])

		parsed, err := strconv.Atoi(string(value))

		if err != nil {
			return nil, nil, fmt.Errorf("invalid message %v", err)
		}

		return &Message{
			MessageType: IntegerMessageType,
			Integer:     parsed,
			String:      "",
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
			String:      string(value),
			Integer:     0,
		}, valueRemaining, nil

	case '*':
		length, remaining, err := retrieveLength(received[1:])

		if err != nil {
			return nil, nil, fmt.Errorf("invalid message %v", err)
		}

		var values []Message
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
		}, remaining, nil

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
