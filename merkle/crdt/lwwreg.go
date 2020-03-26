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

var (
	lwwFactoryFn = MerkleCRDTFactory(func(mstore core.MultiStore) MerkleCRDTInitFn {
		return func(key ds.Key) MerkleCRDT {
			return NewMerkleLWWRegister(mstore.Data(), mstore.Head(), mstore.Dag(), ds.NewKey(""), key, mstore.Log())
		}
	})
)

func init() {
	DefaultFactory.Register(LWW_REGISTER, &lwwFactoryFn)
}

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
func NewMerkleLWWRegister(store store.DSReaderWriter, headstore store.DSReaderWriter, dagstore *store.DAGStore, ns, dockey ds.Key, log logging.StandardLogger) *MerkleLWWRegister {
	// New Register
	reg := corecrdt.NewLWWRegister(store, ns, dockey.String() /* stuff like namespace and ID */)
	// New Clock
	clk := clock.NewMerkleClock(headstore, dagstore, dockey.String(), reg, crdt.LWWRegDeltaExtractorFn, log)
	// newBaseMerkleCRDT(clock, register)
	base := &baseMerkleCRDT{clk, reg}
	// instatiate MerkleLWWRegister
	// return
	return &MerkleLWWRegister{
		baseMerkleCRDT: base,
		// clock:          clk,
		reg: reg,
	}
}

// Set the value of the register
func (mlwwreg *MerkleLWWRegister) Set(value []byte) error {
	// Set() call on underlying LWWRegister CRDT
	// persist/publish delta
	delta := mlwwreg.reg.Set(value)
	_, err := mlwwreg.Publish(delta)
	return err
}

// Value will retrieve the current value from the db
func (mlwwreg *MerkleLWWRegister) Value() ([]byte, error) {
	return mlwwreg.reg.Value()
}

// Merge writes the provided delta to state using a supplied
// merge semantic
func (mlwwreg *MerkleLWWRegister) Merge(other core.Delta, id string) error {
	return mlwwreg.reg.Merge(other, id)
}
