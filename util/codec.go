package util

import "github.com/tidwall/redcon"

func MessageError(s string) []byte {
	return redcon.AppendError(nil, s)
}

func MessageNull() []byte {
	return redcon.AppendNull(nil)
}

func MessageOK() []byte {
	return redcon.AppendOK(nil)
}

func MessageInt(i int64) []byte {
	return redcon.AppendInt(nil, i)
}

func Message(b []byte) []byte {
	return redcon.AppendBulk(nil, b)
}

func MessageString(s string) []byte {
	return redcon.AppendString(nil, s)
}

func MessageArray(array [][]byte) []byte {
	out := redcon.AppendArray(nil, len(array))
	for _, item := range array {
		out = redcon.AppendBulk(out, item)
	}
	return out
}
