// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package merklecrdt

import (
	"context"

	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	corecrdt "github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/merkle/clock"
)

// MerkleLWWRegister is a MerkleCRDT implementation of the LWWRegister using MerkleClocks.
type MerkleLWWRegister struct {
	clock           *clock.MerkleClock
	key             keys.DataStoreKey
	schemaVersionID string
	fieldName       string
}

var _ FieldLevelMerkleCRDT = (*MerkleLWWRegister)(nil)

// NewMerkleLWWRegister creates a new instance (or loaded from DB) of a MerkleCRDT
// backed by a LWWRegister CRDT.
func NewMerkleLWWRegister(
	store Stores,
	schemaVersionID string,
	key keys.DataStoreKey,
	fieldName string,
) *MerkleLWWRegister {
	register := corecrdt.NewLWWRegister(store.Datastore(), key)
	clk := clock.NewMerkleClock(store.Headstore(), store.Blockstore(), store.Encstore(), key.ToHeadStoreKey(),
		register)

	return &MerkleLWWRegister{
		clock:           clk,
		key:             key,
		schemaVersionID: schemaVersionID,
		fieldName:       fieldName,
	}
}

func (m *MerkleLWWRegister) Clock() *clock.MerkleClock {
	return m.clock
}

// Save the value of the register to the DAG.
func (m *MerkleLWWRegister) Save(ctx context.Context, data *DocField) (cidlink.Link, []byte, error) {
	bytes, err := data.FieldValue.Bytes()
	if err != nil {
		return cidlink.Link{}, nil, err
	}

	return m.clock.AddDelta(
		ctx,
		&corecrdt.LWWRegDelta{
			Data:            bytes,
			DocID:           []byte(m.key.DocID),
			FieldName:       m.fieldName,
			SchemaVersionID: m.schemaVersionID,
		},
	)
}
