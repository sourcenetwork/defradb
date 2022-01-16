package core

import cid "github.com/ipfs/go-cid"

// Log represents a new DAG node added to the
// append-only MerkleCRDT Clock graph of a
// document or sub-field.
// Note: This may need to be an interface :/
type Log struct {
	DocKey string
	CID    cid.Cid
	Delta  Delta
}
