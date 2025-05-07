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

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/db/base"
	"github.com/sourcenetwork/defradb/internal/keys"
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

// MerkleLWWRegister is a MerkleCRDT implementation of the LWWRegister using MerkleClocks.
type MerkleLWWRegister struct {
	store           datastore.DSReaderWriter
	key             keys.DataStoreKey
	schemaVersionID string
	fieldName       string
}

var _ FieldLevelMerkleCRDT = (*MerkleLWWRegister)(nil)
var _ core.ReplicatedData = (*MerkleLWWRegister)(nil)

// NewMerkleLWWRegister creates a new instance (or loaded from DB) of a MerkleCRDT
// backed by a LWWRegister CRDT.
func NewMerkleLWWRegister(
	store datastore.DSReaderWriter,
	schemaVersionID string,
	key keys.DataStoreKey,
	fieldName string,
) *MerkleLWWRegister {
	return &MerkleLWWRegister{
		key:             key,
		store:           store,
		schemaVersionID: schemaVersionID,
		fieldName:       fieldName,
	}
}

func (m *MerkleLWWRegister) HeadstorePrefix() keys.HeadstoreKey {
	return m.key.ToHeadStoreKey()
}

// Save the value of the register to the DAG.
func (m *MerkleLWWRegister) Delta(ctx context.Context, data *DocField) (core.Delta, error) {
	bytes, err := data.FieldValue.Bytes()
	if err != nil {
		return nil, err
	}

	return &LWWRegDelta{
		Data:            bytes,
		DocID:           []byte(m.key.DocID),
		FieldName:       m.fieldName,
		SchemaVersionID: m.schemaVersionID,
	}, nil
}

// Merge implements ReplicatedData interface
// Merge two LWWRegisty based on the order of the timestamp (ts),
// if they are equal, compare IDs
// MUTATE STATE
func (reg *MerkleLWWRegister) Merge(ctx context.Context, delta core.Delta) error {
	d, ok := delta.(*LWWRegDelta)
	if !ok {
		return ErrMismatchedMergeType
	}

	return reg.setValue(ctx, d.Data, d.GetPriority())
}

func (reg *MerkleLWWRegister) setValue(ctx context.Context, val []byte, priority uint64) error {
	curPrio, err := getPriority(ctx, reg.store, reg.key)
	if err != nil {
		return NewErrFailedToGetPriority(err)
	}

	// if the current priority is higher ignore put
	// else if the current value is lexicographically
	// greater than the new then ignore
	key := reg.key.WithValueFlag()
	marker, err := reg.store.Get(ctx, reg.key.ToPrimaryDataStoreKey().Bytes())
	if err != nil && !errors.Is(err, corekv.ErrNotFound) {
		return err
	}
	if bytes.Equal(marker, []byte{base.DeletedObjectMarker}) {
		key = key.WithDeletedFlag()
	}
	if priority < curPrio {
		return nil
	} else if priority == curPrio {
		curValue, err := reg.store.Get(ctx, key.Bytes())
		if err != nil {
			return err
		}

		if bytes.Compare(curValue, val) >= 0 {
			return nil
		}
	}

	if bytes.Equal(val, client.CborNil) {
		// If len(val) is 1 or less the property is nil and there is no reason for
		// the field datastore key to exist.  Ommiting the key saves space and is
		// consistent with what would be found if the user omitted the property on
		// create.
		err = reg.store.Delete(ctx, key.Bytes())
		if err != nil {
			return err
		}
	} else {
		err = reg.store.Set(ctx, key.Bytes(), val)
		if err != nil {
			return NewErrFailedToStoreValue(err)
		}
	}

	return setPriority(ctx, reg.store, reg.key, priority)
}
