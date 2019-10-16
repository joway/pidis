package db

import "fmt"

type Node struct {
	host string
	port string
}

func (n Node) String() string {
	return fmt.Sprintf("%s:%s", n.host, n.port)
}
