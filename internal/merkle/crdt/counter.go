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
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"math"
	"math/big"

	"github.com/fxamacker/cbor/v2"
	"golang.org/x/exp/constraints"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/db/base"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// MerkleCounter is a MerkleCRDT implementation of the Counter using MerkleClocks.
type MerkleCounter struct {
	store           datastore.DSReaderWriter
	key             keys.DataStoreKey
	schemaVersionID string
	fieldName       string
	allowDecrement  bool
	kind            client.ScalarKind
}

var _ FieldLevelMerkleCRDT = (*MerkleCounter)(nil)
var _ core.ReplicatedData = (*MerkleCounter)(nil)

// NewMerkleCounter creates a new instance (or loaded from DB) of a MerkleCRDT
// backed by a Counter CRDT.
func NewMerkleCounter(
	store datastore.DSReaderWriter,
	schemaVersionID string,
	key keys.DataStoreKey,
	fieldName string,
	allowDecrement bool,
	kind client.ScalarKind,
) *MerkleCounter {
	return &MerkleCounter{
		store:           store,
		key:             key,
		schemaVersionID: schemaVersionID,
		fieldName:       fieldName,
		allowDecrement:  allowDecrement,
		kind:            kind,
	}
}

func (m *MerkleCounter) HeadstorePrefix() keys.HeadstoreKey {
	return m.key.ToHeadStoreKey()
}

// Save the value of the  Counter to the DAG.
//
// WARNING: Incrementing an integer and causing it to overflow the int64 max value
// will cause the value to roll over to the int64 min value. Incremeting a float and
// causing it to overflow the float64 max value will act like a no-op.
func (m *MerkleCounter) Delta(ctx context.Context, data *DocField) (core.Delta, error) {
	bytes, err := data.FieldValue.Bytes()
	if err != nil {
		return nil, err
	}

	// To ensure that the dag block is unique, we add a random number to the delta.
	// This is done only on update (if the doc doesn't already exist) to ensure that the
	// initial dag block of a document can be reproducible.
	exists, err := m.store.Has(ctx, m.key.ToPrimaryDataStoreKey().Bytes())
	if err != nil {
		return nil, err
	}

	var nonce int64
	if exists {
		r, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
		if err != nil {
			return nil, err
		}
		nonce = r.Int64()
	}

	return &crdt.CounterDelta{
		DocID:           []byte(m.key.DocID),
		FieldName:       m.fieldName,
		Data:            bytes,
		SchemaVersionID: m.schemaVersionID,
		Nonce:           nonce,
	}, nil
}

// Merge implements ReplicatedData interface.
// It merges two CounterRegisty by adding the values together.
func (c *MerkleCounter) Merge(ctx context.Context, delta core.Delta) error {
	d, ok := delta.(*crdt.CounterDelta)
	if !ok {
		return crdt.ErrMismatchedMergeType
	}

	return c.incrementValue(ctx, d.Data, d.GetPriority())
}

func (c *MerkleCounter) incrementValue(
	ctx context.Context,
	valueAsBytes []byte,
	priority uint64,
) error {
	key := c.key.WithValueFlag()
	marker, err := c.store.Get(ctx, c.key.ToPrimaryDataStoreKey().Bytes())
	if err != nil && !errors.Is(err, corekv.ErrNotFound) {
		return err
	}
	if bytes.Equal(marker, []byte{base.DeletedObjectMarker}) {
		key = key.WithDeletedFlag()
	}

	var resultAsBytes []byte

	switch c.kind {
	case client.FieldKind_NILLABLE_INT:
		resultAsBytes, err = validateAndIncrement[int64](ctx, c.store, key, valueAsBytes, c.allowDecrement)
		if err != nil {
			return err
		}
	case client.FieldKind_NILLABLE_FLOAT32:
		resultAsBytes, err = validateAndIncrement[float32](ctx, c.store, key, valueAsBytes, c.allowDecrement)
		if err != nil {
			return err
		}
	case client.FieldKind_NILLABLE_FLOAT64:
		resultAsBytes, err = validateAndIncrement[float64](ctx, c.store, key, valueAsBytes, c.allowDecrement)
		if err != nil {
			return err
		}
	default:
		return crdt.NewErrUnsupportedCounterType(c.kind)
	}

	err = c.store.Set(ctx, key.Bytes(), resultAsBytes)
	if err != nil {
		return crdt.NewErrFailedToStoreValue(err)
	}

	return setPriority(ctx, c.store, c.key, priority)
}

func (c *MerkleCounter) CType() client.CType {
	if c.allowDecrement {
		return client.PN_COUNTER
	}
	return client.P_COUNTER
}

type Incrementable interface {
	constraints.Integer | constraints.Float
}

func validateAndIncrement[T Incrementable](
	ctx context.Context,
	store datastore.DSReaderWriter,
	key keys.DataStoreKey,
	valueAsBytes []byte,
	allowDecrement bool,
) ([]byte, error) {
	value, err := getNumericFromBytes[T](valueAsBytes)
	if err != nil {
		return nil, err
	}

	if !allowDecrement && value < 0 {
		return nil, crdt.NewErrNegativeValue(value)
	}

	curValue, err := getCurrentValue[T](ctx, store, key)
	if err != nil {
		return nil, err
	}

	newValue := curValue + value
	return cbor.Marshal(newValue)
}

func getCurrentValue[T Incrementable](
	ctx context.Context,
	store datastore.DSReaderWriter,
	key keys.DataStoreKey,
) (T, error) {
	curValue, err := store.Get(ctx, key.Bytes())
	if err != nil {
		if errors.Is(err, corekv.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}

	return getNumericFromBytes[T](curValue)
}

func getNumericFromBytes[T Incrementable](b []byte) (T, error) {
	var val T
	err := cbor.Unmarshal(b, &val)
	if err != nil {
		return val, err
	}
	return val, nil
}

func setPriority(
	ctx context.Context,
	store datastore.DSReaderWriter,
	key keys.DataStoreKey,
	priority uint64,
) error {
	prioK := key.WithPriorityFlag()
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, priority)
	if n == 0 {
		return crdt.ErrEncodingPriority
	}

	return store.Set(ctx, prioK.Bytes(), buf[0:n])
}
