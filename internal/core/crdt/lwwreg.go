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

	ds "github.com/ipfs/go-datastore"

	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/db/base"
	"github.com/sourcenetwork/defradb/internal/encryption"
)

// LWWRegDelta is a single delta operation for an LWWRegister
// @todo: Expand delta metadata (investigate if needed)
type LWWRegDelta struct {
	DocID     []byte
	FieldName string
	Priority  uint64
	// SchemaVersionID is the schema version datastore key at the time of commit.
	//
	// It can be used to identify the collection datastructure state at the time of commit.
	SchemaVersionID string
	Data            []byte
}

var _ core.Delta = (*LWWRegDelta)(nil)

// IPLDSchemaBytes returns the IPLD schema representation for the type.
//
// This needs to match the [LWWRegDelta] struct or [coreblock.mustSetSchema] will panic on init.
func (delta LWWRegDelta) IPLDSchemaBytes() []byte {
	return []byte(`
	type LWWRegDelta struct {
		docID     		Bytes
		fieldName 		String
		priority  		Int
		schemaVersionID String
		data            Bytes
	}`)
}

// GetPriority gets the current priority for this delta.
func (delta *LWWRegDelta) GetPriority() uint64 {
	return delta.Priority
}

// SetPriority will set the priority for this delta.
func (delta *LWWRegDelta) SetPriority(prio uint64) {
	delta.Priority = prio
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
		DocID:           []byte(reg.key.DocID),
		FieldName:       reg.fieldName,
		SchemaVersionID: reg.schemaVersionKey.SchemaVersionID,
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

	data := d.Data

	var err error
	data, err = encryption.DecryptDoc(ctx, string(d.DocID), 0, data)
	if err != nil {
		return err
	}

	return reg.setValue(ctx, data, d.GetPriority())
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
