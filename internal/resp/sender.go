package resp

import (
	"io"
)

func Sender(w io.Writer) func(msg []byte) {
	return func(msg []byte) {
		w.Write(msg)
	}
}
