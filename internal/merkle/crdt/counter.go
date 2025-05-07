// Copyright 2024 Democratized Data Foundation
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
	"crypto/rand"
	"math"
	"math/big"

	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/merkle/clock"
)

// MerkleCounter is a MerkleCRDT implementation of the Counter using MerkleClocks.
type MerkleCounter struct {
	clock           *clock.MerkleClock
	store           datastore.DSReaderWriter
	key             keys.DataStoreKey
	schemaVersionID string
	fieldName       string
}

var _ FieldLevelMerkleCRDT = (*MerkleCounter)(nil)

// NewMerkleCounter creates a new instance (or loaded from DB) of a MerkleCRDT
// backed by a Counter CRDT.
func NewMerkleCounter(
	store Stores,
	schemaVersionID string,
	key keys.DataStoreKey,
	fieldName string,
	allowDecrement bool,
	kind client.ScalarKind,
) *MerkleCounter {
	register := crdt.NewCounter(store.Datastore(), key, allowDecrement, kind)
	clk := clock.NewMerkleClock(store.Headstore(), store.Blockstore(), store.Encstore(), key.ToHeadStoreKey(),
		register)

	return &MerkleCounter{
		clock:           clk,
		store:           store.Datastore(),
		key:             key,
		schemaVersionID: schemaVersionID,
		fieldName:       fieldName,
	}
}

func (m *MerkleCounter) Clock() *clock.MerkleClock {
	return m.clock
}

// Save the value of the  Counter to the DAG.
//
// WARNING: Incrementing an integer and causing it to overflow the int64 max value
// will cause the value to roll over to the int64 min value. Incremeting a float and
// causing it to overflow the float64 max value will act like a no-op.
func (m *MerkleCounter) Save(ctx context.Context, data *DocField) (cidlink.Link, []byte, error) {
	bytes, err := data.FieldValue.Bytes()
	if err != nil {
		return cidlink.Link{}, nil, err
	}

	// To ensure that the dag block is unique, we add a random number to the delta.
	// This is done only on update (if the doc doesn't already exist) to ensure that the
	// initial dag block of a document can be reproducible.
	exists, err := m.store.Has(ctx, m.key.ToPrimaryDataStoreKey().Bytes())
	if err != nil {
		return cidlink.Link{}, nil, err
	}

	var nonce int64
	if exists {
		r, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
		if err != nil {
			return cidlink.Link{}, nil, err
		}
		nonce = r.Int64()
	}

	return m.clock.AddDelta(
		ctx,
		&crdt.CounterDelta{
			DocID:           []byte(m.key.DocID),
			FieldName:       m.fieldName,
			Data:            bytes,
			SchemaVersionID: m.schemaVersionID,
			Nonce:           nonce,
		},
	)
}
