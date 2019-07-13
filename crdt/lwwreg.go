package crdt

import (
	"time"
)

var (
	_ ReplicatedData = (*LWWRegistry)(nil)
)

// LWWRegistry Last-Writer-Wins Registry
// a simple CRDT type that allows set/get of an
// arbitrary data type that ensures convergence
type LWWRegistry struct {
	id    string
	data  []byte
	ts    int64
	clock Clock
}

// LWWRegState is a loaded LWWRegistry with its state loaded into memory
// type LWWRegState struct {
// 	id   string
// 	data []byte
// 	ts   time.Time
// }

// LWWRegDelta is a single delta operation for an LWWRegistry
type LWWRegDelta struct {
	data []byte
	ts   time.Time
}

// NewLWWRegistry returns a new instance of the LWWReg with the given ID
func NewLWWRegistry(id string, data []byte, ts int64, clock Clock) LWWRegistry {
	return LWWRegistry{
		id:    id,
		data:  data,
		ts:    ts,
		clock: clock,
	}
}

func (reg LWWRegistry) value() []byte {
	return reg.data
}

func (reg LWWRegistry) set(value []byte) LWWRegistry {
	return NewLWWRegistry(reg.id, value, reg.clock.Apply(), reg.clock)
}

func (reg LWWRegistry) setWithClock(value []byte, clock Clock) LWWRegistry {
	return NewLWWRegistry(reg.id, value, clock.Apply(), clock)
}

// Merge implements ReplicatedData interface
// Merge two LWWRegisty based on the order of the timestamp (ts),
// if they are equal, compare IDs
func (reg LWWRegistry) Merge(other ReplicatedData) (ReplicatedData, error) {
	otherReg, ok := other.(LWWRegistry)
	if !ok {
		return nil, ErrMismatchedMergeType
	}

	if otherReg.ts < reg.ts {
		return reg, nil
	} else if reg.ts < otherReg.ts {
		return otherReg, nil
	} else if otherReg.id < reg.id {
		return otherReg, nil
	} else {
		return reg, nil
	}
}
