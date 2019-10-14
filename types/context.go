package types

import (
	"github.com/joway/pikv/storage"
)

type Context struct {
	Args    [][]byte
	Out     []byte
	Storage storage.Storage
}
