package types

type Action int

const (
	None Action = iota
	Close
	Shutdown
)
