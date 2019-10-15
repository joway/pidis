package command

const (
	//types
	PING     = "PING"
	ECHO     = "ECHO"
	SHUTDOWN = "SHUTDOWN"
	QUIT     = "QUIT"

	//kv
	GET = "GET"
	SET = "SET"
	DEL = "DEL"
)

var (
	WRITE_COMMAND_LIST = []string{
		SET,
		DEL,
	}
)

var writeCommandsCache = make(map[string]bool)

func init() {
	for _, k := range WRITE_COMMAND_LIST {
		writeCommandsCache[k] = true
	}
}

func IsWriteCommand(cmd string) bool {
	return writeCommandsCache[cmd]
}
