package resp

import (
	"fmt"
	"io"
)

type Sender struct {
	w io.Writer
}

func NewSender(w io.Writer) *Sender {
	return &Sender{
		w: w,
	}
}

/* ----------------------------- Main Functions ----------------------------- */
func (s *Sender) Send(bytes []byte) {
	s.w.Write(bytes)
}

func (s *Sender) SendMsg(msg string) {
	s.w.Write([]byte(msg))
}

/* ------------------------------ Send by Type ------------------------------ */
func (s *Sender) SendInteger(integer int) {
	s.SendMsg(fmt.Sprintf(":%d\r\n", integer))
}

func (s *Sender) SendBulkString(msg string) {
	s.SendMsg(fmt.Sprintf("$%d\r\n%s\r\n", len(msg), msg))
}

func (s *Sender) SendNil() {
	s.SendMsg("$-1\r\n")
}

func (s *Sender) SendError(msg string) {
	s.w.Write(fmt.Appendf(nil, "-ERR %s\r\n", msg))
}

/* --------------------------- Send by Error Type --------------------------- */
func (s *Sender) SendWrongNumberOfArguments(cmd string) {
	s.SendError(fmt.Sprintf("wrong number of arguments for '%s' command", cmd))
}

func (s *Sender) SendUnknownCommand(cmd string) {
	s.SendError(fmt.Sprintf("unknown command '%s'", cmd))
}
