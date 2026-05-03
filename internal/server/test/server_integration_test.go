package server_test

import (
	"fmt"
	"go-redis-clone/internal/server"
	"net"
	"strings"
	"testing"
	"time"
)

func startTestServer(t *testing.T) (string, func()) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test listener: %v", err)
	}

	go server.Serve(listener)

	return listener.Addr().String(), func() { listener.Close() }
}

func dial(t *testing.T, addr string) net.Conn {
	t.Helper()

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("failed to dial test server: %v", err)
	}

	return conn
}

func sendCommand(t *testing.T, conn net.Conn, args ...string) string {
	t.Helper()

	var b strings.Builder
	fmt.Fprintf(&b, "*%d\r\n", len(args))
	for _, arg := range args {
		fmt.Fprintf(&b, "$%d\r\n%s\r\n", len(arg), arg)
	}

	if _, err := conn.Write([]byte(b.String())); err != nil {
		t.Fatalf("failed to write command: %v", err)
	}

	return readResponse(t, conn)
}

func sendRaw(t *testing.T, conn net.Conn, raw string) string {
	t.Helper()

	if _, err := conn.Write([]byte(raw)); err != nil {
		t.Fatalf("failed to write raw payload: %v", err)
	}

	return readResponse(t, conn)
}

func readResponse(t *testing.T, conn net.Conn) string {
	t.Helper()

	conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	return string(buf[:n])
}

func uniqueKey(t *testing.T, name string) string {
	t.Helper()
	return fmt.Sprintf("test:%s:%d", name, time.Now().UnixNano())
}

func TestPing(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	t.Run("no arguments returns PONG", func(t *testing.T) {
		conn := dial(t, addr)
		defer conn.Close()

		got := sendCommand(t, conn, "PING")
		want := "+PONG\r\n"
		if got != want {
			t.Errorf("expected %q, got %q", want, got)
		}
	})

	t.Run("with message echoes payload", func(t *testing.T) {
		conn := dial(t, addr)
		defer conn.Close()

		got := sendCommand(t, conn, "PING", "hello")
		if got != "hello" {
			t.Errorf("expected %q, got %q", "hello", got)
		}
	})

	t.Run("too many arguments returns error", func(t *testing.T) {
		conn := dial(t, addr)
		defer conn.Close()

		got := sendCommand(t, conn, "PING", "a", "b")
		if !strings.HasPrefix(got, "-ERR") {
			t.Errorf("expected error reply, got %q", got)
		}
	})
}

func TestSetAndGet(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	t.Run("set then get returns stored value", func(t *testing.T) {
		conn := dial(t, addr)
		defer conn.Close()

		key := uniqueKey(t, "setget")

		if got := sendCommand(t, conn, "SET", key, "world"); got != "+OK\r\n" {
			t.Errorf("SET expected +OK, got %q", got)
		}

		got := sendCommand(t, conn, "GET", key)
		want := "$5\r\nworld\r\n"
		if got != want {
			t.Errorf("GET expected %q, got %q", want, got)
		}
	})

	t.Run("get missing key returns nil bulk string", func(t *testing.T) {
		conn := dial(t, addr)
		defer conn.Close()

		got := sendCommand(t, conn, "GET", uniqueKey(t, "missing"))
		if got != "$-1\r\n" {
			t.Errorf("expected nil bulk string, got %q", got)
		}
	})

	t.Run("set with wrong arity returns error", func(t *testing.T) {
		conn := dial(t, addr)
		defer conn.Close()

		got := sendCommand(t, conn, "SET", "onlykey")
		if !strings.HasPrefix(got, "-ERR") {
			t.Errorf("expected error reply, got %q", got)
		}
	})

	t.Run("get with wrong arity returns error", func(t *testing.T) {
		conn := dial(t, addr)
		defer conn.Close()

		got := sendCommand(t, conn, "GET", "a", "b")
		if !strings.HasPrefix(got, "-ERR") {
			t.Errorf("expected error reply, got %q", got)
		}
	})

	t.Run("set overwrites previous value", func(t *testing.T) {
		conn := dial(t, addr)
		defer conn.Close()

		key := uniqueKey(t, "overwrite")
		sendCommand(t, conn, "SET", key, "first")
		sendCommand(t, conn, "SET", key, "second")

		got := sendCommand(t, conn, "GET", key)
		want := "$6\r\nsecond\r\n"
		if got != want {
			t.Errorf("expected %q, got %q", want, got)
		}
	})
}

