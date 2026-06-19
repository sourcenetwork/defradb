// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package id

import (
	"bytes"
	"context"
	stderrors "errors"

	"github.com/ipfs/go-cid"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/keys"
)

func SetDocIDMapping(
	ctx context.Context,
	collectionShortID uint32,
	docShortID uint32,
	docID string,
) error {
	txn := datastore.CtxMustGetTxn(ctx)
	if err := txn.Systemstore().Set(
		ctx,
		keys.NewShortIDToDocIDKey(collectionShortID, docShortID).Bytes(),
		[]byte(docID),
	); err != nil {
		return err
	}

	if err := txn.Systemstore().Set(
		ctx,
		keys.NewNodeDocIDToShortIDKey(docID).Bytes(),
		keys.EncodeDocRef(collectionShortID, docShortID),
	); err != nil {
		return err
	}

	return nil
}

func SetDocIDAlias(ctx context.Context, collectionShortID uint32, docShortID uint32, docID string) error {
	txn := datastore.CtxMustGetTxn(ctx)
	return txn.Systemstore().Set(
		ctx,
		keys.NewNodeDocIDToShortIDKey(docID).Bytes(),
		keys.EncodeDocRef(collectionShortID, docShortID),
	)
}

func GetDocID(
	ctx context.Context,
	collectionShortID uint32,
	docShortID uint32,
) (string, bool, error) {
	txn := datastore.CtxMustGetTxn(ctx)
	return GetDocIDFromStore(ctx, txn.Systemstore(), collectionShortID, docShortID)
}

func GetDocIDFromStore(
	ctx context.Context,
	store corekv.Reader,
	collectionShortID uint32,
	docShortID uint32,
) (string, bool, error) {
	value, err := store.Get(ctx, keys.NewShortIDToDocIDKey(collectionShortID, docShortID).Bytes())
	if errors.Is(err, corekv.ErrNotFound) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return string(value), true, nil
}

func GetShortDocID(
	ctx context.Context,
	collectionShortID uint32,
	docID string,
) (uint32, bool, error) {
	txn := datastore.CtxMustGetTxn(ctx)
	return GetShortDocIDFromStore(ctx, txn.Systemstore(), collectionShortID, docID)
}

func GetShortDocIDFromStore(
	ctx context.Context,
	store corekv.Reader,
	collectionShortID uint32,
	docID string,
) (uint32, bool, error) {
	docRef, found, err := GetDocRefFromStore(ctx, store, docID)
	if err != nil || !found {
		return 0, found, err
	}
	if docRef.CollectionShortID != collectionShortID {
		return 0, false, nil
	}
	return docRef.DocShortID, true, nil
}

func GetDocRef(ctx context.Context, docID string) (keys.DocRef, bool, error) {
	txn := datastore.CtxMustGetTxn(ctx)
	return GetDocRefFromStore(ctx, txn.Systemstore(), docID)
}

func GetDocRefFromStore(ctx context.Context, store corekv.Reader, docID string) (keys.DocRef, bool, error) {
	value, err := store.Get(ctx, keys.NewNodeDocIDToShortIDKey(docID).Bytes())
	if errors.Is(err, corekv.ErrNotFound) {
		return keys.DocRef{}, false, nil
	}
	if err != nil {
		return keys.DocRef{}, false, err
	}
	docRef, err := keys.DecodeDocRef(value)
	if err != nil {
		return keys.DocRef{}, false, err
	}
	return docRef, true, nil
}

func SetBlockDocIDMapping(
	ctx context.Context,
	collectionShortID uint32,
	blockCID cid.Cid,
	docID string,
) error {
	if !blockCID.Defined() || docID == "" {
		return nil
	}

	txn := datastore.CtxMustGetTxn(ctx)
	return txn.Systemstore().Set(
		ctx,
		keys.NewBlockCIDToDocIDKey(collectionShortID, blockCID.String()).Bytes(),
		[]byte(docID),
	)
}

