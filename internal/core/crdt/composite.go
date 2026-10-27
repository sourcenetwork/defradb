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

type DocCompositeDelta struct {
	Priority uint64
	// CollectionVersionID is the collection version datastore key at the time of commit.
	//
	// It can be used to identify the collection datastructure state at the time of commit.
	//
	// This property is deliberately duplicated from field-level blocks as it makes the P2P code
	// quite a lot easier - we can remove this from here at some point if we want to.
	//
	// Conversely we could remove this from the field-level commits and leave it on the composite,
	// however that would complicate commit-queries and would require us to maintain an index elsewhere.
	CollectionVersionID string
	// Status represents the status of the document.
	Status client.DocumentStatus
}

var _ Delta = (*DocCompositeDelta)(nil)

// IPLDSchemaBytes returns the IPLD schema representation for the type.
//
// This needs to match the [DocCompositeDelta] struct or [coreblock.mustSetSchema] will panic on init.
func (delta *DocCompositeDelta) IPLDSchemaBytes() []byte {
	return []byte(`
	type DocCompositeDelta struct {
		priority  			Int
		collectionVersionID String
		status          	Int
	}`)
}

// GetPriority gets the current priority for this delta.
func (delta *DocCompositeDelta) GetPriority() uint64 {
	return delta.Priority
}

type DocComposite struct{}

var _ DocumentValueCRDT = (*DocComposite)(nil)

func NewDocComposite() *DocComposite {
	return &DocComposite{}
}

func (m *DocComposite) Delete(
	collectionVersionID string,
	priority uint64,
) *DocCompositeDelta {
	return &DocCompositeDelta{
		CollectionVersionID: collectionVersionID,
		Status:              client.Deleted,
		Priority:            priority,
	}
}

func (m *DocComposite) Upsert(
	collectionVersionID string,
	priority uint64,
) *DocCompositeDelta {
	return &DocCompositeDelta{
		CollectionVersionID: collectionVersionID,
		Status:              client.Active,
		Priority:            priority,
	}
}

// It ensures that the object marker exists for the given key.
// If it doesn't, it adds it to the store.
func (m *DocComposite) Merge(
	ctx context.Context,
	store datastore.Keyedstore,
	key keys.PrimaryDataStoreKey,
	delta Delta,
) error {
	dagDelta, ok := delta.(*DocCompositeDelta)
	if !ok {
		return ErrMismatchedMergeType
	}

	if dagDelta.Status.IsDeleted() {
		err := store.Set(ctx, key, []byte{base.DeletedObjectMarker})
		if err != nil {
			return NewErrSetDocAsDeleted(err)
		}
		return m.deleteWithPrefix(ctx, store, key.ToDataStoreKey().WithValueFlag())
	}

	versionKey := key.ToDataStoreKey().WithValueFlag().WithFieldID(keys.DATASTORE_DOC_VERSION_FIELD_ID)

	// We cannot rely on the dagDelta.Status here as it may have been deleted locally, this is not
	// reflected in `dagDelta.Status` if sourced via P2P.  Updates synced via P2P should not undelete
	// the local representation of the document.
	objectMarker, err := store.Get(ctx, key)
	hasObjectMarker := !errors.Is(err, corekv.ErrNotFound)
	if err != nil && hasObjectMarker {
		return NewErrGetDocMarker(err)
	}

	if bytes.Equal(objectMarker, []byte{base.DeletedObjectMarker}) {
		versionKey = versionKey.WithDeletedFlag()
	}

	err = store.Set(ctx, versionKey, []byte(dagDelta.CollectionVersionID))
	if err != nil {
		return NewErrSetDocVersion(err)
	}

	if !hasObjectMarker {
		// ensure object marker exists
		return store.Set(ctx, key, []byte{base.ObjectMarker})
	}

	return nil
}

func (m DocComposite) deleteWithPrefix(ctx context.Context, store datastore.Keyedstore, key keys.DataStoreKey) error {
	iter, err := store.Iterator(ctx, datastore.IterOptions{
		Prefix: key,
	})
	if err != nil {
		return NewErrCreateDeleteIter(err)
	}

	// Since some of the underlying datastores don't support mutating state in the middle of iterating, we
	// collect the affected key/values and apply the mutations afterwards.
	type kv struct {
		key   keys.DataStoreKey
		value []byte
	}
	kvArray := []kv{}
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

		value, err := iter.Value()
		if err != nil {
			return errors.Join(err, iter.Close())
		}

		kvArray = append(kvArray, kv{
			key:   dsKey,
			value: value,
		})
	}

	err = iter.Close()
	if err != nil {
		return err
	}

	for _, item := range kvArray {
		err = store.Set(ctx, item.key.WithDeletedFlag(), item.value)
		if err != nil {
			return NewErrSetDeletedFlag(err)
		}
		err = store.Delete(ctx, item.key)
		if err != nil {
			return NewErrDeleteFieldValue(err)
		}
	}

	return nil
}
