package core

import (
	"errors"

	cid "github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
)

var (
	// ErrMismatchedMergeType - Tying to merge two ReplicatedData of different types
	ErrMismatchedMergeType = errors.New("Given type to merge does not match source")
)

// ReplicatedData is a data type that allows concurrent writers
// to deterministicly merge other replicated data so as to
// converge on the same state
type ReplicatedData interface {
	Merge(other Delta, id string) error
	DeltaDecode(node ipld.Node) (Delta, error) // possibly rename to just Decode
}

// PersistedReplicatedData persists a ReplicatedData to an underlying datastore
type PersistedReplicatedData interface {
	ReplicatedData
	Publish(Delta) (cid.Cid, error)
}
