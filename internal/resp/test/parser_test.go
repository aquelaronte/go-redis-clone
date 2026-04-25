package resp_test

import (
	"go-redis-clone/internal/resp"
	"testing"
)

func TestParse(t *testing.T) {
	t.Run("simple string", func(t *testing.T) {
		input := []byte("+OK\r\n")
		msg, _, _ := resp.Parse(input)

		if msg.String != "OK" {
			t.Errorf("expected OK, got %s", msg.String)
		}
	})

	t.Run("simple error", func(t *testing.T) {
		input := []byte("-Error: Invalid Message\r\n")
		msg, _, _ := resp.Parse(input)

		if msg.String != "Error: Invalid Message" {
			t.Errorf("expected Error: Invalid Message, got %s", msg.String)
		}
	})

	t.Run("integer", func(t *testing.T) {
		input := []byte(":102\r\n")
		msg, _, _ := resp.Parse(input)

		if msg.Integer != 102 {
			t.Errorf("expected 102, got %d", msg.Integer)
		}
	})

	t.Run("bulk string", func(t *testing.T) {
		input := []byte("$5\r\nhello\r\n")
		msg, _, _ := resp.Parse(input)

		if msg.String != "hello" {
			t.Errorf("expected hello, got %q", msg.String)
		}
	})

	t.Run("simple strings array", func(t *testing.T) {
		input := []byte("*2\r\n+OK\r\n+NO\r\n")
		msg, _, _ := resp.Parse(input)

		if len(msg.Values) != 2 {
			t.Errorf("expected 2 of length, got %d", len(msg.Values))
		}

		if msg.Values[0].String != "OK" {
			t.Errorf("expected OK in position 0, got %q", msg.Values[0].String)
		}

		if msg.Values[1].String != "NO" {
			t.Errorf("expected NO in position 1, got %q", msg.Values[1].String)
		}
	})

	t.Run("bulk strings array", func(t *testing.T) {
		input := []byte("*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n")
		msg, _, _ := resp.Parse(input)

		if len(msg.Values) != 2 {
			t.Errorf("expected 2 of length, got %d", len(msg.Values))
		}

		if msg.Values[0].String != "hello" {
			t.Errorf("expected hello in position 0, got %q", msg.Values[0].String)
		}

		if msg.Values[1].String != "world" {
			t.Errorf("expected world in position 1, got %q", msg.Values[1].String)
		}
	})

	t.Run("integers array", func(t *testing.T) {
		input := []byte("*2\r\n:15\r\n:5\r\n")
		msg, _, _ := resp.Parse(input)

		if len(msg.Values) != 2 {
			t.Errorf("expected 2 of length, got %d", len(msg.Values))
		}

		if msg.Values[0].Integer != 15 {
			t.Errorf("expected 15 in position 0, got %q", msg.Values[0].String)
		}

		if msg.Values[1].Integer != 5 {
			t.Errorf("expected 5 in position 1, got %q", msg.Values[1].String)
		}
	})

	t.Run("mixed array", func(t *testing.T) {
		input := []byte("*3\r\n$4\r\nword\r\n:15\r\n+OK\r\n")
		msg, _, _ := resp.Parse(input)

		if len(msg.Values) != 3 {
			t.Errorf("expected 2 of length, got %d", len(msg.Values))
		}

		if msg.Values[0].String != "word" {
			t.Errorf("expected word in position 0, got %q", msg.Values[0].String)
		}

		if msg.Values[1].Integer != 15 {
			t.Errorf("expected 15 in position 1, got %q", msg.Values[1].String)
		}

		if msg.Values[2].String != "OK" {
			t.Errorf("expected OK in position 2, got %q", msg.Values[2].String)
		}
	})
}
