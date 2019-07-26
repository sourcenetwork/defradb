package crdt

import (
	"github.com/sourcenetwork/defradb/core"

	ds "github.com/ipfs/go-datastore"
)

type MerkleCRDT interface {
	core.ReplicatedData
	core.MerkleClock
	WithStore(ds.Datastore)
	WithNS(ds.Key)
	// NewObject() error
}

type MerkleCRDTInitFn func(ds.Key) MerkleCRDT
type MerkleCRDTFactory func(store ds.Datastore, namespace ds.Key) MerkleCRDTInitFn

type Type byte

const (
	LWW_REGISTER = Type(iota)
)

var (
	defaultMerkleCRDTs = make(map[Type]MerkleCRDTFactory)
)