func GetDocIDForBlockFromStore(
	ctx context.Context,
	store corekv.Reader,
	collectionShortID uint32,
	blockCID cid.Cid,
) (string, bool, error) {
	if !blockCID.Defined() {
		return "", false, nil
	}
	if collectionShortID != 0 {
		value, err := store.Get(ctx, keys.NewBlockCIDToDocIDKey(collectionShortID, blockCID.String()).Bytes())
		if errors.Is(err, corekv.ErrNotFound) {
			return "", false, nil
		}
		if err != nil {
			return "", false, err
		}
		return string(value), true, nil
	}
	prefix := keys.NewBlockCIDToDocIDKey(0, blockCID.String()).ToString() + "/"
	iter, err := store.Iterator(ctx, corekv.IterOptions{Prefix: []byte(prefix)})
	if err != nil {
		return "", false, err
	}
	hasNext, err := iter.Next()
	if err != nil {
		return "", false, stderrors.Join(err, iter.Close())
	}
	if !hasNext {
		return "", false, iter.Close()
	}
	value, err := iter.Value()
	if err != nil {
		return "", false, stderrors.Join(err, iter.Close())
	}
	docID := string(value)
	if err := iter.Close(); err != nil {
		return "", false, err
	}
	return docID, true, nil
}

func DeleteBlockDocIDMappings(
	ctx context.Context,
	store corekv.ReaderWriter,
	collectionShortID uint32,
	docID string,
) error {
	if collectionShortID == 0 || docID == "" {
		return nil
	}

	blockPrefix := keys.NewBlockCIDToDocIDKey(0, "").ToString() + "/"
	iter, err := store.Iterator(ctx, corekv.IterOptions{Prefix: []byte(blockPrefix)})
	if err != nil {
		return err
	}

	collectionSuffix := append([]byte{'/'}, keys.EncodeCollectionShortID(collectionShortID)...)
	docIDBytes := []byte(docID)
	mappingKeys := make([][]byte, 0)
	for {
		hasNext, err := iter.Next()
		if err != nil {
			return stderrors.Join(err, iter.Close())
		}
		if !hasNext {
			break
		}

		value, err := iter.Value()
		if err != nil {
			return stderrors.Join(err, iter.Close())
		}
		if !bytes.Equal(value, docIDBytes) || !bytes.HasSuffix(iter.Key(), collectionSuffix) {
			continue
		}
		mappingKeys = append(mappingKeys, append([]byte(nil), iter.Key()...))
	}
	if err := iter.Close(); err != nil {
		return err
	}

	for _, key := range mappingKeys {
		if err := store.Delete(ctx, key); err != nil && !errors.Is(err, corekv.ErrNotFound) {
			return err
		}
	}
	return nil
}

func DeleteDocIDMappings(
	ctx context.Context,
	store corekv.ReaderWriter,
	collectionShortID uint32,
	docShortID uint32,
) error {
	if collectionShortID == 0 || docShortID == 0 {
		return nil
	}

	docID, found, err := GetDocIDFromStore(ctx, store, collectionShortID, docShortID)
	if err != nil {
		return err
	}

	if err := deleteKeyIfExists(ctx, store, keys.NewShortIDToDocIDKey(collectionShortID, docShortID).Bytes()); err != nil {
		return err
	}
	if err := DeleteNodeDocIDAliasesForShortDocID(ctx, store, collectionShortID, docShortID); err != nil {
		return err
	}
	if found {
		return DeleteBlockDocIDMappings(ctx, store, collectionShortID, docID)
	}
	return nil
}

func DeleteNodeDocIDAliasesForShortDocID(
	ctx context.Context,
	store corekv.ReaderWriter,
	collectionShortID uint32,
	docShortID uint32,
) error {
	if collectionShortID == 0 || docShortID == 0 {
		return nil
	}

	prefix := keys.NewNodeDocIDToShortIDKey("").ToString() + "/"
	iter, err := store.Iterator(ctx, corekv.IterOptions{Prefix: []byte(prefix)})
	if err != nil {
		return err
	}

	var aliasKeys [][]byte
	for {
		hasNext, err := iter.Next()
		if err != nil {
			return stderrors.Join(err, iter.Close())
		}
		if !hasNext {
			break
		}
		value, err := iter.Value()
		if err != nil {
			return stderrors.Join(err, iter.Close())
		}
		docRef, err := keys.DecodeDocRef(value)
		if err != nil {
			return stderrors.Join(err, iter.Close())
		}
		if docRef.CollectionShortID == collectionShortID && docRef.DocShortID == docShortID {
			aliasKeys = append(aliasKeys, append([]byte(nil), iter.Key()...))
		}
	}
	if err := iter.Close(); err != nil {
		return err
	}

	for _, key := range aliasKeys {
		if err := store.Delete(ctx, key); err != nil && !errors.Is(err, corekv.ErrNotFound) {
			return err
		}
	}
	return nil
}

func deleteKeyIfExists(ctx context.Context, store corekv.ReaderWriter, key []byte) error {
	if err := store.Delete(ctx, key); err != nil && !errors.Is(err, corekv.ErrNotFound) {
		return err
	}
	return nil
}
