package resp_test

import (
	"go-redis-clone/internal/resp"
	"testing"
)

func TestParse(t *testing.T) {
	t.Run("simple string", func(t *testing.T) {
		input := []byte("+OK\r\n")
		msg, _, _ := resp.Parse(input)
		str := string(msg.Bytes)

		if str != "OK" {
			t.Errorf("expected OK, got %s", str)
		}
	})

	t.Run("simple error", func(t *testing.T) {
		input := []byte("-Error: Invalid Message\r\n")
		msg, _, _ := resp.Parse(input)
		str := string(msg.Bytes)

		if str != "Error: Invalid Message" {
			t.Errorf("expected Error: Invalid Message, got %s", str)
		}
	})

	t.Run("integer", func(t *testing.T) {
		input := []byte(":102\r\n")
		msg, _, _ := resp.Parse(input)
		str := string(msg.Bytes)

		if str != "102" {
			t.Errorf("expected 102, got %s", str)
		}
	})

	t.Run("bulk string", func(t *testing.T) {
		input := []byte("$5\r\nhello\r\n")
		msg, _, _ := resp.Parse(input)
		str := string(msg.Bytes)

		if str != "hello" {
			t.Errorf("expected hello, got %q", str)
		}
	})

	t.Run("simple strings array", func(t *testing.T) {
		input := []byte("*2\r\n+OK\r\n+NO\r\n")
		msg, _, _ := resp.Parse(input)

		str1 := string(msg.Values[0].Bytes)
		str2 := string(msg.Values[1].Bytes)

		if len(msg.Values) != 2 {
			t.Errorf("expected 2 of length, got %d", len(msg.Values))
		}

		if str1 != "OK" {
			t.Errorf("expected OK in position 0, got %q", str1)
		}

		if str2 != "NO" {
			t.Errorf("expected NO in position 1, got %q", str2)
		}
	})

	t.Run("bulk strings array", func(t *testing.T) {
		input := []byte("*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n")
		msg, _, _ := resp.Parse(input)
		str1 := string(msg.Values[0].Bytes)
		str2 := string(msg.Values[1].Bytes)

		if len(msg.Values) != 2 {
			t.Errorf("expected 2 of length, got %d", len(msg.Values))
		}

		if str1 != "hello" {
			t.Errorf("expected hello in position 0, got %q", str1)
		}

		if str2 != "world" {
			t.Errorf("expected world in position 1, got %q", str2)
		}
	})

	t.Run("integers array", func(t *testing.T) {
		input := []byte("*2\r\n:15\r\n:5\r\n")
		msg, _, _ := resp.Parse(input)

		if len(msg.Values) != 2 {
			t.Errorf("expected 2 of length, got %d", len(msg.Values))
		}

		str1 := string(msg.Values[0].Bytes)
		str2 := string(msg.Values[1].Bytes)

		if str1 != "15" {
			t.Errorf("expected 15 in position 0, got %q", str1)
		}

		if str2 != "5" {
			t.Errorf("expected 5 in position 1, got %q", str2)
		}
	})

	t.Run("mixed array", func(t *testing.T) {
		input := []byte("*3\r\n$4\r\nword\r\n:15\r\n+OK\r\n")
		msg, _, _ := resp.Parse(input)

		if len(msg.Values) != 3 {
			t.Errorf("expected 2 of length, got %d", len(msg.Values))
		}

		str1 := string(msg.Values[0].Bytes)
		str2 := string(msg.Values[1].Bytes)
		str3 := string(msg.Values[2].Bytes)

		if str1 != "word" {
			t.Errorf("expected word in position 0, got %q", str1)
		}

		if str2 != "15" {
			t.Errorf("expected 15 in position 1, got %q", str2)
		}

		if str3 != "OK" {
			t.Errorf("expected OK in position 2, got %q", str3)
		}
	})
}
