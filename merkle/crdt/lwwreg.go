package crdt

import (
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/core/crdt"
	corecrdt "github.com/sourcenetwork/defradb/core/crdt"
	"github.com/sourcenetwork/defradb/merkle/clock"
	"github.com/sourcenetwork/defradb/store"

	ds "github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log"
)

// MerkleLWWRegister is a MerkleCRDT implementation of the LWWRegister
// using MerkleClocks
type MerkleLWWRegister struct {
	*baseMerkleCRDT
	// core.ReplicatedData

	reg   corecrdt.LWWRegister
	clock core.MerkleClock
}

// NewMerkleLWWRegister creates a new instance (or loaded from DB) of a MerkleCRDT
// backed by a LWWRegister CRDT
func NewMerkleLWWRegister(store ds.Datastore, dagstore *store.DAGStore, ns, key ds.Key, log logging.StandardLogger) *MerkleLWWRegister {
	// New Register
	reg := corecrdt.NewLWWRegister(store, ns.ChildString("data"), key.String() /* stuff like namespace and ID */)
	// New Clock
	clk := clock.NewMerkleClock(store, dagstore, ns.ChildString("heads").Child(key), reg, crdt.LWWRegDeltaExtractorFn, log)
	// newBaseMerkleCRDT(clock, register)
	base := &baseMerkleCRDT{clk, reg}
	// instatiate MerkleLWWRegister
	// return
	return &MerkleLWWRegister{
		baseMerkleCRDT: base,
		clock:          clk,
		reg:            reg,
	}
}

// Set the value of the register
func (mlww *MerkleLWWRegister) Set(value []byte) error {
	// Set() call on underlying LWWRegister CRDT
	// persist/publish delta
	delta := mlww.reg.Set(value)
	_, err := mlww.Publish(delta)
	return err
}

// Value will retrieve the current value from the db
func (mlww *MerkleLWWRegister) Value() ([]byte, error) {
	return mlww.reg.Value()
}

// Merge writes the provided delta to state using a supplied
// merge semantic
func (mlww *MerkleLWWRegister) Merge(other core.Delta, id string) error {
	return mlww.reg.Merge(other, id)
}
