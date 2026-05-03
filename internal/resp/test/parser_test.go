package resp_test

import (
	"bytes"
	"go-redis-clone/internal/resp"
	"go-redis-clone/internal/resp/parser"
	"testing"
)

func TestParse(t *testing.T) {
	t.Run("simple string", func(t *testing.T) {
		input := []byte("+OK\r\n")
		msgs, _, _ := parser.Parse(input)
		str := string(msgs[0].Bytes)

		if str != "OK" {
			t.Errorf("expected OK, got %s", str)
		}
	})

	t.Run("simple error", func(t *testing.T) {
		input := []byte("-Error: Invalid Message\r\n")
		msgs, _, _ := parser.Parse(input)
		str := string(msgs[0].Bytes)

		if str != "Error: Invalid Message" {
			t.Errorf("expected Error: Invalid Message, got %s", str)
		}
	})

	t.Run("integer", func(t *testing.T) {
		input := []byte(":102\r\n")
		msgs, _, _ := parser.Parse(input)
		str := string(msgs[0].Bytes)

		if str != "102" {
			t.Errorf("expected 102, got %s", str)
		}
	})

	t.Run("bulk string", func(t *testing.T) {
		input := []byte("$5\r\nhello\r\n")
		msgs, _, _ := parser.Parse(input)
		str := string(msgs[0].Bytes)

		if str != "hello" {
			t.Errorf("expected hello, got %q", str)
		}
	})

	t.Run("simple strings array", func(t *testing.T) {
		input := []byte("*2\r\n+OK\r\n+NO\r\n")
		msgs, _, _ := parser.Parse(input)

		str1 := string(msgs[0].Values[0].Bytes)
		str2 := string(msgs[0].Values[1].Bytes)

		if len(msgs[0].Values) != 2 {
			t.Errorf("expected 2 of length, got %d", len(msgs[0].Values))
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
		msgs, _, _ := parser.Parse(input)
		str1 := string(msgs[0].Values[0].Bytes)
		str2 := string(msgs[0].Values[1].Bytes)

		if len(msgs[0].Values) != 2 {
			t.Errorf("expected 2 of length, got %d", len(msgs[0].Values))
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
		msgs, _, _ := parser.Parse(input)

		if len(msgs[0].Values) != 2 {
			t.Errorf("expected 2 of length, got %d", len(msgs[0].Values))
		}

		str1 := string(msgs[0].Values[0].Bytes)
		str2 := string(msgs[0].Values[1].Bytes)

		if str1 != "15" {
			t.Errorf("expected 15 in position 0, got %q", str1)
		}

		if str2 != "5" {
			t.Errorf("expected 5 in position 1, got %q", str2)
		}
	})

	t.Run("mixed array", func(t *testing.T) {
		input := []byte("*3\r\n$4\r\nword\r\n:15\r\n+OK\r\n")
		msgs, _, _ := parser.Parse(input)

		if len(msgs[0].Values) != 3 {
			t.Errorf("expected 2 of length, got %d", len(msgs[0].Values))
		}

		str1 := string(msgs[0].Values[0].Bytes)
		str2 := string(msgs[0].Values[1].Bytes)
		str3 := string(msgs[0].Values[2].Bytes)

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

	t.Run("incomplete simple string (one and a half)", func(t *testing.T) {
		input := []byte("+OK\r\n+K")
		msgs, remaining, _ := parser.Parse(input)
		str := string(msgs[0].Bytes)

		if str != "OK" {
			t.Errorf("expected OK, got %s", str)
		}

		if string(remaining) != "+K" {
			t.Errorf("expected +K remaining, got %q", string(remaining))
		}
	})

	t.Run("incomplete simple string (half)", func(t *testing.T) {
		input := []byte("+O")
		msgs, remaining, _ := parser.Parse(input)

		if len(msgs) != 0 {
			t.Errorf("expected no messages, got %d", len(msgs))
		}

		if string(remaining) != "+O" {
			t.Errorf("expected +O remaining, got %q", string(remaining))
		}

		input = append(input, 'K', '\r', '\n')
		msgs, remaining, _ = parser.Parse(input)

		if len(msgs) != 1 {
			t.Errorf("expected 1 message, got %d", len(msgs))
		}

		if string(remaining) != "" {
			t.Errorf("expected empty remaining, got %q", string(remaining))
		}

		if string(msgs[0].Bytes) != "OK" {
			t.Errorf("expected OK, got %q", string(msgs[0].Bytes))
		}
	})

	t.Run("double simple string", func(t *testing.T) {
		input := []byte("+OK\r\n+KO\r\n")
		msgs, remaining, _ := parser.Parse(input)

		str1 := string(msgs[0].Bytes)
		str2 := string(msgs[1].Bytes)

		if str1 != "OK" {
			t.Errorf("expected OK, got %s", str1)
		}

		if str2 != "KO" {
			t.Errorf("expected KO, got %s", str2)
		}

		if remaining != nil {
			t.Errorf("expected no remaining, got %s", string(remaining))
		}
	})

	// --- Bulk string edges ---

	t.Run("empty bulk string", func(t *testing.T) {
		input := []byte("$0\r\n\r\n")
		msgs, remaining, err := parser.Parse(input)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(msgs) != 1 {
			t.Fatalf("expected 1 message, got %d", len(msgs))
		}

		if msgs[0].MessageType != resp.BulkStringMessageType {
			t.Errorf("expected BulkString, got %s", msgs[0].MessageType)
		}

		if len(msgs[0].Bytes) != 0 {
			t.Errorf("expected empty bytes, got %q", string(msgs[0].Bytes))
		}

		if remaining != nil {
			t.Errorf("expected no remaining, got %q", string(remaining))
		}
	})

	t.Run("bulk string with CRLF in payload", func(t *testing.T) {
		// 7-byte payload "ab\r\ncde" — embeds a CRLF that the length-prefix must skip past.
		input := []byte("$7\r\nab\r\ncde\r\n")
		msgs, _, err := parser.Parse(input)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(msgs) != 1 {
			t.Fatalf("expected 1 message, got %d", len(msgs))
		}

		if string(msgs[0].Bytes) != "ab\r\ncde" {
			t.Errorf("expected %q, got %q", "ab\r\ncde", string(msgs[0].Bytes))
		}
	})

	t.Run("bulk string fragmented body", func(t *testing.T) {
		input := []byte("$5\r\nhel")
		msgs, remaining, err := parser.Parse(input)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(msgs) != 0 {
			t.Errorf("expected 0 messages, got %d", len(msgs))
		}

		if !bytes.Equal(remaining, input) {
			t.Errorf("expected entire input as remaining, got %q", string(remaining))
		}
	})

	t.Run("bulk string missing terminating CRLF", func(t *testing.T) {
		input := []byte("$5\r\nhelloXX")
		_, _, err := parser.Parse(input)

		if err == nil {
			t.Errorf("expected error for missing CRLF terminator, got nil")
		}
	})

	t.Run("bulk string non-numeric length", func(t *testing.T) {
		input := []byte("$abc\r\nxxx\r\n")
		_, _, err := parser.Parse(input)

		if err == nil {
			t.Errorf("expected error for non-numeric length, got nil")
		}
	})

	t.Run("null bulk string", func(t *testing.T) {
		input := []byte("$-1\r\n")
		msgs, _, err := parser.Parse(input)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(msgs) != 1 {
			t.Fatalf("expected 1 null message, got %d", len(msgs))
		}

		if msgs[0].MessageType != resp.BulkStringMessageType {
			t.Errorf("expected BulkString type for null, got %s", msgs[0].MessageType)
		}

		if msgs[0].Bytes != nil {
			t.Errorf("expected nil bytes for null bulk string, got %q", string(msgs[0].Bytes))
		}
	})

	// --- Array edges ---

	t.Run("empty array", func(t *testing.T) {
		input := []byte("*0\r\n")
		msgs, _, err := parser.Parse(input)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(msgs) != 1 {
			t.Fatalf("expected 1 message, got %d", len(msgs))
		}

		if msgs[0].MessageType != resp.ArrayMessageType {
			t.Errorf("expected Array, got %s", msgs[0].MessageType)
		}

		if len(msgs[0].Values) != 0 {
			t.Errorf("expected 0 values, got %d", len(msgs[0].Values))
		}
	})

	t.Run("nested array", func(t *testing.T) {
		input := []byte("*2\r\n*2\r\n+a\r\n+b\r\n+c\r\n")
		msgs, _, err := parser.Parse(input)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(msgs) != 1 {
			t.Fatalf("expected 1 message, got %d", len(msgs))
		}

		outer := msgs[0]
		if outer.MessageType != resp.ArrayMessageType || len(outer.Values) != 2 {
			t.Fatalf("expected outer array of 2, got type=%s len=%d", outer.MessageType, len(outer.Values))
		}

		inner := outer.Values[0]
		if inner.MessageType != resp.ArrayMessageType || len(inner.Values) != 2 {
			t.Fatalf("expected inner array of 2, got type=%s len=%d", inner.MessageType, len(inner.Values))
		}

		if string(inner.Values[0].Bytes) != "a" || string(inner.Values[1].Bytes) != "b" {
			t.Errorf("expected inner [a, b], got [%q, %q]", string(inner.Values[0].Bytes), string(inner.Values[1].Bytes))
		}

		if string(outer.Values[1].Bytes) != "c" {
			t.Errorf("expected outer[1]=c, got %q", string(outer.Values[1].Bytes))
		}
	})

	t.Run("null array", func(t *testing.T) {
		input := []byte("*-1\r\n")
		msgs, _, err := parser.Parse(input)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(msgs) != 1 {
			t.Fatalf("expected 1 null message, got %d", len(msgs))
		}

		if msgs[0].MessageType != resp.ArrayMessageType {
			t.Errorf("expected Array type for null, got %s", msgs[0].MessageType)
		}

		if msgs[0].Values != nil {
			t.Errorf("expected nil values for null array, got len=%d", len(msgs[0].Values))
		}
	})

	t.Run("partial array waiting for more elements", func(t *testing.T) {
		input := []byte("*2\r\n+OK\r\n")
		msgs, remaining, err := parser.Parse(input)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(msgs) != 0 {
			t.Errorf("expected 0 messages (incomplete array), got %d", len(msgs))
		}

		if !bytes.Equal(remaining, input) {
			t.Errorf("expected entire input as remaining, got %q", string(remaining))
		}
	})

	// --- Integer / simple string edges ---

	t.Run("negative integer", func(t *testing.T) {
		input := []byte(":-100\r\n")
		msgs, _, err := parser.Parse(input)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(msgs) != 1 {
			t.Fatalf("expected 1 message, got %d", len(msgs))
		}

		if msgs[0].MessageType != resp.IntegerMessageType {
			t.Errorf("expected Integer type, got %s", msgs[0].MessageType)
		}

		if string(msgs[0].Bytes) != "-100" {
			t.Errorf("expected -100, got %q", string(msgs[0].Bytes))
		}
	})

	t.Run("empty simple string", func(t *testing.T) {
		input := []byte("+\r\n")
		msgs, _, err := parser.Parse(input)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(msgs) != 1 {
			t.Fatalf("expected 1 message, got %d", len(msgs))
		}

		if msgs[0].MessageType != resp.SimpleStringMessageType {
			t.Errorf("expected SimpleString, got %s", msgs[0].MessageType)
		}

		if len(msgs[0].Bytes) != 0 {
			t.Errorf("expected empty bytes, got %q", string(msgs[0].Bytes))
		}
	})

	// --- Top-level dispatch / error paths ---

	t.Run("empty input", func(t *testing.T) {
		msgs, remaining, err := parser.Parse([]byte{})

		if err == nil {
			t.Errorf("expected error for empty input, got nil")
		}

		if len(msgs) != 0 {
			t.Errorf("expected no messages, got %d", len(msgs))
		}

		if remaining != nil {
			t.Errorf("expected nil remaining, got %q", string(remaining))
		}
	})

	t.Run("single byte input", func(t *testing.T) {
		input := []byte("+")
		msgs, remaining, err := parser.Parse(input)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(msgs) != 0 {
			t.Errorf("expected no messages, got %d", len(msgs))
		}

		if !bytes.Equal(remaining, input) {
			t.Errorf("expected input as remaining, got %q", string(remaining))
		}
	})

	t.Run("unknown type byte", func(t *testing.T) {
		input := []byte("?garbage\r\n")
		_, _, err := parser.Parse(input)

		if err == nil {
			t.Errorf("expected error for unknown type byte, got nil")
		}
	})

	// --- Streaming ---

	t.Run("complete simple string followed by partial bulk", func(t *testing.T) {
		input := []byte("+OK\r\n$5\r\nhel")
		msgs, remaining, err := parser.Parse(input)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(msgs) != 1 {
			t.Fatalf("expected 1 message, got %d", len(msgs))
		}

		if string(msgs[0].Bytes) != "OK" {
			t.Errorf("expected OK, got %q", string(msgs[0].Bytes))
		}

		if string(remaining) != "$5\r\nhel" {
			t.Errorf("expected $5\\r\\nhel as remaining, got %q", string(remaining))
		}
	})

	// --- MessageType assertions (one per type) ---

	t.Run("message types are tagged correctly", func(t *testing.T) {
		cases := []struct {
			input []byte
			want  resp.MessageType
		}{
			{[]byte("+OK\r\n"), resp.SimpleStringMessageType},
			{[]byte("-ERR\r\n"), resp.SimpleErrorMessageType},
			{[]byte(":1\r\n"), resp.IntegerMessageType},
			{[]byte("$2\r\nOK\r\n"), resp.BulkStringMessageType},
			{[]byte("*1\r\n+OK\r\n"), resp.ArrayMessageType},
		}

		for _, c := range cases {
			msgs, _, err := parser.Parse(c.input)
			if err != nil {
				t.Errorf("input %q: unexpected error %v", string(c.input), err)
				continue
			}
			if len(msgs) != 1 {
				t.Errorf("input %q: expected 1 message, got %d", string(c.input), len(msgs))
				continue
			}
			if msgs[0].MessageType != c.want {
				t.Errorf("input %q: expected type %s, got %s", string(c.input), c.want, msgs[0].MessageType)
			}
		}
	})
}
