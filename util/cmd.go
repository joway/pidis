package util

import "strings"

func CommandToArgs(cmd string) [][]byte {
	cmds := strings.Split(cmd, " ")
	args := make([][]byte, 0)
	for _, cmd := range cmds {
		args = append(args, []byte(cmd))
	}
	return args
}

func ArgsToCommand(args [][]byte) string {
	cmd := ""
	for _, arg := range args {
		cmd += string(arg) + " "
	}
	return cmd
}
