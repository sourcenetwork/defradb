package crdt

import (
	"github.com/sourcenetwork/defradb/core"
	corecrdt "github.com/sourcenetwork/defradb/core/crdt"
	"github.com/sourcenetwork/defradb/merkle/clock"

	ds "github.com/ipfs/go-datastore"
)

// MerkleLWWRegister is a MerkleCRDT implementation of the LWWRegister
// using MerkleClocks
type MerkleLWWRegister struct {
	baseMerkleCRDT
	core.ReplicatedData

	clock clock.MerkleClock
	reg   corecrdt.LWWRegister
}

// NewMerkleLWWRegisterContainer
func NewMerkleLWWRegister(ns, key ds.Key) *MerkleLWWRegister {
	// New Clock
	clk := clock.NewMerkleClock( /*stuff like extractDeltaFn*/ )
	// New Register
	reg := corecrdt.NewLWWRegister( /*stuff like namespace and ID */ )
	// newBaseMerkleCRDT(clock, register)
	base := newBaseMerkleCRDT(clk, reg, store)
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
	return mlww.publish(delta)
}

// Value will retrieve the current value from the db
func (mlww *MerkleLWWRegister) Value() []byte {
	return mlww.reg.Value()
}
