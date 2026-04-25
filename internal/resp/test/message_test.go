package resp_test

import (
	"go-redis-clone/internal/resp"
	"testing"
)

func TestMessage(t *testing.T) {
	tests := []struct {
		Name     string
		Input    resp.Message
		Expected string
	}{
		{
			Name:     "simple string",
			Input:    resp.Message{MessageType: resp.SimpleStringMessageType, Bytes: []byte("OK")},
			Expected: "+OK\r\n",
		},
		{
			Name:     "simple error",
			Input:    resp.Message{MessageType: resp.SimpleErrorMessageType, Bytes: []byte("ERR")},
			Expected: "-ERR\r\n",
		},
		{
			Name:     "integer",
			Input:    resp.Message{MessageType: resp.IntegerMessageType, Bytes: []byte("12")},
			Expected: ":12\r\n",
		},
		{
			Name:     "bulk string",
			Input:    resp.Message{MessageType: resp.BulkStringMessageType, Bytes: []byte("hello")},
			Expected: "$5\r\nhello\r\n",
		},
		{
			Name: "bulk strings array",
			Input: resp.Message{MessageType: resp.ArrayMessageType, Values: []resp.Message{
				{
					MessageType: resp.BulkStringMessageType, Bytes: []byte("hello"),
				},
				{
					MessageType: resp.BulkStringMessageType, Bytes: []byte("world"),
				},
			}},
			Expected: "*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n",
		},
		{
			Name: "integers array",
			Input: resp.Message{MessageType: resp.ArrayMessageType, Values: []resp.Message{
				{
					MessageType: resp.IntegerMessageType, Bytes: []byte("15"),
				},
				{
					MessageType: resp.IntegerMessageType, Bytes: []byte("5"),
				},
			}},
			Expected: "*2\r\n:15\r\n:5\r\n",
		},
		{
			Name: "mixed array",
			Input: resp.Message{MessageType: resp.ArrayMessageType, Values: []resp.Message{
				{
					MessageType: resp.BulkStringMessageType, Bytes: []byte("word"),
				},
				{
					MessageType: resp.IntegerMessageType, Bytes: []byte("15"),
				},
				{
					MessageType: resp.SimpleStringMessageType, Bytes: []byte("OK"),
				},
			}},
			Expected: "*3\r\n$4\r\nword\r\n:15\r\n+OK\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := tt.Input.ToRaw()

			if result != tt.Expected {
				t.Errorf("expected %q, received %q", tt.Expected, result)
			}
		})
	}
}
