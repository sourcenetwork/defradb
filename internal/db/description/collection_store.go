// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package description

import (
	"context"
	"encoding/json"
	"errors"
	"sync"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/cache"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/db/lock"
	"github.com/sourcenetwork/defradb/internal/keys"
)

type collectionStore struct {
	forbiddenLock sync.RWMutex
	// Collections may be forbidden, in which case we must prevent as much interaction with
	// them as possible, regardless of what the underlying corekv store thinks.
	//
	// Collections are forbidden when the last local version is deleted from the node.  They
	// will become unforbidden if they are re-added.
	forbiddenCollectionIDs map[string]struct{}

	lockSet              *lock.LockSet
	txnFreeDatastore     corekv.Reader
	collectionRepository *cache.TxnRepository[CollectionIndex, client.CollectionVersion]
}

var _ cache.Repository[CollectionIndex, client.CollectionVersion] = (*collectionStore)(nil)

func newCollectionStore(lockSet *lock.LockSet, txnFreeDatastore corekv.Reader) *collectionStore {
	return &collectionStore{
		lockSet:                lockSet,
		txnFreeDatastore:       txnFreeDatastore,
		forbiddenCollectionIDs: map[string]struct{}{},
	}
}

func (i *collectionStore) Write(ctx context.Context, value client.CollectionVersion) error {
	txn := datastore.CtxMustGetTxn(ctx)

	err := id.SetShortCollectionID(ctx, value.CollectionID)
	if err != nil {
		return err
	}

	err = id.SetShortFieldIDs(ctx, value)
	if err != nil {
		return err
	}

	buf, err := json.Marshal(value)
	if err != nil {
		return err
	}

	key := keys.NewCollectionKey(value.VersionID)
	err = txn.Systemstore().Set(ctx, key.Bytes(), buf)
	if err != nil {
		return err
	}

	if !value.IsActive {
		nameKey := keys.NewCollectionNameKey(value.Name)
		idBytes, err := txn.Systemstore().Get(ctx, nameKey.Bytes())
		if err != nil {
			if !errors.Is(err, corekv.ErrNotFound) {
				return err
			}
		}

		if string(idBytes) == value.VersionID {
			err := txn.Systemstore().Delete(ctx, nameKey.Bytes())
			if err != nil {
				return err
			}
		}
	}

	if value.IsActive {
		nameKey := keys.NewCollectionNameKey(value.Name)
		err = txn.Systemstore().Set(ctx, nameKey.Bytes(), []byte(value.VersionID))
		if err != nil {
			return err
		}
	}

	isNew := value.CollectionID == value.VersionID
	if !isNew {
		// We don't need to index the version by collection id, if the version id is the collection id
		collectionVersionKey := keys.NewCollectionVersionKey(value.CollectionID, value.VersionID)
		err = txn.Systemstore().Set(ctx, collectionVersionKey.Bytes(), []byte{})
		if err != nil {
			return err
		}
	}

	i.forbiddenLock.Lock()
	// If this transaction writes a collection version that was previously forbidden, we must unforbid it
	// within the context of this transaction.
	delete(i.forbiddenCollectionIDs, value.CollectionID)
	i.forbiddenLock.Unlock()

	return nil
}

func (i *collectionStore) TryGet(ctx context.Context, key CollectionIndex) (client.CollectionVersion, bool, error) {
	var col client.CollectionVersion
	var err error
	var hasValue bool

	switch key.Kind {
	case CollectionVersionID:
		col, hasValue, err = i.getCollectionByVersionID(ctx, key.Value)

	case CollectionID:
		col, hasValue, err = i.getActiveCollectionByCollectionID(ctx, key.Value)

	case CollectionName:
		col, hasValue, err = i.getCollectionByName(ctx, key.Value)
	}

	if err != nil {
		return client.CollectionVersion{}, false, err
	}

	if !hasValue {
		return client.CollectionVersion{}, false, client.ErrCollectionNotFound
	}

	i.forbiddenLock.RLock()
	if _, ok := i.forbiddenCollectionIDs[col.CollectionID]; ok {
		i.forbiddenLock.RUnlock()
		return client.CollectionVersion{}, false, client.ErrCollectionNotFound
	}
	i.forbiddenLock.RUnlock()

	return col, true, nil
}

func (i *collectionStore) getCollectionByVersionID(
	ctx context.Context,
	versionID string,
) (client.CollectionVersion, bool, error) {
	txn := datastore.CtxMustGetTxn(ctx)

	key := keys.NewCollectionKey(versionID)
	buf, err := txn.Systemstore().Get(ctx, key.Bytes())
	if err != nil {
		if errors.Is(err, corekv.ErrNotFound) {
			return client.CollectionVersion{}, false, nil
		}
		return client.CollectionVersion{}, false, err
	}

	var col client.CollectionVersion
	err = json.Unmarshal(buf, &col)
	if err != nil {
		return client.CollectionVersion{}, false, err
	}

	return col, true, nil
}

