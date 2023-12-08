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

// LWWRegDelta is a single delta operation for an LWWRegister
// @todo: Expand delta metadata (investigate if needed)
type LWWRegDelta struct {
	SchemaVersionID string
	Priority        uint64
	Data            []byte
	DocKey          []byte
	FieldName       string
}

var _ core.Delta = (*LWWRegDelta)(nil)

// GetPriority gets the current priority for this delta.
func (delta *LWWRegDelta) GetPriority() uint64 {
	return delta.Priority
}

// SetPriority will set the priority for this delta.
func (delta *LWWRegDelta) SetPriority(prio uint64) {
	delta.Priority = prio
}

// Marshal encodes the delta using CBOR.
// for now le'ts do cbor (quick to implement)
func (delta *LWWRegDelta) Marshal() ([]byte, error) {
	h := &codec.CborHandle{}
	buf := bytes.NewBuffer(nil)
	enc := codec.NewEncoder(buf, h)
	err := enc.Encode(struct {
		SchemaVersionID string
		Priority        uint64
		Data            []byte
		DocKey          []byte
		FieldName       string
	}{delta.SchemaVersionID, delta.Priority, delta.Data, delta.DocKey, delta.FieldName})
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (delta *LWWRegDelta) Value() any {
	return delta.Data
}

// LWWRegister, Last-Writer-Wins Register, is a simple CRDT type that allows set/get
// of an arbitrary data type that ensures convergence.
type LWWRegister struct {
	baseCRDT
}

var _ core.ReplicatedData = (*LWWRegister)(nil)

// NewLWWRegister returns a new instance of the LWWReg with the given ID.
func NewLWWRegister(
	store datastore.DSReaderWriter,
	schemaVersionKey core.CollectionSchemaVersionKey,
	key core.DataStoreKey,
	fieldName string,
) LWWRegister {
	return LWWRegister{newBaseCRDT(store, key, schemaVersionKey, fieldName)}
}

// Value gets the current register value
// RETURN STATE
func (reg LWWRegister) Value(ctx context.Context) ([]byte, error) {
	valueK := reg.key.WithValueFlag()
	buf, err := reg.store.Get(ctx, valueK.ToDS())
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// Set generates a new delta with the supplied value
// RETURN DELTA
func (reg LWWRegister) Set(value []byte) *LWWRegDelta {
	return &LWWRegDelta{
		Data:            value,
		DocKey:          []byte(reg.key.DocKey),
		FieldName:       reg.fieldName,
		SchemaVersionID: reg.schemaVersionKey.SchemaVersionId,
	}
}

// Merge implements ReplicatedData interface
// Merge two LWWRegisty based on the order of the timestamp (ts),
// if they are equal, compare IDs
// MUTATE STATE
func (reg LWWRegister) Merge(ctx context.Context, delta core.Delta) error {
	d, ok := delta.(*LWWRegDelta)
	if !ok {
		return ErrMismatchedMergeType
	}

	return reg.setValue(ctx, d.Data, d.GetPriority())
}

func (reg LWWRegister) setValue(ctx context.Context, val []byte, priority uint64) error {
	curPrio, err := reg.getPriority(ctx, reg.key)
	if err != nil {
		return NewErrFailedToGetPriority(err)
	}

	// if the current priority is higher ignore put
	// else if the current value is lexicographically
	// greater than the new then ignore
	key := reg.key.WithValueFlag()
	marker, err := reg.store.Get(ctx, reg.key.ToPrimaryDataStoreKey().ToDS())
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return err
	}
	if bytes.Equal(marker, []byte{base.DeletedObjectMarker}) {
		key = key.WithDeletedFlag()
	}
	if priority < curPrio {
		return nil
	} else if priority == curPrio {
		curValue, err := reg.store.Get(ctx, key.ToDS())
		if err != nil {
			return err
		}

		if bytes.Compare(curValue, val) >= 0 {
			return nil
		}
	}

	err = reg.store.Put(ctx, key.ToDS(), val)
	if err != nil {
		return NewErrFailedToStoreValue(err)
	}

	return reg.setPriority(ctx, reg.key, priority)
}

// DeltaDecode is a typed helper to extract
// a LWWRegDelta from a ipld.Node
// for now let's do cbor (quick to implement)
func (reg LWWRegister) DeltaDecode(node ipld.Node) (core.Delta, error) {
	delta := &LWWRegDelta{}
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
	return delta, nil
}
