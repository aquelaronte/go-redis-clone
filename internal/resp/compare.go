package resp

import "bytes"

func Compare(base Message, target string) bool {
	return bytes.EqualFold(base.Bytes, []byte(target))
}

func Comparer(base Message) func(target string) bool {
	return func(target string) bool {
		return Compare(base, target)
	}
}
