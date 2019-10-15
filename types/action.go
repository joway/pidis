package types

type Action int

const (
	ActionNone Action = iota
	ActionClose
	ActionShutdown
)
