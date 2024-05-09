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
	ds "github.com/ipfs/go-datastore"
	"golang.org/x/exp/constraints"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/db/base"
)

type Incrementable interface {
	constraints.Integer | constraints.Float
}

// CounterDelta is a single delta operation for a Counter
type CounterDelta[T Incrementable] struct {
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

var _ core.Delta = (*CounterDelta[float64])(nil)
var _ core.Delta = (*CounterDelta[int64])(nil)

// GetPriority gets the current priority for this delta.
func (delta *CounterDelta[T]) GetPriority() uint64 {
	return delta.Priority
}

// SetPriority will set the priority for this delta.
func (delta *CounterDelta[T]) SetPriority(prio uint64) {
	delta.Priority = prio
}

// Counter, is a simple CRDT type that allows increment/decrement
// of an Int and Float data types that ensures convergence.
type Counter[T Incrementable] struct {
	baseCRDT
	AllowDecrement bool
}

var _ core.ReplicatedData = (*Counter[float64])(nil)
var _ core.ReplicatedData = (*Counter[int64])(nil)

// NewCounter returns a new instance of the Counter with the given ID.
func NewCounter[T Incrementable](
	store datastore.DSReaderWriter,
	schemaVersionKey core.CollectionSchemaVersionKey,
	key core.DataStoreKey,
	fieldName string,
	allowDecrement bool,
) Counter[T] {
	return Counter[T]{newBaseCRDT(store, key, schemaVersionKey, fieldName), allowDecrement}
}

// Value gets the current counter value
func (c Counter[T]) Value(ctx context.Context) ([]byte, error) {
	valueK := c.key.WithValueFlag()
	buf, err := c.store.Get(ctx, valueK.ToDS())
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// Set generates a new delta with the supplied value.
//
// WARNING: Incrementing an integer and causing it to overflow the int64 max value
// will cause the value to roll over to the int64 min value. Incremeting a float and
// causing it to overflow the float64 max value will act like a no-op.
func (c Counter[T]) Increment(ctx context.Context, value T) (*CounterDelta[T], error) {
	// To ensure that the dag block is unique, we add a random number to the delta.
	// This is done only on update (if the doc doesn't already exist) to ensure that the
	// initial dag block of a document can be reproducible.
	exists, err := c.store.Has(ctx, c.key.ToPrimaryDataStoreKey().ToDS())
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

	return &CounterDelta[T]{
		DocID:           []byte(c.key.DocID),
		FieldName:       c.fieldName,
		Data:            value,
		SchemaVersionID: c.schemaVersionKey.SchemaVersionId,
		Nonce:           nonce,
	}, nil
}

// Merge implements ReplicatedData interface.
// It merges two CounterRegisty by adding the values together.
func (c Counter[T]) Merge(ctx context.Context, delta core.Delta) error {
	d, ok := delta.(*CounterDelta[T])
	if !ok {
		return ErrMismatchedMergeType
	}

	return c.incrementValue(ctx, d.Data, d.GetPriority())
}

func (c Counter[T]) incrementValue(ctx context.Context, value T, priority uint64) error {
	if !c.AllowDecrement && value < 0 {
		return NewErrNegativeValue(value)
	}
	key := c.key.WithValueFlag()
	marker, err := c.store.Get(ctx, c.key.ToPrimaryDataStoreKey().ToDS())
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return err
	}
	if bytes.Equal(marker, []byte{base.DeletedObjectMarker}) {
		key = key.WithDeletedFlag()
	}

	curValue, err := c.getCurrentValue(ctx, key)
	if err != nil {
		return err
	}

	newValue := curValue + value
	b, err := cbor.Marshal(newValue)
	if err != nil {
		return err
	}

	err = c.store.Put(ctx, key.ToDS(), b)
	if err != nil {
		return NewErrFailedToStoreValue(err)
	}

	return c.setPriority(ctx, c.key, priority)
}

func (c Counter[T]) getCurrentValue(ctx context.Context, key core.DataStoreKey) (T, error) {
	curValue, err := c.store.Get(ctx, key.ToDS())
	if err != nil {
		if errors.Is(err, ds.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}

	return getNumericFromBytes[T](curValue)
}

func (c Counter[T]) CType() client.CType {
	if c.AllowDecrement {
		return client.PN_COUNTER
	}
	return client.P_COUNTER
}

func getNumericFromBytes[T Incrementable](b []byte) (T, error) {
	var val T
	err := cbor.Unmarshal(b, &val)
	if err != nil {
		return val, err
	}
	return val, nil
}
