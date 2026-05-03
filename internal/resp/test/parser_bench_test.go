package resp_test

import (
	"go-redis-clone/internal/resp/parser"
	"strings"
	"testing"
)

func BenchmarkParse(b *testing.B) {
	largeBulk := "$4096\r\n" + strings.Repeat("x", 4096) + "\r\n"
	pipelined := strings.Repeat("*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n", 10)

	cases := []struct {
		name  string
		input []byte
	}{
		{"Ping", []byte("*1\r\n$4\r\nPING\r\n")},
		{"Set", []byte("*3\r\n$3\r\nSET\r\n$5\r\nhello\r\n$5\r\nworld\r\n")},
		{"LargeBulkString", []byte(largeBulk)},
		{"Pipelined10", []byte(pipelined)},
		// Final byte missing — parser must report remainder without allocating a message.
		{"Fragmented", []byte("*3\r\n$3\r\nSET\r\n$5\r\nhello\r\n$5\r\nworl")},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.SetBytes(int64(len(tc.input)))
			for b.Loop() {
				parser.Parse(tc.input)
			}
		})
	}
}
