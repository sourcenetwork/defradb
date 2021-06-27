// Copyright 2020 Source Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package crdt

import (
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/base"

	"github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
	logging "github.com/ipfs/go-log"
)

var (
	log = logging.Logger("defradb.merkle.crdt")
)

// MerkleCRDT is the implementation of a Merkle Clock along with a
// CRDT payload. It implements the ReplicatedData interface
// so it can be merged with any given semantics.
type MerkleCRDT interface {
	core.ReplicatedData

	Clock() core.MerkleClock
	// core.MerkleClock
	// WithStore(core.DSReaderWriter)
	// WithNS(ds.Key)
	// ProcessNode(ng core.NodeGetter, root cid.Cid, rootPrio uint64, delta core.Delta, node ipld.Node) ([]cid.Cid, error)
	// NewObject() error
}

// type MerkleCRDTInitFn func(ds.Key) MerkleCRDT
// type MerkleCRDTFactory func(store core.DSReaderWriter, namespace ds.Key) MerkleCRDTInitFn

// Type indicates MerkleCRDT type
// type Type byte

// const (
// 	//no lint
// 	none = Type(iota) // reserved none type
// 	LWW_REGISTER
// 	OBJECT
// )

var (
	// defaultMerkleCRDTs                     = make(map[Type]MerkleCRDTFactory)
	_ core.ReplicatedData = (*baseMerkleCRDT)(nil)
)

// The baseMerkleCRDT handles the merkle crdt overhead functions
// that aren't CRDT specific like the mutations and state retrieval
// functions. It handles creating and publishing the crdt DAG with
// the help of the MerkleClock
type baseMerkleCRDT struct {
	clock core.MerkleClock
	crdt  core.ReplicatedData

	// reference to schema
	// @todo: Abstract schema definitions to CORE
	// @body: Currently schema definitions are stored in db/base/descriptions
	// which is suppose to be reserved for implementation specific data.
	// However we need to have some reference of schema here in the MerkleCRDT
	// system, which is a protocol design, and shouldn't rely on implementation
	// specific utilities.
	// So we need to abstract schema work into core or something else to seperate
	// schema from implementation, so that we can reference it here in the protocol
	// sections freely without violating our design isolation.
	schema base.SchemaDescription
}

func (base *baseMerkleCRDT) Merge(other core.Delta, id string) error {
	return base.crdt.Merge(other, id)
}

func (base *baseMerkleCRDT) DeltaDecode(node ipld.Node) (core.Delta, error) {
	return base.crdt.DeltaDecode(node)
}

func (base *baseMerkleCRDT) Value() ([]byte, error) {
	return base.crdt.Value()
}

func (base *baseMerkleCRDT) Clock() core.MerkleClock {
	return base.clock
}

// func (base *baseMerkleCRDT) ProcessNode(ng core.NodeGetter, root cid.Cid, rootPrio uint64, delta core.Delta, node ipld.Node) ([]cid.Cid, error) {
// 	current := node.Cid()
// 	err := base.Merge(delta, dshelp.CidToDsKey(current).String())
// 	if err != nil {
// 		return nil, errors.Wrapf(eff, "error merging delta from %s", current)
// 	}

// 	return base.clock.ProcessNode(ng, root, rootPrio, delta, node)
// }

// Publishes the delta to state
func (base *baseMerkleCRDT) Publish(delta core.Delta) (cid.Cid, error) {
	return base.clock.AddDAGNode(delta)
	// and broadcast
}
