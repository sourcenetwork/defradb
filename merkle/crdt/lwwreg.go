package crdt

import (
	corecrdt "github.com/sourcenetwork/defradb/core/crdt"

	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	dshelp "github.com/ipfs/go-ipfs-ds-help"
)

// MerkleLWWRegister is a MerkleCRDT implementation of the LWWRegister
// using MerkleClocks
type MerkleLWWRegister struct {
	core.ReplicatedData

	mclock clock.MerkleClock
	reg    corecrdt.LWWRegister
}

// NewMerkleLWWRegisterContainer
func NewMerkleLWWRegisterContainer(key ds.Key) *MerkleLWWRegister {
	// todo
	return nil
}

func MerkleLWWRegisterFromKey(key ds.Key) *MerkleLWWRegister {
	//todo
	return nil
}

func (mlww *MerkleLWWRegister) Set(value []byte) error {
	// Set() call on underlying LWWRegister CRDT
	// persist/publish delta
	delta := mlww.reg.Set(value)
	return mlww.clock.persist(delta)
}

func (mlww *MerkleLWWRegister) Value() []byte {
	return mlww.Value()
}

func (mlww *MerkleLWWRegister) Merge(other core.ReplicatedData) error {
	return mlww.Merge(other)
}

func (mlww *MerkleLWWRegister) ProcessNode(ng *crdtNodeGetter, root cid.Cid, rootPrio uint64, delta core.Delta, node ipld.Node) ([]cid.Cid, error) {
	current := node.Cid()
	err := mlww.Merge(delta, dshelp.CidToDsKey(current).String())
	if err != nil {
		return nil, errors.Wrapf(eff, "error merging delta from %s", current)
	}
	
	return mlww.clock.ProcessNode(ng *crdtNodeGetter, root cid.Cid, rootPrio uint64, delta core.Delta, node ipld.Node)
}
