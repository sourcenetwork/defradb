package core

import (
	"errors"
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
}

// PersistedReplicatedData persists a ReplicatedData to an underlying datastore
type PersistedReplicatedData interface {
	ReplicatedData
	Publish(Delta) (cid, error)
}
