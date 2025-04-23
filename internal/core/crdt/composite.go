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

// CompositeDAGDelta represents a delta-state update made of sub-MerkleCRDTs.
type CompositeDAGDelta struct {
	// This property is duplicated from field-level blocks.
	//
	// We could remove this without much hassle from the composite, however long-term
	// the ideal solution would be to remove it from the field-level commits *excluding*
	// the initial field level commit where it must exist in order to scope it to a particular
	// document.  This would require a local index in order to handle field level commit-queries.
	DocID    []byte
	Priority uint64
	// SchemaVersionID is the schema version datastore key at the time of commit.
	//
	// It can be used to identify the collection datastructure state at the time of commit.
	//
	// This property is deliberately duplicated from field-level blocks as it makes the P2P code
	// quite a lot easier - we can remove this from here at some point if we want to.
	//
	// Conversely we could remove this from the field-level commits and leave it on the composite,
	// however that would complicate commit-queries and would require us to maintain an index elsewhere.
	SchemaVersionID string
	// Status represents the status of the document. By default it is `Active`.
	// Alternatively, if can be set to `Deleted`.
	Status client.DocumentStatus
}

var _ core.Delta = (*CompositeDAGDelta)(nil)

// IPLDSchemaBytes returns the IPLD schema representation for the type.
//
// This needs to match the [CompositeDAGDelta] struct or [coreblock.mustSetSchema] will panic on init.
func (delta *CompositeDAGDelta) IPLDSchemaBytes() []byte {
	return []byte(`
	type CompositeDAGDelta struct {
		docID     		Bytes
		priority  		Int
		schemaVersionID String
		status          Int
	}`)
}

// GetPriority gets the current priority for this delta.
func (delta *CompositeDAGDelta) GetPriority() uint64 {
	return delta.Priority
}

// SetPriority will set the priority for this delta.
func (delta *CompositeDAGDelta) SetPriority(prio uint64) {
	delta.Priority = prio
}

// CompositeDAG is a CRDT structure that is used to track a collection of sub MerkleCRDTs.
type CompositeDAG struct {
	store           datastore.DSReaderWriter
	key             keys.DataStoreKey
	schemaVersionID string
}

var _ core.ReplicatedData = (*CompositeDAG)(nil)

func NewCompositeDAG(
	store datastore.DSReaderWriter,
	schemaVersionID string,
	key keys.DataStoreKey,
) CompositeDAG {
	return CompositeDAG{
		store:           store,
		key:             key,
		schemaVersionID: schemaVersionID,
	}
}

// Set returns a new composite DAG delta CRDT with the given status.
func (c CompositeDAG) NewDelta(status client.DocumentStatus) *CompositeDAGDelta {
	return &CompositeDAGDelta{
		DocID:           []byte(c.key.DocID),
		SchemaVersionID: c.schemaVersionID,
		Status:          status,
	}
}

// Merge implements ReplicatedData interface.
// It ensures that the object marker exists for the given key.
// If it doesn't, it adds it to the store.
func (c CompositeDAG) Merge(ctx context.Context, delta core.Delta) error {
	dagDelta, isDagDelta := delta.(*CompositeDAGDelta)

	if isDagDelta && dagDelta.Status.IsDeleted() {
		err := c.store.Set(ctx, c.key.ToPrimaryDataStoreKey().Bytes(), []byte{base.DeletedObjectMarker})
		if err != nil {
			return err
		}
		return c.deleteWithPrefix(ctx, c.key.WithValueFlag().WithFieldID(""))
	}

	// We cannot rely on the dagDelta.Status here as it may have been deleted locally, this is not
	// reflected in `dagDelta.Status` if sourced via P2P.  Updates synced via P2P should not undelete
	// the local representation of the document.
	versionKey := c.key.WithValueFlag().WithFieldID(keys.DATASTORE_DOC_VERSION_FIELD_ID)
	objectMarker, err := c.store.Get(ctx, c.key.ToPrimaryDataStoreKey().Bytes())
	hasObjectMarker := !errors.Is(err, corekv.ErrNotFound)
	if err != nil && hasObjectMarker {
		return err
	}

	if bytes.Equal(objectMarker, []byte{base.DeletedObjectMarker}) {
		versionKey = versionKey.WithDeletedFlag()
	}

	var schemaVersionId string
	if isDagDelta {
		// If this is a CompositeDAGDelta take the datastore schema version from there.
		// This is particularly important for P2P synced dags, as they may arrive here without having
		// been migrated yet locally.
		schemaVersionId = dagDelta.SchemaVersionID
	} else {
		schemaVersionId = c.schemaVersionID
	}

	err = c.store.Set(ctx, versionKey.Bytes(), []byte(schemaVersionId))
	if err != nil {
		return err
	}

	if !hasObjectMarker {
		// ensure object marker exists
		return c.store.Set(ctx, c.key.ToPrimaryDataStoreKey().Bytes(), []byte{base.ObjectMarker})
	}

	return nil
}

func (c CompositeDAG) deleteWithPrefix(ctx context.Context, key keys.DataStoreKey) error {
	iter, err := c.store.Iterator(ctx, corekv.IterOptions{
		Prefix: key.Bytes(),
	})
	if err != nil {
		return err
	}

	for {
		hasNext, err := iter.Next()
		if err != nil {
			return errors.Join(err, iter.Close())
		}
		if !hasNext {
			break
		}

		dsKey, err := keys.NewDataStoreKey(string(iter.Key()))
		if err != nil {
			return errors.Join(err, iter.Close())
		}

		if dsKey.InstanceType == keys.ValueKey {
			value, err := iter.Value()
			if err != nil {
				return errors.Join(err, iter.Close())
			}

			err = c.store.Set(ctx, dsKey.WithDeletedFlag().Bytes(), value)
			if err != nil {
				return errors.Join(err, iter.Close())
			}
		}

		err = c.store.Delete(ctx, dsKey.Bytes())
		if err != nil {
			return errors.Join(err, iter.Close())
		}
	}

	return iter.Close()
}
