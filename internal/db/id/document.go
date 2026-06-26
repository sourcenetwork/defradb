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
	docShortID uint64,
	docID string,
) error {
	txn := datastore.CtxMustGetTxn(ctx)
	if err := txn.Systemstore().Set(
		ctx,
		keys.NewDocShortIDToDocIDKey(docShortID).Bytes(),
		[]byte(docID),
	); err != nil {
		return err
	}

	return SetDocIDToDocRefMapping(ctx, collectionShortID, docShortID, docID)
}

func SetDocIDToDocRefMapping(ctx context.Context, collectionShortID uint32, docShortID uint64, docID string) error {
	if collectionShortID == 0 || docShortID == 0 || docID == "" {
		return nil
	}

	txn := datastore.CtxMustGetTxn(ctx)
	if err := txn.Systemstore().Set(
		ctx,
		keys.NewDocIDToDocRefKey(docID).Bytes(),
		keys.EncodeDocRef(collectionShortID, docShortID),
	); err != nil {
		return err
	}

	return txn.Systemstore().Set(ctx, keys.NewDocShortIDToDocIDAliasKey(docShortID, docID).Bytes(), []byte(docID))
}

func GetDocID(
	ctx context.Context,
	docShortID uint64,
) (string, bool, error) {
	txn := datastore.CtxMustGetTxn(ctx)
	return GetDocIDFromStore(ctx, txn.Systemstore(), docShortID)
}

func GetDocIDFromStore(
	ctx context.Context,
	store corekv.Reader,
	docShortID uint64,
) (string, bool, error) {
	value, err := store.Get(ctx, keys.NewDocShortIDToDocIDKey(docShortID).Bytes())
	if errors.Is(err, corekv.ErrNotFound) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return string(value), true, nil
}

func GetDocShortID(
	ctx context.Context,
	collectionShortID uint32,
	docID string,
) (uint64, bool, error) {
	txn := datastore.CtxMustGetTxn(ctx)
	return GetDocShortIDFromStore(ctx, txn.Systemstore(), collectionShortID, docID)
}

func GetDocShortIDFromStore(
	ctx context.Context,
	store corekv.Reader,
	collectionShortID uint32,
	docID string,
) (uint64, bool, error) {
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
	value, err := store.Get(ctx, keys.NewDocIDToDocRefKey(docID).Bytes())
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
	blockCID cid.Cid,
	docID string,
) error {
	if !blockCID.Defined() || docID == "" {
		return nil
	}

	txn := datastore.CtxMustGetTxn(ctx)
	return txn.Systemstore().Set(
		ctx,
		keys.NewBlockCIDToDocIDKey(blockCID.String(), docID).Bytes(),
		[]byte{},
	)
}

func GetDocIDsForBlockFromStore(
	ctx context.Context,
	store corekv.Reader,
	blockCID cid.Cid,
) ([]string, error) {
	if !blockCID.Defined() {
		return nil, nil
	}

	prefix := keys.NewBlockCIDToDocIDKey(blockCID.String(), "").Bytes()
	prefix = append(prefix, '/')
	iter, err := store.Iterator(ctx, corekv.IterOptions{
		Prefix:   prefix,
		KeysOnly: true,
	})
	if err != nil {
		return nil, err
	}

	var docIDs []string
	for {
		hasNext, err := iter.Next()
		if err != nil {
			return nil, stderrors.Join(err, iter.Close())
		}
		if !hasNext {
			break
		}
		docID := bytes.TrimPrefix(iter.Key(), prefix)
		if len(docID) != 0 {
			docIDs = append(docIDs, string(docID))
		}
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return docIDs, nil
}

func GetDocIDForBlockFromStore(
	ctx context.Context,
	store corekv.Reader,
	blockCID cid.Cid,
) (string, bool, error) {
	docIDs, err := GetDocIDsForBlockFromStore(ctx, store, blockCID)
	if err != nil {
		return "", false, err
	}
	if len(docIDs) == 0 {
		return "", false, nil
	}
	return docIDs[0], true, nil
}

func DeleteBlockDocIDMapping(
	ctx context.Context,
	store corekv.ReaderWriter,
	blockCID cid.Cid,
	docID string,
) error {
	if !blockCID.Defined() || docID == "" {
		return nil
	}
	return deleteKeyIfExists(ctx, store, keys.NewBlockCIDToDocIDKey(blockCID.String(), docID).Bytes())
}

func DeleteDocIDMappings(
	ctx context.Context,
	store corekv.ReaderWriter,
	docShortID uint64,
) error {
	if docShortID == 0 {
		return nil
	}

	if err := deleteKeyIfExists(ctx, store, keys.NewDocShortIDToDocIDKey(docShortID).Bytes()); err != nil {
		return err
	}
	return DeleteDocRefMappings(ctx, store, docShortID)
}

func DeleteDocRefMappings(
	ctx context.Context,
	store corekv.ReaderWriter,
	docShortID uint64,
) error {
	if docShortID == 0 {
		return nil
	}

	prefix := keys.NewDocShortIDToDocIDAliasKey(docShortID, "").ToString() + "/"
	iter, err := store.Iterator(ctx, corekv.IterOptions{Prefix: []byte(prefix)})
	if err != nil {
		return err
	}

	type docRefMapping struct {
		docID string
		key   []byte
	}
	var mappings []docRefMapping
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
		mappings = append(mappings, docRefMapping{
			docID: string(value),
			key:   append([]byte(nil), iter.Key()...),
		})
	}
	if err := iter.Close(); err != nil {
		return err
	}

	for _, mapping := range mappings {
		if mapping.docID != "" {
			if err := deleteKeyIfExists(ctx, store, keys.NewDocIDToDocRefKey(mapping.docID).Bytes()); err != nil {
				return err
			}
		}
		if err := store.Delete(ctx, mapping.key); err != nil && !errors.Is(err, corekv.ErrNotFound) {
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
