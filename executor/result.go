package executor

type Result struct {
	output []byte
	action Action
	err    error
}

func (r Result) Err() error {
	return r.err
}

func (r Result) Action() Action {
	return r.action
}

func (r Result) Output() []byte {
	return r.output
}
