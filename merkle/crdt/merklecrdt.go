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

	ipld "github.com/ipfs/go-ipld-format"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
)

type Stores interface {
	Datastore() datastore.DSReaderWriter
	DAGstore() datastore.DAGStore
	Headstore() datastore.DSReaderWriter
}

// MerkleCRDT is the implementation of a Merkle Clock along with a
// CRDT payload. It implements the ReplicatedData interface
// so it can be merged with any given semantics.
type MerkleCRDT interface {
	core.ReplicatedData
	Clock() core.MerkleClock
	Save(ctx context.Context, data any) (ipld.Node, uint64, error)
}

// baseMerkleCRDT handles the MerkleCRDT overhead functions that aren't CRDT specific like the mutations and state
// retrieval functions. It handles creating and publishing the CRDT DAG with the help of the MerkleClock.
type baseMerkleCRDT struct {
	clock core.MerkleClock
	crdt  core.ReplicatedData
}

var _ core.ReplicatedData = (*baseMerkleCRDT)(nil)

func (base *baseMerkleCRDT) Clock() core.MerkleClock {
	return base.clock
}

func (base *baseMerkleCRDT) Merge(ctx context.Context, other core.Delta) error {
	return base.crdt.Merge(ctx, other)
}

func (base *baseMerkleCRDT) DeltaDecode(node ipld.Node) (core.Delta, error) {
	return base.crdt.DeltaDecode(node)
}

func (base *baseMerkleCRDT) Value(ctx context.Context) ([]byte, error) {
	return base.crdt.Value(ctx)
}

func InstanceWithStore(
	store Stores,
	schemaVersionKey core.CollectionSchemaVersionKey,
	ctype client.CType,
	kind client.FieldKind,
	key core.DataStoreKey,
	fieldName string,
) (MerkleCRDT, error) {
	switch ctype {
	case client.LWW_REGISTER:
		return NewMerkleLWWRegister(
			store,
			schemaVersionKey,
			key,
			fieldName,
		), nil
	case client.PN_COUNTER:
		switch kind {
		case client.FieldKind_NILLABLE_INT:
			return NewMerklePNCounter[int64](
				store,
				schemaVersionKey,
				key,
				fieldName,
			), nil
		case client.FieldKind_NILLABLE_FLOAT:
			return NewMerklePNCounter[float64](
				store,
				schemaVersionKey,
				key,
				fieldName,
			), nil
		}
	case client.P_COUNTER:
		switch kind {
		case client.FieldKind_NILLABLE_INT:
			return NewMerklePCounter[int64](
				store,
				schemaVersionKey,
				key,
				fieldName,
			), nil
		case client.FieldKind_NILLABLE_FLOAT:
			return NewMerklePCounter[float64](
				store,
				schemaVersionKey,
				key,
				fieldName,
			), nil
		}
	case client.COMPOSITE:
		return NewMerkleCompositeDAG(
			store,
			schemaVersionKey,
			key,
			fieldName,
		), nil
	}
	return nil, client.NewErrUnknownCRDT(ctype)
}
