package crdt

import (
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/core/crdt"
	corecrdt "github.com/sourcenetwork/defradb/core/crdt"
	"github.com/sourcenetwork/defradb/merkle/clock"

	ds "github.com/ipfs/go-datastore"
)

// MerkleLWWRegister is a MerkleCRDT implementation of the LWWRegister
// using MerkleClocks
type MerkleLWWRegister struct {
	*baseMerkleCRDT
	// core.ReplicatedData

	reg   corecrdt.LWWRegister
	clock core.MerkleClock
}

// NewMerkleLWWRegister
func NewMerkleLWWRegister(store ds.Datastore, ns, key ds.Key) *MerkleLWWRegister {
	// New Register
	reg := corecrdt.NewLWWRegister(store, key.String() /* stuff like namespace and ID */)
	// New Clock
	clk := clock.NewMerkleClock(store, ns, reg, crdt.LWWRegDeltaExtractorFn /* + stuff like extractDeltaFn*/)
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
	_, err := mlww.Publish(&delta)
	return err
}

// Value will retrieve the current value from the db
func (mlww *MerkleLWWRegister) Value() []byte {
	return mlww.reg.Value()
}

// Merge writes the provided delta to state using a supplied
// merge semantic
func (mlww *MerkleLWWRegister) Merge(other core.Delta, id string) error {
	return mlww.reg.Merge(other, id)
}
