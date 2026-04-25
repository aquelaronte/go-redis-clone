package resp

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func Parse(received string) (*Message, string, error) {
	if received == "" {
		return nil, "", errors.New("invalid message")
	}

	firstCharacter := received[0]

	switch firstCharacter {
	case '+':
		str, remaining := readUntilNextCRLF(received[1:])

		return &Message{
			MessageType: SimpleStringMessageType,
			String:      str,
			Integer:     0,
		}, remaining, nil

	case '-':
		str, remaining := readUntilNextCRLF(received[1:])

		return &Message{
			MessageType: SimpleErrorMessageType,
			String:      str,
			Integer:     0,
		}, remaining, nil

	case ':':
		str, remaining := readUntilNextCRLF(received[1:])

		parsed, err := strconv.Atoi(str)

		if err != nil {
			return nil, "", fmt.Errorf("invalid message %v", err)
		}

		return &Message{
			MessageType: IntegerMessageType,
			Integer:     parsed,
			String:      "",
		}, remaining, nil

	case '$':
		_, index, _, err := retrieveLength(received[1:])

		if err != nil {
			return nil, "", fmt.Errorf("invalid message %v", err)
		}

		str, remaining := readUntilNextCRLF(received[index:])

		return &Message{
			MessageType: BulkStringMessageType,
			String:      str,
			Integer:     0,
		}, remaining, nil

	case '*':
		length, _, remaining, err := retrieveLength(received[1:])

		if err != nil {
			return nil, "", fmt.Errorf("invalid message %v", err)
		}

		var values []Message
		lastValue := remaining

		for range length {
			message, remaining, err := Parse(lastValue)

			if err != nil {
				return nil, "", fmt.Errorf("invalid message %v", err)
			}

			values = append(values, *message)
			lastValue = remaining
		}

		return &Message{
			MessageType: ArrayMessageType,
			Values:      values,
		}, remaining, nil

	default:
		return nil, "", errors.New("invalid message")
	}
}

func readUntilNextCRLF(received string) (string, string) {
	parts := strings.Split(received, "\r\n")

	return parts[0], strings.Join(parts[1:], "\r\n")
}

func retrieveLength(received string) (int, int, string, error) {
	strLength, remaining := readUntilNextCRLF(received)

	parsedLength, err := strconv.Atoi(strLength)

	if err != nil {
		return 0, 0, remaining, err
	}

	return parsedLength, len(strLength) + 3, remaining, nil
}
