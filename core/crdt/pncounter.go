// Copyright 2023 Democratized Data Foundation
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

	"github.com/fxamacker/cbor/v2"
	dag "github.com/ipfs/boxo/ipld/merkledag"
	ds "github.com/ipfs/go-datastore"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/ugorji/go/codec"
	"golang.org/x/exp/constraints"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/errors"
)

var (
	// ensure types implements core interfaces
	_ core.ReplicatedData = (*PNCounter[float64])(nil)
	_ core.ReplicatedData = (*PNCounter[int64])(nil)
	_ core.Delta          = (*PNCounterDelta[float64])(nil)
	_ core.Delta          = (*PNCounterDelta[int64])(nil)
)

type Incrementable interface {
	constraints.Integer | constraints.Float
}

// PNCounterDelta is a single delta operation for an PNCounter
type PNCounterDelta[T Incrementable] struct {
	DocID     []byte
	FieldName string
	Priority  uint64
	// SchemaVersionID is the schema version datastore key at the time of commit.
	//
	// It can be used to identify the collection datastructure state at the time of commit.
	SchemaVersionID string
	Data            T
}

// GetPriority gets the current priority for this delta.
func (delta *PNCounterDelta[T]) GetPriority() uint64 {
	return delta.Priority
}

// SetPriority will set the priority for this delta.
func (delta *PNCounterDelta[T]) SetPriority(prio uint64) {
	delta.Priority = prio
}

// Marshal encodes the delta using CBOR.
func (delta *PNCounterDelta[T]) Marshal() ([]byte, error) {
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
func (delta *PNCounterDelta[T]) Unmarshal(b []byte) error {
	h := &codec.CborHandle{}
	dec := codec.NewDecoderBytes(b, h)
	return dec.Decode(delta)
}

// PNCounter, is a simple CRDT type that allows increment/decrement
// of an Int and Float data types that ensures convergence.
type PNCounter[T Incrementable] struct {
	baseCRDT
}

// NewPNCounter returns a new instance of the PNCounter with the given ID.
func NewPNCounter[T Incrementable](
	store datastore.DSReaderWriter,
	schemaVersionKey core.CollectionSchemaVersionKey,
	key core.DataStoreKey,
	fieldName string,
) PNCounter[T] {
	return PNCounter[T]{newBaseCRDT(store, key, schemaVersionKey, fieldName)}
}

// Value gets the current register value
func (reg PNCounter[T]) Value(ctx context.Context) ([]byte, error) {
	valueK := reg.key.WithValueFlag()
	buf, err := reg.store.Get(ctx, valueK.ToDS())
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// Set generates a new delta with the supplied value
func (reg PNCounter[T]) Increment(value T) *PNCounterDelta[T] {
	return &PNCounterDelta[T]{
		DocID:           []byte(reg.key.DocID),
		FieldName:       reg.fieldName,
		Data:            value,
		SchemaVersionID: reg.schemaVersionKey.SchemaVersionId,
	}
}

// Merge implements ReplicatedData interface.
// It merges two PNCounterRegisty by adding the values together.
func (reg PNCounter[T]) Merge(ctx context.Context, delta core.Delta) error {
	d, ok := delta.(*PNCounterDelta[T])
	if !ok {
		return ErrMismatchedMergeType
	}

	return reg.incrementValue(ctx, d.Data, d.GetPriority())
}

func (reg PNCounter[T]) incrementValue(ctx context.Context, value T, priority uint64) error {
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

func (reg PNCounter[T]) getCurrentValue(ctx context.Context, key core.DataStoreKey) (T, error) {
	curValue, err := reg.store.Get(ctx, key.ToDS())
	if err != nil {
		if errors.Is(err, ds.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}

	return getNumericFromBytes[T](curValue)
}

// DeltaDecode is a typed helper to extract a PNCounterDelta from a ipld.Node
func (reg PNCounter[T]) DeltaDecode(node ipld.Node) (core.Delta, error) {
	pbNode, ok := node.(*dag.ProtoNode)
	if !ok {
		return nil, client.NewErrUnexpectedType[*dag.ProtoNode]("ipld.Node", node)
	}

	delta := &PNCounterDelta[T]{}
	err := delta.Unmarshal(pbNode.Data())
	if err != nil {
		return nil, err
	}

	return delta, nil
}

func getNumericFromBytes[T Incrementable](b []byte) (T, error) {
	var val T
	err := cbor.Unmarshal(b, &val)
	if err != nil {
		return val, err
	}
	return val, nil
}
