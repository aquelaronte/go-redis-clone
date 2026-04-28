package resp

import "bytes"

type Comparer struct {
	base Message
}

func NewComparer(base Message) *Comparer {
	return &Comparer{
		base: base,
	}
}

func (c *Comparer) Compare(target string) bool {
	return bytes.EqualFold(c.base.Bytes, []byte(target))
}

func (c *Comparer) RetrieveCommand(supportedCommands []string) string {
	for _, cmd := range supportedCommands {
		if c.Compare(cmd) {
			return cmd
		}
	}

	return ""
}

func Compare(base Message, target string) bool {
	return bytes.EqualFold(base.Bytes, []byte(target))
}
