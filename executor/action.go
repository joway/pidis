package executor

type Action int

const (
	_ Action = iota
	ActionShutdown
	ActionConnClose
	ActionSlaveOf
)
