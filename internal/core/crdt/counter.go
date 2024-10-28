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
type CounterDelta struct {
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
	Data            []byte
}

var _ core.Delta = (*CounterDelta)(nil)

// IPLDSchemaBytes returns the IPLD schema representation for the type.
//
// This needs to match the [CounterDelta] struct or [coreblock.mustSetSchema] will panic on init.
func (delta *CounterDelta) IPLDSchemaBytes() []byte {
	return []byte(`
	type CounterDelta struct {
		docID     		Bytes
		fieldName 		String
		priority  		Int
		nonce 			Int
		schemaVersionID String
		data            Bytes
	}`)
}

// GetPriority gets the current priority for this delta.
func (delta *CounterDelta) GetPriority() uint64 {
	return delta.Priority
}

// SetPriority will set the priority for this delta.
func (delta *CounterDelta) SetPriority(prio uint64) {
	delta.Priority = prio
}

// Counter, is a simple CRDT type that allows increment/decrement
// of an Int and Float data types that ensures convergence.
type Counter struct {
	baseCRDT
	AllowDecrement bool
	Kind           client.ScalarKind
}

var _ core.ReplicatedData = (*Counter)(nil)

// NewCounter returns a new instance of the Counter with the given ID.
func NewCounter(
	store datastore.DSReaderWriter,
	schemaVersionKey core.CollectionSchemaVersionKey,
	key core.DataStoreKey,
	fieldName string,
	allowDecrement bool,
	kind client.ScalarKind,
) Counter {
	return Counter{newBaseCRDT(store, key, schemaVersionKey, fieldName), allowDecrement, kind}
}

// Set generates a new delta with the supplied value.
//
// WARNING: Incrementing an integer and causing it to overflow the int64 max value
// will cause the value to roll over to the int64 min value. Incremeting a float and
// causing it to overflow the float64 max value will act like a no-op.
func (c Counter) Increment(ctx context.Context, value []byte) (*CounterDelta, error) {
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

	return &CounterDelta{
		DocID:           []byte(c.key.DocID),
		FieldName:       c.fieldName,
		Data:            value,
		SchemaVersionID: c.schemaVersionKey.SchemaVersionID,
		Nonce:           nonce,
	}, nil
}

// Merge implements ReplicatedData interface.
// It merges two CounterRegisty by adding the values together.
func (c Counter) Merge(ctx context.Context, delta core.Delta) error {
	d, ok := delta.(*CounterDelta)
	if !ok {
		return ErrMismatchedMergeType
	}

	return c.incrementValue(ctx, d.Data, d.GetPriority())
}

func (c Counter) incrementValue(
	ctx context.Context,
	valueAsBytes []byte,
	priority uint64,
) error {
	key := c.key.WithValueFlag()
	marker, err := c.store.Get(ctx, c.key.ToPrimaryDataStoreKey().ToDS())
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return err
	}
	if bytes.Equal(marker, []byte{base.DeletedObjectMarker}) {
		key = key.WithDeletedFlag()
	}

	var resultAsBytes []byte

	switch c.Kind {
	case client.FieldKind_NILLABLE_INT:
		resultAsBytes, err = validateAndIncrement[int64](ctx, c.store, key, valueAsBytes, c.AllowDecrement)
		if err != nil {
			return err
		}
	case client.FieldKind_NILLABLE_FLOAT:
		resultAsBytes, err = validateAndIncrement[float64](ctx, c.store, key, valueAsBytes, c.AllowDecrement)
		if err != nil {
			return err
		}
	default:
		return NewErrUnsupportedCounterType(c.Kind)
	}

	err = c.store.Put(ctx, key.ToDS(), resultAsBytes)
	if err != nil {
		return NewErrFailedToStoreValue(err)
	}

	return c.setPriority(ctx, c.key, priority)
}

func (c Counter) CType() client.CType {
	if c.AllowDecrement {
		return client.PN_COUNTER
	}
	return client.P_COUNTER
}

func validateAndIncrement[T Incrementable](
	ctx context.Context,
	store datastore.DSReaderWriter,
	key core.DataStoreKey,
	valueAsBytes []byte,
	allowDecrement bool,
) ([]byte, error) {
	value, err := getNumericFromBytes[T](valueAsBytes)
	if err != nil {
		return nil, err
	}

	if !allowDecrement && value < 0 {
		return nil, NewErrNegativeValue(value)
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
	key core.DataStoreKey,
) (T, error) {
	curValue, err := store.Get(ctx, key.ToDS())
	if err != nil {
		if errors.Is(err, ds.ErrNotFound) {
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