// GetCollectionByName returns the collection with the given name.
//
// If no collection of that name is found, it will return an error.
func (i *collectionStore) getCollectionByName(
	ctx context.Context,
	name string,
) (client.CollectionVersion, bool, error) {
	txn := datastore.CtxMustGetTxn(ctx)

	nameKey := keys.NewCollectionNameKey(name)
	idBuf, err := txn.Systemstore().Get(ctx, nameKey.Bytes())
	if err != nil {
		if errors.Is(err, corekv.ErrNotFound) {
			return client.CollectionVersion{}, false, nil
		}
		return client.CollectionVersion{}, false, err
	}

	col, ok, err := i.collectionRepository.TryGet(ctx, CollectionIndex{
		Kind:  CollectionVersionID,
		Value: string(idBuf),
	})

	return col, ok, err
}

func (i *collectionStore) getActiveCollectionByCollectionID(
	ctx context.Context,
	collectionID string,
) (client.CollectionVersion, bool, error) {
	// The first collection version is not indexed by CollectionVersionKey, so try get it directly
	col, ok, err := i.collectionRepository.TryGet(ctx, CollectionIndex{
		Kind:  CollectionVersionID,
		Value: collectionID,
	})
	if err != nil && !errors.Is(err, client.ErrCollectionNotFound) {
		return client.CollectionVersion{}, false, err
	}
	if ok && col.IsActive {
		return col, true, nil
	}

	txn := datastore.CtxMustGetTxn(ctx)
	iter, err := txn.Systemstore().Iterator(ctx, corekv.IterOptions{
		Prefix:   keys.NewCollectionVersionKey(collectionID, "").Bytes(),
		KeysOnly: true,
	})
	if err != nil {
		return client.CollectionVersion{}, false, err
	}

	for {
		hasValue, err := iter.Next()
		if err != nil {
			return client.CollectionVersion{}, false, errors.Join(err, iter.Close())
		}

		if !hasValue {
			break
		}

		key, err := keys.NewCollectionVersionKeyFromString(string(iter.Key()))
		if err != nil {
			return client.CollectionVersion{}, false, errors.Join(err, iter.Close())
		}

		col, ok, err := i.collectionRepository.TryGet(ctx, CollectionIndex{
			Kind:  CollectionVersionID,
			Value: key.VersionID,
		})

		if err != nil {
			if errors.Is(err, client.ErrCollectionNotFound) {
				continue
			}
			return client.CollectionVersion{}, false, errors.Join(err, iter.Close())
		}

		if !ok {
			continue
		}

		if col.IsActive {
			return col, true, iter.Close()
		}
	}

	return client.CollectionVersion{}, false, iter.Close()
}

func (i *collectionStore) Delete(ctx context.Context, key CollectionIndex) error {
	version, ok, err := i.collectionRepository.TryGet(ctx, key)
	if err != nil {
		if errors.Is(err, client.ErrCollectionNotFound) {
			return nil
		}
		return err
	}
	if !ok {
		// If the collection does not exist, we don't need to delete it
		return nil
	}

	txn := datastore.CtxMustGetTxn(ctx)
	shortID, err := id.GetShortCollectionID(ctx, version.CollectionID)
	if err != nil {
		return err
	}

	versions, err := i.getCollectionsByCollectionID(ctx, version.CollectionID)
	if err != nil {
		return err
	}

	collectionKey := keys.NewCollectionKey(version.VersionID)
	err = txn.Systemstore().Delete(ctx, collectionKey.Bytes())
	if err != nil {
		return err
	}

	if version.IsActive {
		nameKey := keys.NewCollectionNameKey(version.Name)
		err = txn.Systemstore().Delete(ctx, nameKey.Bytes())
		if err != nil {
			return err
		}
	}

	isNew := version.CollectionID == version.VersionID
	if !isNew {
		collectionVersionKey := keys.NewCollectionVersionKey(version.CollectionID, version.VersionID)
		err = txn.Systemstore().Delete(ctx, collectionVersionKey.Bytes())
		if err != nil {
			return err
		}
	}

	// WARNING - DeleteShortFieldIDs is dependent on the collection short id still existing, it should be called
	// before deleting the collection short id.
	err = id.DeleteShortFieldIDs(ctx, i.lockSet, version, versions)
	if err != nil {
		return err
	}

	if len(versions) == 1 {
		// It is impossible to recreate the collection short ID once it is deleted, so we must lock the collection
		// whilst we finalize this operation, otherwise other threads/operations may try and make use of it.
		i.lockSet.CollectionLock(txn, shortID)

		hasDocs, err := i.collectionHasDocuments(ctx, version)
		if err != nil {
			return err
		}
		if hasDocs {
			// If the collection contains any documents, we do not allow deletion of any version in the
			// collection - they must first delete the documents locally, and then delete the collection.
			//
			// This is thought to be much safer than allowing document deletion along with the collection.
			//
			// This check *must* be performed after the write lock on the collection has been aquired otherwise
			// there will be a race condition.
			return NewErrCannotDeleteCollectionWithDocs(version.Name, version.VersionID)
		}

		// Only delete the collection short ID if this was the last local version
		err = id.DeleteShortCollectionID(ctx, version.CollectionID)
		if err != nil {
			return err
		}

		txn.OnSuccess(
			func() {
				// If the last local version of a collection is deleted, we must immediately prevent its
				// usage by other transactions.  This deliberately violates their transaction-isolation.
				i.collectionRepository.Forbid(version)
			},
		)
	}

	return nil
}