func TestDel(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	t.Run("delete existing key returns 1", func(t *testing.T) {
		conn := dial(t, addr)
		defer conn.Close()

		key := uniqueKey(t, "del-existing")
		sendCommand(t, conn, "SET", key, "v")

		if got := sendCommand(t, conn, "DEL", key); got != ":1\r\n" {
			t.Errorf("expected :1, got %q", got)
		}

		if got := sendCommand(t, conn, "GET", key); got != "$-1\r\n" {
			t.Errorf("expected key to be gone, got %q", got)
		}
	})

	t.Run("delete missing key returns 0", func(t *testing.T) {
		conn := dial(t, addr)
		defer conn.Close()

		got := sendCommand(t, conn, "DEL", uniqueKey(t, "del-missing"))
		if got != ":0\r\n" {
			t.Errorf("expected :0, got %q", got)
		}
	})

	t.Run("delete with wrong arity returns error", func(t *testing.T) {
		conn := dial(t, addr)
		defer conn.Close()

		got := sendCommand(t, conn, "DEL")
		if !strings.HasPrefix(got, "-ERR") {
			t.Errorf("expected error reply, got %q", got)
		}
	})
}

func TestExpire(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	t.Run("expire on existing key returns 1", func(t *testing.T) {
		conn := dial(t, addr)
		defer conn.Close()

		key := uniqueKey(t, "exp-existing")
		sendCommand(t, conn, "SET", key, "v")

		if got := sendCommand(t, conn, "EXPIRE", key, "60"); got != ":1\r\n" {
			t.Errorf("expected :1, got %q", got)
		}
	})

	t.Run("expire on missing key returns 0", func(t *testing.T) {
		conn := dial(t, addr)
		defer conn.Close()

		got := sendCommand(t, conn, "EXPIRE", uniqueKey(t, "exp-missing"), "10")
		if got != ":0\r\n" {
			t.Errorf("expected :0, got %q", got)
		}
	})

	t.Run("expire with non-integer seconds returns error", func(t *testing.T) {
		conn := dial(t, addr)
		defer conn.Close()

		key := uniqueKey(t, "exp-bad")
		sendCommand(t, conn, "SET", key, "v")

		got := sendCommand(t, conn, "EXPIRE", key, "not-a-number")
		if !strings.HasPrefix(got, "-ERR") {
			t.Errorf("expected error reply, got %q", got)
		}
	})

	t.Run("expired key is no longer readable", func(t *testing.T) {
		conn := dial(t, addr)
		defer conn.Close()

		key := uniqueKey(t, "exp-elapsed")
		sendCommand(t, conn, "SET", key, "v")
		sendCommand(t, conn, "EXPIRE", key, "1")

		// TTL has second resolution: EXPIRE 1 stores expiresAt = Unix()+1,
		// and GET keeps the value while Unix() <= expiresAt. Wait until the
		// next-next second tick is guaranteed to have elapsed.
		time.Sleep(2100 * time.Millisecond)

		if got := sendCommand(t, conn, "GET", key); got != "$-1\r\n" {
			t.Errorf("expected expired key to return nil, got %q", got)
		}
	})

	t.Run("expire with wrong arity returns error", func(t *testing.T) {
		conn := dial(t, addr)
		defer conn.Close()

		got := sendCommand(t, conn, "EXPIRE", "onlykey")
		if !strings.HasPrefix(got, "-ERR") {
			t.Errorf("expected error reply, got %q", got)
		}
	})
}

func TestCommand(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	conn := dial(t, addr)
	defer conn.Close()

	got := sendCommand(t, conn, "COMMAND")
	if got != ":1\r\n" {
		t.Errorf("expected :1, got %q", got)
	}
}

func TestUnknownCommand(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	conn := dial(t, addr)
	defer conn.Close()

	got := sendCommand(t, conn, "NOTREAL")
	if !strings.Contains(got, "unknown command") {
		t.Errorf("expected unknown command error, got %q", got)
	}
}

func TestCaseInsensitiveCommands(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	conn := dial(t, addr)
	defer conn.Close()

	if got := sendCommand(t, conn, "ping"); got != "+PONG\r\n" {
		t.Errorf("lowercase ping expected +PONG, got %q", got)
	}

	if got := sendCommand(t, conn, "PiNg"); got != "+PONG\r\n" {
		t.Errorf("mixed-case ping expected +PONG, got %q", got)
	}
}

func TestPipelinedCommands(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	conn := dial(t, addr)
	defer conn.Close()

	key := uniqueKey(t, "pipeline")
	pipelined := fmt.Sprintf(
		"*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$3\r\nbar\r\n*2\r\n$3\r\nGET\r\n$%d\r\n%s\r\n",
		len(key), key, len(key), key,
	)

	got := sendRaw(t, conn, pipelined)
	want := "+OK\r\n$3\r\nbar\r\n"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}
