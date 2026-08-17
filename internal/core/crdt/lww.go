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
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/base"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// LWWDelta is a single delta operation for an LWWRegister
// @todo: Expand delta metadata (investigate if needed)
type LWWDelta struct {
	FieldName string
	Priority  uint64
	// CollectionVersionID is the collection version datastore key at the time of commit.
	//
	// It can be used to identify the collection datastructure state at the time of commit.
	CollectionVersionID string
	Data                []byte
}

var _ Delta = (*LWWDelta)(nil)

// IPLDSchemaBytes returns the IPLD schema representation for the type.
//
// This needs to match the [LWWDelta] struct or [coreblock.mustSetSchema] will panic on init.
func (d LWWDelta) IPLDSchemaBytes() []byte {
	return []byte(`
	type LWWDelta struct {
		fieldName 			String
		priority  			Int
		collectionVersionID String
		data            	Bytes
	}`)
}

// GetPriority gets the current priority for this delta.
func (d *LWWDelta) GetPriority() uint64 {
	return d.Priority
}

// LWW is a MerkleCRDT implementation of the LWW using MerkleClocks.
type LWW struct{}

var _ FieldValueCRDT = (*LWW)(nil)

// NewLWW creates a new instance (or loaded from DB) of a MerkleCRDT
// backed by a LWWRegister CRDT.
func NewLWW() *LWW {
	return &LWW{}
}

// Save the value of the register to the DAG.
func (l *LWW) Delta(
	ctx context.Context,
	collectionVersionID string,
	data *DocField,
	isAdd bool,
	priority uint64,
) (Delta, error) {
	bytes, err := data.FieldValue.Bytes()
	if err != nil {
		return nil, NewErrSerializeLWWValue(err, data.FieldName)
	}

	return &LWWDelta{
		Data:                bytes,
		FieldName:           data.FieldName,
		CollectionVersionID: collectionVersionID,
		Priority:            priority,
	}, nil
}

// Merge implements ReplicatedData interface
// Merge two LWWRegisty based on the order of the timestamp (ts),
// if they are equal, compare IDs
// MUTATE STATE
func (l *LWW) Merge(
	ctx context.Context,
	store datastore.Keyedstore,
	key keys.DataStoreKey,
	delta Delta,
) error {
	d, ok := delta.(*LWWDelta)
	if !ok {
		return ErrMismatchedMergeType
	}

	return l.setValue(ctx, store, key, d)
}

func (l *LWW) setValue(
	ctx context.Context,
	store datastore.Keyedstore,
	key keys.DataStoreKey,
	delta *LWWDelta,
) error {
	valueKey := key.WithValueFlag()

	curPrio, err := getPriority(ctx, store, key)
	if err != nil {
		return NewErrFailedToGetPriority(err)
	}

	marker, err := store.Get(ctx, key.ToPrimaryDataStoreKey())
	if err != nil && !errors.Is(err, corekv.ErrNotFound) {
		return NewErrGetRegisterStatus(err, delta.FieldName)
	}
	if bytes.Equal(marker, []byte{base.DeletedObjectMarker}) {
		valueKey = valueKey.WithDeletedFlag()
	}
	if delta.Priority < curPrio {
		return nil
	} else if delta.Priority == curPrio {
		curValue, err := store.Get(ctx, valueKey)
		if err != nil {
			return NewErrGetRegisterValue(err, delta.FieldName)
		}

		if bytes.Compare(curValue, delta.Data) >= 0 {
			return nil
		}
	}

	if bytes.Equal(delta.Data, client.CborNil) {
		// If len(val) is 1 or less the property is nil and there is no reason for
		// the field datastore key to exist.  Ommiting the key saves space and is
		// consistent with what would be found if the user omitted the property on
		// create.
		if err := store.Delete(ctx, valueKey); err != nil {
			return NewErrDeleteRegisterValue(err, delta.FieldName)
		}
	} else {
		if err := store.Set(ctx, valueKey, delta.Data); err != nil {
			return NewErrFailedToStoreValue(err)
		}
	}

	return setPriority(ctx, store, key, delta.Priority)
}