func (i *collectionStore) getCollectionsByCollectionID(
	ctx context.Context,
	collectionID string,
) ([]client.CollectionVersion, error) {
	txn := datastore.CtxMustGetTxn(ctx)
	cols := []client.CollectionVersion{}

	// The first collection version is not indexed by CollectionVersionKey, so try get it directly
	col, ok, err := i.collectionRepository.TryGet(ctx, CollectionIndex{
		Kind:  CollectionVersionID,
		Value: collectionID,
	})
	if err != nil && !errors.Is(err, client.ErrCollectionNotFound) {
		return nil, err
	}
	if ok {
		cols = append(cols, col)
	}

	iter, err := txn.Systemstore().Iterator(ctx, corekv.IterOptions{
		Prefix:   keys.NewCollectionVersionKey(collectionID, "").Bytes(),
		KeysOnly: true,
	})
	if err != nil {
		return nil, err
	}

	for {
		hasValue, err := iter.Next()
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}

		if !hasValue {
			break
		}

		key, err := keys.NewCollectionVersionKeyFromString(string(iter.Key()))
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}

		versionCol, ok, err := i.getCollectionByVersionID(ctx, key.VersionID)
		if err != nil {
			if errors.Is(err, client.ErrCollectionNotFound) {
				continue
			}
			return nil, errors.Join(err, iter.Close())
		}

		if ok {
			cols = append(cols, versionCol)
		}
	}

	return cols, iter.Close()
}

func (i *collectionStore) Forbid(value client.CollectionVersion) {
	i.forbiddenLock.Lock()
	i.forbiddenCollectionIDs[value.CollectionID] = struct{}{}
	i.forbiddenLock.Unlock()
}

// collectionHasDocuments checks that both the transaction, and the underlying transaction-free
// root does not have any documents in the given collection.
//
// It is nessecary to check without a transaction, as the deletion of the last collection version
// in a collection locally must be applied immediately to all existing transactions, and must also
// take into consideration writes that may have been made since the *deleting* transaction was created.
//
// The deleting transaction must be checked, as this transaction may have written documents that have
// not yet been committed and will not show up when iterating through the underlying transaction-free
// store.
func (i *collectionStore) collectionHasDocuments(
	ctx context.Context,
	version client.CollectionVersion,
) (bool, error) {
	hasDocs, err := collectionHasDocumentsTxn(ctx, version)
	if hasDocs || err != nil {
		return hasDocs, err
	}

	return i.collectionHasDocumentsRoot(ctx, version)
}

func collectionHasDocumentsTxn(
	ctx context.Context,
	version client.CollectionVersion,
) (bool, error) {
	type unsafestore interface {
		Unsafe() corekv.ReaderWriter
	}
	txn := datastore.CtxMustGetTxn(ctx)

	// We use the unsafe store here as it is convenient and allows us to re-use the same reading
	// code in a location where we do not require the lock system, as a write lock is guarenteed
	// to already be held. The Keyed store iterator function has a slightly different signature.
	return collectionHasDocumentsReader(ctx, txn.Datastore().(unsafestore).Unsafe(), version) //nolint:forcetypeassert
}

func (i *collectionStore) collectionHasDocumentsRoot(
	ctx context.Context,
	version client.CollectionVersion,
) (bool, error) {
	// corekv will pick up the previously used transaction and force its use if we do not first
	// remove it from the context.
	// This line can be removed as part of https://github.com/sourcenetwork/defradb/issues/4658
	ctx = corekv.SetCtxTxn(ctx, nil)
	return collectionHasDocumentsReader(ctx, i.txnFreeDatastore, version)
}

func collectionHasDocumentsReader(
	ctx context.Context,
	reader corekv.Reader,
	version client.CollectionVersion,
) (bool, error) {
	if !version.IsMaterialized {
		// Assume that if the collection *was* materialized, and is no longer materialized, that the cached
		// state was properly disposed of (it should have been).
		return false, nil
	}

	shortID, err := id.GetShortCollectionID(ctx, version.CollectionID)
	if err != nil {
		return false, err
	}

	var prefixKey keys.Key
	if version.Query.HasValue() {
		prefixKey = keys.NewViewCacheColPrefix(shortID)
	} else {
		prefixKey = keys.PrimaryDataStoreKey{
			CollectionShortID: shortID,
		}
	}

	iter, err := reader.Iterator(ctx, corekv.IterOptions{
		Prefix:   prefixKey.ToDS().Bytes(),
		KeysOnly: true,
	})
	if err != nil {
		return false, errors.Join(err, iter.Close())
	}

	hasValue, err := iter.Next()
	if err != nil {
		return false, errors.Join(err, iter.Close())
	}

	return hasValue, iter.Close()
}
