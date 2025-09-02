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
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/db/base"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// LWWDelta is a single delta operation for an LWWRegister
// @todo: Expand delta metadata (investigate if needed)
type LWWDelta struct {
	DocID     []byte
	FieldName string
	Priority  uint64
	// SchemaVersionID is the schema version datastore key at the time of commit.
	//
	// It can be used to identify the collection datastructure state at the time of commit.
	SchemaVersionID string
	Data            []byte
}

var _ core.Delta = (*LWWDelta)(nil)

// IPLDSchemaBytes returns the IPLD schema representation for the type.
//
// This needs to match the [LWWDelta] struct or [coreblock.mustSetSchema] will panic on init.
func (d LWWDelta) IPLDSchemaBytes() []byte {
	return []byte(`
	type LWWDelta struct {
		docID     		Bytes
		fieldName 		String
		priority  		Int
		schemaVersionID String
		data            Bytes
	}`)
}

// GetPriority gets the current priority for this delta.
func (d *LWWDelta) GetPriority() uint64 {
	return d.Priority
}

// SetPriority will set the priority for this delta.
func (d *LWWDelta) SetPriority(prio uint64) {
	d.Priority = prio
}

// LWW is a MerkleCRDT implementation of the LWW using MerkleClocks.
type LWW struct {
	store           corekv.ReaderWriter
	key             keys.DataStoreKey
	schemaVersionID string
	fieldName       string
}

var _ FieldLevelCRDT = (*LWW)(nil)
var _ core.ReplicatedData = (*LWW)(nil)

// NewLWW creates a new instance (or loaded from DB) of a MerkleCRDT
// backed by a LWWRegister CRDT.
func NewLWW(
	store corekv.ReaderWriter,
	schemaVersionID string,
	key keys.DataStoreKey,
	fieldName string,
) *LWW {
	return &LWW{
		key:             key,
		store:           store,
		schemaVersionID: schemaVersionID,
		fieldName:       fieldName,
	}
}

func (l *LWW) HeadstorePrefix() keys.HeadstoreKey {
	return l.key.ToHeadStoreKey()
}

// Save the value of the register to the DAG.
func (l *LWW) Delta(ctx context.Context, data *DocField) (core.Delta, error) {
	bytes, err := data.FieldValue.Bytes()
	if err != nil {
		return nil, err
	}

	return &LWWDelta{
		Data:            bytes,
		DocID:           []byte(l.key.DocID),
		FieldName:       l.fieldName,
		SchemaVersionID: l.schemaVersionID,
	}, nil
}

// Merge implements ReplicatedData interface
// Merge two LWWRegisty based on the order of the timestamp (ts),
// if they are equal, compare IDs
// MUTATE STATE
func (l *LWW) Merge(ctx context.Context, delta core.Delta) error {
	d, ok := delta.(*LWWDelta)
	if !ok {
		return ErrMismatchedMergeType
	}

	return l.setValue(ctx, d.Data, d.GetPriority())
}

func (l *LWW) setValue(ctx context.Context, val []byte, priority uint64) error {
	curPrio, err := getPriority(ctx, l.store, l.key)
	if err != nil {
		return NewErrFailedToGetPriority(err)
	}

	// if the current priority is higher ignore put
	// else if the current value is lexicographically
	// greater than the new then ignore
	key := l.key.WithValueFlag()
	marker, err := l.store.Get(ctx, l.key.ToPrimaryDataStoreKey().Bytes())
	if err != nil && !errors.Is(err, corekv.ErrNotFound) {
		return err
	}
	if bytes.Equal(marker, []byte{base.DeletedObjectMarker}) {
		key = key.WithDeletedFlag()
	}
	if priority < curPrio {
		return nil
	} else if priority == curPrio {
		curValue, err := l.store.Get(ctx, key.Bytes())
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
		err = l.store.Delete(ctx, key.Bytes())
		if err != nil {
			return err
		}
	} else {
		err = l.store.Set(ctx, key.Bytes(), val)
		if err != nil {
			return NewErrFailedToStoreValue(err)
		}
	}

	return setPriority(ctx, l.store, l.key, priority)
}
