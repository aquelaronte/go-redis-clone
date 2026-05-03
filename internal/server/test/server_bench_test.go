package server_test

import (
	"bufio"
	"fmt"
	"go-redis-clone/internal/server"
	"io"
	"net"
	"strings"
	"sync/atomic"
	"testing"
)

func startBenchServer(b *testing.B) (string, func()) {
	b.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatalf("failed to start bench listener: %v", err)
	}

	go server.Serve(listener)
	return listener.Addr().String(), func() { listener.Close() }
}

// Reads one RESP reply from r. Supports the reply types the server emits:
// simple string (+), error (-), integer (:), and bulk string ($).
func readReply(r *bufio.Reader) error {
	line, err := r.ReadString('\n')
	if err != nil {
		return err
	}
	if len(line) < 3 {
		return fmt.Errorf("short reply: %q", line)
	}
	switch line[0] {
	case '+', '-', ':':
		return nil
	case '$':
		size := 0
		fmt.Sscanf(line[1:], "%d", &size)
		if size < 0 {
			return nil
		}
		if _, err := io.CopyN(io.Discard, r, int64(size)+2); err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("unexpected reply byte: %q", line[0])
	}
}

func BenchmarkRoundtrip(b *testing.B) {
	addr, cleanup := startBenchServer(b)
	defer cleanup()

	pipelined10 := strings.Repeat("*1\r\n$4\r\nPING\r\n", 10)

	cases := []struct {
		name string
		run  func(b *testing.B, addr string)
	}{
		{"Ping", func(b *testing.B, addr string) {
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				b.Fatalf("dial: %v", err)
			}
			defer conn.Close()

			r := bufio.NewReader(conn)
			cmd := []byte("*1\r\n$4\r\nPING\r\n")
			b.SetBytes(int64(len(cmd)))

			for b.Loop() {
				if _, err := conn.Write(cmd); err != nil {
					b.Fatalf("write: %v", err)
				}
				if err := readReply(r); err != nil {
					b.Fatalf("read: %v", err)
				}
			}
		}},
		{"Set", func(b *testing.B, addr string) {
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				b.Fatalf("dial: %v", err)
			}
			defer conn.Close()

			r := bufio.NewReader(conn)
			cmd := []byte("*3\r\n$3\r\nSET\r\n$5\r\nhello\r\n$5\r\nworld\r\n")
			b.SetBytes(int64(len(cmd)))

			for b.Loop() {
				if _, err := conn.Write(cmd); err != nil {
					b.Fatalf("write: %v", err)
				}
				if err := readReply(r); err != nil {
					b.Fatalf("read: %v", err)
				}
			}
		}},
		{"Get", func(b *testing.B, addr string) {
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				b.Fatalf("dial: %v", err)
			}
			defer conn.Close()

			r := bufio.NewReader(conn)

			setCmd := []byte("*3\r\n$3\r\nSET\r\n$9\r\nbench:get\r\n$5\r\nworld\r\n")
			conn.Write(setCmd)
			if err := readReply(r); err != nil {
				b.Fatalf("set: %v", err)
			}

			cmd := []byte("*2\r\n$3\r\nGET\r\n$9\r\nbench:get\r\n")
			b.SetBytes(int64(len(cmd)))

			for b.Loop() {
				if _, err := conn.Write(cmd); err != nil {
					b.Fatalf("write: %v", err)
				}
				if err := readReply(r); err != nil {
					b.Fatalf("read: %v", err)
				}
			}
		}},
		// Sends 10 PINGs in one Write, then drains 10 replies. Measures
		// per-batch cost — divide ns/op by 10 for per-command throughput.
		{"Pipelined10", func(b *testing.B, addr string) {
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				b.Fatalf("dial: %v", err)
			}
			defer conn.Close()

			r := bufio.NewReader(conn)
			cmd := []byte(pipelined10)
			b.SetBytes(int64(len(cmd)))

			for b.Loop() {
				if _, err := conn.Write(cmd); err != nil {
					b.Fatalf("write: %v", err)
				}
				for range 10 {
					if err := readReply(r); err != nil {
						b.Fatalf("read: %v", err)
					}
				}
			}
		}},
		// Spawns one connection per goroutine to measure how the server
		// scales under concurrent clients.
		{"PingParallel", func(b *testing.B, addr string) {
			var dialCount atomic.Int64

			b.RunParallel(func(pb *testing.PB) {
				conn, err := net.Dial("tcp", addr)
				if err != nil {
					b.Errorf("dial: %v", err)
					return
				}
				defer conn.Close()

				dialCount.Add(1)
				r := bufio.NewReader(conn)
				cmd := []byte("*1\r\n$4\r\nPING\r\n")

				for pb.Next() {
					if _, err := conn.Write(cmd); err != nil {
						b.Errorf("write: %v", err)
						return
					}
					if err := readReply(r); err != nil {
						b.Errorf("read: %v", err)
						return
					}
				}
			})
		}},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			tc.run(b, addr)
		})
	}
}
