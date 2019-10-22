package executor

import (
	"github.com/joway/loki"
	"github.com/joway/pidis/storage"
	"strings"
)

var (
	//system
	PING     = "PING"
	ECHO     = "ECHO"
	SHUTDOWN = "SHUTDOWN"
	QUIT     = "QUIT"
	SLAVEOF  = "SLAVEOF"

	//kv
	GET    = "GET"
	SET    = "SET"
	SETNX  = "SETNX"
	DEL    = "DEL"
	KEYS   = "KEYS"
	EXISTS = "EXISTS"
	INCR   = "INCR"
	TTL    = "TTL"
)

const (
	TypeSystem = 0
	TypeRead   = 1
	TypeWrite  = 2
)

var logger = loki.New("pidis:executor")

func New(cmd string) Executor {
	switch strings.ToUpper(cmd) {
	case QUIT, SHUTDOWN, PING, ECHO, SLAVEOF:
		return SystemExecutor{BaseExecutor{cmd: cmd, kind: TypeSystem}}
	case GET, KEYS, TTL, EXISTS:
		return KVExecutor{BaseExecutor{cmd: cmd, kind: TypeRead}}
	case SET, SETNX, DEL, INCR:
		return KVExecutor{BaseExecutor{cmd: cmd, kind: TypeWrite}}
	default:
		return SystemExecutor{BaseExecutor{cmd: cmd, kind: TypeSystem}}
	}
}

type Executor interface {
	IsWrite() bool
	Exec(store storage.Storage, args [][]byte) (*Result, error)
}

type BaseExecutor struct {
	Executor

	cmd  string
	kind int
}

func (c BaseExecutor) IsWrite() bool {
	return c.kind == TypeWrite
}
