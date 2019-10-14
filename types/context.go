package types

import (
	"github.com/joway/pikv/db"
)

type Context struct {
	Args [][]byte
	Out  []byte
	DB   *db.Database
}
