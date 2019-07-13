package crdt

import (
	"time"
)

// LWWReg Last-Writer-Wins Registry
// a simple CRDT type that allows set/get of an
// arbitrary data type that ensures convergence
type LWWReg struct {
	id   string
	data []byte
	ts   time.Time
}

// NewLWWReg returns a new instance of the LWWReg with the given ID
func NewLWWReg(id string) LWWReg {
	return LWWReg{}
}
