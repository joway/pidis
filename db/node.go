package db

import "fmt"

type Node struct {
	host string
	port string
}

func (n Node) Address() string {
	return fmt.Sprintf("%s:%s", n.host, n.port)
}

func (n Node) String() string {
	return n.Address()
}
