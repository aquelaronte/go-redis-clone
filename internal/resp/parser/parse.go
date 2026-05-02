package parser

import (
	"bytes"
	"errors"
	"go-redis-clone/internal/resp"
)

func Parse(received []byte) ([]resp.Message, []byte, error) {
	if len(received) == 0 {
		return nil, nil, errors.New("invalid message")
	}

	if len(received) == 1 {
		return nil, received, nil
	}

	var messages []resp.Message
	var remaining []byte
	var err error

	rawData := received

	for {
		msg, r, parseError := parse(rawData)

		if parseError != nil {
			err = parseError
			break
		}

		if msg != nil {
			messages = append(messages, *msg)
		}

		if bytes.Equal(rawData, r) || r == nil || len(r) == 0 {
			remaining = r

			if len(r) == 0 {
				remaining = nil
			}
			break
		}

		rawData = r
	}

	return messages, remaining, err
}
