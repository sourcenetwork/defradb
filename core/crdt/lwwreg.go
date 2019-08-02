package crdt

import (
	// "time"

	"bytes"

	"github.com/sourcenetwork/defradb/core"

	ds "github.com/ipfs/go-datastore"
	ipld "github.com/ipfs/go-ipld-format"
)

var (
	_ core.ReplicatedData = (*LWWRegister)(nil)
)

// LWWRegState is a loaded LWWRegister with its state loaded into memory
// type LWWRegState struct {
// 	id   string
// 	data []byte
// 	ts   time.Time
// }

// LWWRegDelta is a single delta operation for an LWWRegister
// TODO: Expand delta metadata (investigate if needed)
type LWWRegDelta struct {
	priority uint64
	data     []byte
}

func (delta *LWWRegDelta) GetPriority() uint64 {
	return delta.priority
}

func (delta *LWWRegDelta) SetPriority(prio uint64) {
	delta.priority = prio
}

// @TODO proto or cbor
func (delta *LWWRegDelta) Marshal() ([]byte, error) {
	return nil, nil
}

// @TODO
func LWWRegDeltaExtractorFn(node ipld.Node) (core.Delta, error) {
	return nil, nil
}

// LWWRegister Last-Writer-Wins Register
// a simple CRDT type that allows set/get of an
// arbitrary data type that ensures convergence
type LWWRegister struct {
	baseCRDT
	key  string
	data []byte
}

// NewLWWRegister returns a new instance of the LWWReg with the given ID
func NewLWWRegister(store ds.Datastore, namespace ds.Key, key string) LWWRegister {
	return LWWRegister{
		baseCRDT: newBaseCRDT(store, namespace),
		key:      key,
		// id:    id,
		// data:  data,
		// ts:    ts,
		// clock: clock,
	}
}

// Value gets the current register value
// RETURN STATE
func (reg LWWRegister) Value() ([]byte, error) {
	valueK := reg.valueKey(reg.key)
	return reg.store.Get(valueK)
}

// Set generates a new delta with the supplied value
// RETURN DELTA
func (reg LWWRegister) Set(value []byte) LWWRegDelta {
	// return NewLWWRegister(reg.id, value, reg.clock.Apply(), reg.clock)
	return LWWRegDelta{
		data: value,
	}
}

// RETURN DELTA
// func (reg LWWRegister) setWithClock(value []byte, clock Clock) LWWRegDelta {
// 	// return NewLWWRegister(reg.id, value, clock.Apply(), clock)
// 	return LWWRegDelta{
// 		data: value,
// 	}
// }

// Merge implements ReplicatedData interface
// Merge two LWWRegisty based on the order of the timestamp (ts),
// if they are equal, compare IDs
// MUTATE STATE
func (reg LWWRegister) Merge(delta core.Delta, id string) error {
	d, ok := delta.(*LWWRegDelta)
	if !ok {
		return core.ErrMismatchedMergeType
	}

	return reg.setValue(d.data, d.GetPriority())
}

func (reg LWWRegister) setValue(val []byte, priority uint64) error {
	curPrio, err := reg.getPriority(reg.key)
	if err != nil {
		return err
	}

	// if the current priority is higher ignore put
	// else if the current value is lexographically
	// greater than the new then ignore
	valueK := reg.valueKey(reg.key)
	if priority < curPrio {
		return nil
	} else if priority == curPrio {
		curValue, _ := reg.store.Get(valueK)
		if bytes.Compare(curValue, val) >= 0 {
			return nil
		}
	}

	err = reg.store.Put(valueK, val)
	if err != nil {
		return err
	}

	return reg.setPriority(reg.key, priority)
}
