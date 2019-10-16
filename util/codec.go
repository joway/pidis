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
