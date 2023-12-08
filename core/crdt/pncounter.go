// Copyright 2022 Democratized Data Foundation
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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/errors"
)

var (
	// ensure types implements core interfaces
	_ core.ReplicatedData = (*PNCounterRegister)(nil)
	_ core.Delta          = (*PNCounterDelta)(nil)
)

// PNCounterDelta is a single delta operation for an PNCounterRegister
type PNCounterDelta struct {
	DocKey    []byte
	FieldName string
	Priority  uint64
	// SchemaVersionID is the schema version datastore key at the time of commit.
	//
	// It can be used to identify the collection datastructure state at the time of commit.
	SchemaVersionID string
	Data            []byte
	FieldValue      client.FieldValue `json:"-"` // should not be marshalled
}

// GetPriority gets the current priority for this delta.
func (delta *PNCounterDelta) GetPriority() uint64 {
	return delta.Priority
}

// SetPriority will set the priority for this delta.
func (delta *PNCounterDelta) SetPriority(prio uint64) {
	delta.Priority = prio
}

// Marshal encodes the delta using CBOR.
func (delta *PNCounterDelta) Marshal() ([]byte, error) {
	b, err := delta.FieldValue.Bytes()
	if err != nil {
		return nil, err
	}
	delta.Data = b

	h := &codec.CborHandle{}
	buf := bytes.NewBuffer(nil)
	enc := codec.NewEncoder(buf, h)
	err = enc.Encode(delta)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// PNCounterRegister, Last-Writer-Wins Register, is a simple CRDT type that allows set/get
// of an arbitrary data type that ensures convergence.
type PNCounterRegister struct {
	baseCRDT
}

// NewPNCounterRegister returns a new instance of the PNCounter with the given ID.
func NewPNCounterRegister(
	store datastore.DSReaderWriter,
	schemaVersionKey core.CollectionSchemaVersionKey,
	key core.DataStoreKey,
	fieldName string,
) PNCounterRegister {
	return PNCounterRegister{newBaseCRDT(store, key, schemaVersionKey, fieldName)}
}

// Value gets the current register value
func (reg PNCounterRegister) Value(ctx context.Context) ([]byte, error) {
	valueK := reg.key.WithValueFlag()
	buf, err := reg.store.Get(ctx, valueK.ToDS())
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// Set generates a new delta with the supplied value
func (reg PNCounterRegister) Set(value client.FieldValue) *PNCounterDelta {
	return &PNCounterDelta{
		DocKey:          []byte(reg.key.DocKey),
		FieldName:       reg.fieldName,
		FieldValue:      value,
		SchemaVersionID: reg.schemaVersionKey.SchemaVersionId,
	}
}

// Merge implements ReplicatedData interface.
// It merges two PNCounterRegisty by adding the values together.
func (reg PNCounterRegister) Merge(ctx context.Context, delta core.Delta) error {
	d, ok := delta.(*PNCounterDelta)
	if !ok {
		return ErrMismatchedMergeType
	}

	return reg.addValue(ctx, d.FieldValue, d.GetPriority())
}

func (reg PNCounterRegister) addValue(ctx context.Context, value client.FieldValue, priority uint64) error {
	var number int64
	switch v := value.Value().(type) {
	case int64:
		number = int64(v)
	case uint64:
		number = int64(v)
	default:
		return errors.New("invalid value type. Must be compatible with int64")
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

	b, err := cbor.Marshal(curValue + number)
	if err != nil {
		return err
	}

	err = reg.store.Put(ctx, key.ToDS(), b)
	if err != nil {
		return NewErrFailedToStoreValue(err)
	}

	return reg.setPriority(ctx, reg.key, priority)
}

func (reg PNCounterRegister) getCurrentValue(ctx context.Context, key core.DataStoreKey) (int64, error) {
	curValue, err := reg.store.Get(ctx, key.ToDS())
	if err != nil {
		if errors.Is(err, ds.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}

	return getInt64FromBytes(curValue)
}

// DeltaDecode is a typed helper to extract a PNCounterDelta from a ipld.Node
func (reg PNCounterRegister) DeltaDecode(node ipld.Node) (core.Delta, error) {
	delta := &PNCounterDelta{}
	pbNode, ok := node.(*dag.ProtoNode)
	if !ok {
		return nil, client.NewErrUnexpectedType[*dag.ProtoNode]("ipld.Node", node)
	}
	data := pbNode.Data()
	h := &codec.CborHandle{}
	dec := codec.NewDecoderBytes(data, h)
	err := dec.Decode(delta)
	if err != nil {
		return nil, err
	}

	val, err := getInt64FromBytes(delta.Data)
	if err != nil {
		return nil, err
	}
	delta.FieldValue = *(client.NewFieldValue(client.PN_COUNTER_REGISTER, val))

	return delta, nil
}

func getInt64FromBytes(b []byte) (int64, error) {
	var val any
	err := cbor.Unmarshal(b, &val)
	if err != nil {
		return 0, err
	}

	switch v := val.(type) {
	case int64:
		return int64(v), nil
	case uint64:
		return int64(v), nil
	default:
		return 0, errors.New("invalid value type. Must be int64 or uint64")
	}
}
