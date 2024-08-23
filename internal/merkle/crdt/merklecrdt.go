// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

/*
Package crdt provides CRDT implementations leveraging MerkleClock.
*/
package merklecrdt

import (
	"context"

	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/internal/core"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
)

type Stores interface {
	Datastore() datastore.DSReaderWriter
	Blockstore() datastore.Blockstore
	Headstore() datastore.DSReaderWriter
}

// MerkleCRDT is the implementation of a Merkle Clock along with a
// CRDT payload. It implements the ReplicatedData interface
// so it can be merged with any given semantics.
type MerkleCRDT interface {
	core.ReplicatedData
	Clock() MerkleClock
	Save(ctx context.Context, data any) (cidlink.Link, []byte, error)
}

// MerkleClock is the logical clock implementation that manages writing to and from
// the MerkleDAG structure, ensuring a causal ordering of events.
type MerkleClock interface {
	AddDelta(
		ctx context.Context,
		delta core.Delta,
		links ...coreblock.DAGLink,
	) (cidlink.Link, []byte, error)
	// ProcessBlock processes a block and updates the CRDT state.
	// The bool argument indicates whether only heads need to be updated. It is needed in case
	// merge should be skipped for example if the block is encrypted.
	ProcessBlock(context.Context, *coreblock.Block, cidlink.Link, bool) error
}

// baseMerkleCRDT handles the MerkleCRDT overhead functions that aren't CRDT specific like the mutations and state
// retrieval functions. It handles creating and publishing the CRDT DAG with the help of the MerkleClock.
type baseMerkleCRDT struct {
	clock MerkleClock
	crdt  core.ReplicatedData
}

var _ core.ReplicatedData = (*baseMerkleCRDT)(nil)

func (base *baseMerkleCRDT) Clock() MerkleClock {
	return base.clock
}

func (base *baseMerkleCRDT) Merge(ctx context.Context, other core.Delta) error {
	return base.crdt.Merge(ctx, other)
}

func (base *baseMerkleCRDT) Value(ctx context.Context) ([]byte, error) {
	return base.crdt.Value(ctx)
}

func InstanceWithStore(
	store Stores,
	schemaVersionKey core.CollectionSchemaVersionKey,
	cType client.CType,
	kind client.FieldKind,
	key core.DataStoreKey,
	fieldName string,
) (MerkleCRDT, error) {
	switch cType {
	case client.LWW_REGISTER:
		return NewMerkleLWWRegister(
			store,
			schemaVersionKey,
			key,
			fieldName,
		), nil
	case client.PN_COUNTER, client.P_COUNTER:
		return NewMerkleCounter(
			store,
			schemaVersionKey,
			key,
			fieldName,
			cType == client.PN_COUNTER,
			kind.(client.ScalarKind),
		), nil
	case client.COMPOSITE:
		return NewMerkleCompositeDAG(
			store,
			schemaVersionKey,
			key,
			fieldName,
		), nil
	}
	return nil, client.NewErrUnknownCRDT(cType)
}
