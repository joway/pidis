package db

import (
	"github.com/rs/xid"
	"time"
)

const UIDSize = 12

type UID struct {
	id xid.ID
}

func NewUID() UID {
	return UID{
		id: xid.New(),
	}
}

func (u UID) Size() int {
	return UIDSize
}

func (u UID) Bytes() []byte {
	return u.id.Bytes()
}

func (u UID) String() string {
	return u.String()
}

func (u UID) Time() time.Time {
	return u.id.Time()
}

func (u UID) Timestamp() string {
	return string(u.id.Time().Unix())
}
