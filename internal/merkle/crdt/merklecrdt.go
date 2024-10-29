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
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/merkle/clock"
)

// Stores is a trimmed down [datastore.Multistore] that declares only the sub-stores
// that should be accessed by this package and it's children.
type Stores interface {
	Datastore() datastore.DSReaderWriter
	Blockstore() datastore.Blockstore
	Encstore() datastore.Blockstore
	Headstore() datastore.DSReaderWriter
}

// MerkleCRDT is the implementation of a Merkle Clock along with a
// CRDT payload. It implements the ReplicatedData interface
// so it can be merged with any given semantics.
type MerkleCRDT interface {
	Clock() *clock.MerkleClock
	Save(ctx context.Context, data any) (cidlink.Link, []byte, error)
}

func InstanceWithStore(
	store Stores,
	schemaVersionKey keys.CollectionSchemaVersionKey,
	cType client.CType,
	kind client.FieldKind,
	key keys.DataStoreKey,
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
		), nil
	}
	return nil, client.NewErrUnknownCRDT(cType)
}
