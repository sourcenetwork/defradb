package crdt

import (
	// "time"

	"github.com/sourcenetwork/defradb/core"
)

var (
	_ core.ReplicatedData = (*LWWRegistry)(nil)
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
// TODO: Expand delta metadata (investigate if needed)
type LWWRegDelta struct {
	ts   int64
	data []byte
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

// Value gets the current register value
// RETURN STATE
func (reg LWWRegistry) Value() []byte {
	return reg.data
}

// Set generates a new delta with the supplied value
// RETURN DELTA
func (reg LWWRegistry) Set(value []byte) LWWRegDelta {
	// return NewLWWRegistry(reg.id, value, reg.clock.Apply(), reg.clock)
	return LWWRegDelta{
		ts:   reg.clock.Apply(),
		data: value,
	}
}

// RETURN DELTA
func (reg LWWRegistry) setWithClock(value []byte, clock Clock) LWWRegDelta {
	// return NewLWWRegistry(reg.id, value, clock.Apply(), clock)
	return LWWRegDelta{
		ts:   clock.Apply(),
		data: value,
	}
}

// Merge implements ReplicatedData interface
// Merge two LWWRegisty based on the order of the timestamp (ts),
// if they are equal, compare IDs
// MUTATE STATE
func (reg LWWRegistry) Merge(delta core.Delta, id string) error {
	d, ok := delta.(LWWRegDelta)
	if !ok {
		return core.ErrMismatchedMergeType
	}

	return reg.putValue(d.GetValue(), id, d.GetPriority())
}

// @TODO
func (reg LWWRegistry) putValue(val []byte, id string, priority uint64) error {
	return nil
}
