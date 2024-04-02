// Copyright 2024 Democratized Data Foundation
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
	"bytes"
	"context"
	"crypto/rand"
	"math"
	"math/big"

	"github.com/fxamacker/cbor/v2"
	dag "github.com/ipfs/boxo/ipld/merkledag"
	ds "github.com/ipfs/go-datastore"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/ugorji/go/codec"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/errors"
)

var (
	// ensure types implements core interfaces
	_ core.ReplicatedData = (*PCounter[float64])(nil)
	_ core.ReplicatedData = (*PCounter[int64])(nil)
	_ core.Delta          = (*PCounterDelta[float64])(nil)
	_ core.Delta          = (*PCounterDelta[int64])(nil)
)

// PCounterDelta is a single delta operation for an PCounter
type PCounterDelta[T Incrementable] struct {
	DocID     []byte
	FieldName string
	Priority  uint64
	// Nonce is an added randomly generated number that ensures
	// that each increment operation is unique.
	Nonce int64
	// SchemaVersionID is the schema version datastore key at the time of commit.
	//
	// It can be used to identify the collection datastructure state at the time of commit.
	SchemaVersionID string
	Data            T
}

// GetPriority gets the current priority for this delta.
func (delta *PCounterDelta[T]) GetPriority() uint64 {
	return delta.Priority
}

// SetPriority will set the priority for this delta.
func (delta *PCounterDelta[T]) SetPriority(prio uint64) {
	delta.Priority = prio
}

// Marshal encodes the delta using CBOR.
func (delta *PCounterDelta[T]) Marshal() ([]byte, error) {
	h := &codec.CborHandle{}
	buf := bytes.NewBuffer(nil)
	enc := codec.NewEncoder(buf, h)
	err := enc.Encode(delta)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Unmarshal decodes the delta from CBOR.
func (delta *PCounterDelta[T]) Unmarshal(b []byte) error {
	h := &codec.CborHandle{}
	dec := codec.NewDecoderBytes(b, h)
	return dec.Decode(delta)
}

// PCounter, is a simple CRDT type that allows increment/decrement
// of an Int and Float data types that ensures convergence.
type PCounter[T Incrementable] struct {
	baseCRDT
}

// NewPCounter returns a new instance of the PCounter with the given ID.
func NewPCounter[T Incrementable](
	store datastore.DSReaderWriter,
	schemaVersionKey core.CollectionSchemaVersionKey,
	key core.DataStoreKey,
	fieldName string,
) PCounter[T] {
	return PCounter[T]{newBaseCRDT(store, key, schemaVersionKey, fieldName)}
}

// Value gets the current register value
func (reg PCounter[T]) Value(ctx context.Context) ([]byte, error) {
	valueK := reg.key.WithValueFlag()
	buf, err := reg.store.Get(ctx, valueK.ToDS())
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// Set generates a new delta with the supplied value
func (reg PCounter[T]) Increment(ctx context.Context, value T) (*PCounterDelta[T], error) {
	// To ensure that the dag block is unique, we add a random number to the delta.
	// This is done only on update (if the doc doesn't already exist) to ensure that the
	// initial dag block of a document can be reproducible.
	exists, err := reg.store.Has(ctx, reg.key.ToPrimaryDataStoreKey().ToDS())
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

	return &PCounterDelta[T]{
		DocID:           []byte(reg.key.DocID),
		FieldName:       reg.fieldName,
		Data:            value,
		SchemaVersionID: reg.schemaVersionKey.SchemaVersionId,
		Nonce:           nonce,
	}, nil
}

// Merge implements ReplicatedData interface.
// It merges two PCounterRegisty by adding the values together.
func (reg PCounter[T]) Merge(ctx context.Context, delta core.Delta) error {
	d, ok := delta.(*PCounterDelta[T])
	if !ok {
		return ErrMismatchedMergeType
	}

	return reg.incrementValue(ctx, d.Data, d.GetPriority())
}

func (reg PCounter[T]) incrementValue(ctx context.Context, value T, priority uint64) error {
	if value < 0 {
		return NewErrNegativeValue(value)
	}
	key := reg.key.WithValueFlag()
	marker, err := reg.store.Get(ctx, reg.key.ToPrimaryDataStoreKey().ToDS())
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return err
	}
	if bytes.Equal(marker, []byte{base.DeletedObjectMarker}) {
		key = key.WithDeletedFlag()
	}

	curValue, err := reg.getCurrentValue(ctx, key)
	if err != nil {
		return err
	}

	newValue := curValue + value
	b, err := cbor.Marshal(newValue)
	if err != nil {
		return err
	}

	err = reg.store.Put(ctx, key.ToDS(), b)
	if err != nil {
		return NewErrFailedToStoreValue(err)
	}

	return reg.setPriority(ctx, reg.key, priority)
}

func (reg PCounter[T]) getCurrentValue(ctx context.Context, key core.DataStoreKey) (T, error) {
	curValue, err := reg.store.Get(ctx, key.ToDS())
	if err != nil {
		if errors.Is(err, ds.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}

	return getNumericFromBytes[T](curValue)
}

// DeltaDecode is a typed helper to extract a PCounterDelta from a ipld.Node
func (reg PCounter[T]) DeltaDecode(node ipld.Node) (core.Delta, error) {
	pbNode, ok := node.(*dag.ProtoNode)
	if !ok {
		return nil, client.NewErrUnexpectedType[*dag.ProtoNode]("ipld.Node", node)
	}

	delta := &PCounterDelta[T]{}
	err := delta.Unmarshal(pbNode.Data())
	if err != nil {
		return nil, err
	}

	return delta, nil
}
