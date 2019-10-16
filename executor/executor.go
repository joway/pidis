package executor

import (
	"github.com/joway/loki"
	"github.com/joway/pikv/common"
	"strings"
)

type Action int

var (
	//system
	PING     = "PING"
	ECHO     = "ECHO"
	SHUTDOWN = "SHUTDOWN"
	QUIT     = "QUIT"
	SLAVEOF  = "SLAVEOF"

	//kv
	GET = "GET"
	SET = "SET"
	DEL = "DEL"

	logger = loki.New("executor")
)

const (
	ActionNone Action = iota
	ActionUnknown
	ActionInvalidNumberOfArgs
	ActionInvalidSyntax
	ActionRuntimeError
	ActionClose
	ActionShutdown

	TypeSystem = 0
	TypeRead   = 1
	TypeWrite  = 2
)

func New(cmd string) Executor {
	switch strings.ToUpper(cmd) {
	case QUIT, SHUTDOWN, PING, ECHO, SLAVEOF:
		return SystemExecutor{BaseExecutor{cmd: cmd, kind: TypeSystem}}
	case GET:
		return KVExecutor{BaseExecutor{cmd: cmd, kind: TypeRead}}
	case SET, DEL:
		return KVExecutor{BaseExecutor{cmd: cmd, kind: TypeWrite}}
	default:
		return SystemExecutor{BaseExecutor{cmd: cmd, kind: TypeSystem}}
	}
}

type Executor interface {
	Type() int
	IsWrite() bool
	Exec(db common.Database, args [][]byte) (output []byte, action Action)
}

type BaseExecutor struct {
	Executor

	cmd  string
	kind int
}

func (c BaseExecutor) IsWrite() bool {
	return c.kind == TypeWrite
}
