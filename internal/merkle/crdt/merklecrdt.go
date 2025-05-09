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
}

type FieldLevelMerkleCRDT interface {
	MerkleCRDT
	Save(ctx context.Context, data *DocField) (cidlink.Link, []byte, error)
}

func FieldLevelCRDTWithStore(
	store Stores,
	schemaVersionID string,
	cType client.CType,
	kind client.FieldKind,
	key keys.DataStoreKey,
	fieldName string,
) (FieldLevelMerkleCRDT, error) {
	switch cType {
	case client.LWW_REGISTER:
		return NewMerkleLWWRegister(
			store,
			schemaVersionID,
			key,
			fieldName,
		), nil
	case client.PN_COUNTER, client.P_COUNTER:
		return NewMerkleCounter(
			store,
			schemaVersionID,
			key,
			fieldName,
			cType == client.PN_COUNTER,
			kind.(client.ScalarKind),
		), nil
	}
	return nil, client.NewErrUnknownCRDT(cType)
}
