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
	"strings"

	"github.com/ipfs/go-cid"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/keys"
)

func SetDocIDMapping(
	ctx context.Context,
	collectionShortID uint32,
	shortDocID uint32,
	docID string,
) error {
	txn := datastore.CtxMustGetTxn(ctx)
	if err := txn.Systemstore().Set(
		ctx,
		keys.NewShortIDToDocIDKey(collectionShortID, shortDocID).Bytes(),
		[]byte(docID),
	); err != nil {
		return err
	}

	if err := txn.Systemstore().Set(
		ctx,
		keys.NewNodeDocIDToShortIDKey(docID).Bytes(),
		keys.EncodeLocalDocID(collectionShortID, shortDocID),
	); err != nil {
		return err
	}

	return nil
}

func SetDocIDAlias(ctx context.Context, collectionShortID uint32, shortDocID uint32, docID string) error {
	txn := datastore.CtxMustGetTxn(ctx)
	return txn.Systemstore().Set(
		ctx,
		keys.NewNodeDocIDToShortIDKey(docID).Bytes(),
		keys.EncodeLocalDocID(collectionShortID, shortDocID),
	)
}

func GetDocID(
	ctx context.Context,
	collectionShortID uint32,
	shortDocID uint32,
) (string, bool, error) {
	txn := datastore.CtxMustGetTxn(ctx)
	return GetDocIDFromStore(ctx, txn.Systemstore(), collectionShortID, shortDocID)
}

func GetDocIDFromStore(
	ctx context.Context,
	store corekv.Reader,
	collectionShortID uint32,
	shortDocID uint32,
) (string, bool, error) {
	value, err := store.Get(ctx, keys.NewShortIDToDocIDKey(collectionShortID, shortDocID).Bytes())
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
	localDocID, found, err := GetLocalDocIDFromStore(ctx, store, docID)
	if err != nil || !found {
		return 0, found, err
	}
	if localDocID.CollectionShortID != collectionShortID {
		return 0, false, nil
	}
	return localDocID.DocShortID, true, nil
}

func GetLocalDocID(ctx context.Context, docID string) (keys.LocalDocID, bool, error) {
	txn := datastore.CtxMustGetTxn(ctx)
	return GetLocalDocIDFromStore(ctx, txn.Systemstore(), docID)
}

func GetLocalDocIDFromStore(ctx context.Context, store corekv.Reader, docID string) (keys.LocalDocID, bool, error) {
	value, err := store.Get(ctx, keys.NewNodeDocIDToShortIDKey(docID).Bytes())
	if errors.Is(err, corekv.ErrNotFound) {
		return keys.LocalDocID{}, false, nil
	}
	if err != nil {
		return keys.LocalDocID{}, false, err
	}
	localDocID, err := keys.DecodeLocalDocID(value)
	if err != nil {
		return keys.LocalDocID{}, false, err
	}
	return localDocID, true, nil
}

func GetNodeDocIDAliasesForShortDocID(
	ctx context.Context,
	store corekv.Reader,
	collectionShortID uint32,
	shortDocID uint32,
) ([]string, error) {
	if collectionShortID == 0 || shortDocID == 0 {
		return nil, nil
	}

	prefix := keys.NewNodeDocIDToShortIDKey("").ToString() + "/"
	iter, err := store.Iterator(ctx, corekv.IterOptions{Prefix: []byte(prefix)})
	if err != nil {
		return nil, err
	}

	var aliases []string
	for {
		hasNext, err := iter.Next()
		if err != nil {
			return nil, stderrors.Join(err, iter.Close())
		}
		if !hasNext {
			break
		}
		value, err := iter.Value()
		if err != nil {
			return nil, stderrors.Join(err, iter.Close())
		}
		localDocID, err := keys.DecodeLocalDocID(value)
		if err != nil {
			return nil, stderrors.Join(err, iter.Close())
		}
		if localDocID.CollectionShortID == collectionShortID && localDocID.DocShortID == shortDocID {
			aliases = append(aliases, strings.TrimPrefix(string(iter.Key()), prefix))
		}
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return aliases, nil
}

func SetBlockDocIDMapping(
	ctx context.Context,
	collectionShortID uint32,
	blockCID cid.Cid,
	publicDocID string,
) error {
	if !blockCID.Defined() || publicDocID == "" {
		return nil
	}

	txn := datastore.CtxMustGetTxn(ctx)
	return txn.Systemstore().Set(
		ctx,
		keys.NewBlockCIDToDocIDKey(collectionShortID, blockCID.String(), publicDocID).Bytes(),
		[]byte{},
	)
}

func GetDocIDsForBlockFromStore(
	ctx context.Context,
	store corekv.Reader,
	collectionShortID uint32,
	blockCID cid.Cid,
) ([]string, error) {
	if !blockCID.Defined() {
		return nil, nil
	}

	prefix := keys.NewBlockCIDToDocIDKey(collectionShortID, blockCID.String(), "").ToString() + "/"
	iter, err := store.Iterator(ctx, corekv.IterOptions{Prefix: []byte(prefix)})
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
		docID, err := docIDFromBlockMappingSuffix(iter.Key()[len(prefix):], collectionShortID)
		if err != nil {
			return nil, stderrors.Join(err, iter.Close())
		}
		if docID != "" {
			docIDs = append(docIDs, docID)
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
	collectionShortID uint32,
	blockCID cid.Cid,
) (string, bool, error) {
	docIDs, err := GetDocIDsForBlockFromStore(ctx, store, collectionShortID, blockCID)
	if err != nil {
		return "", false, err
	}
	if len(docIDs) == 0 {
		return "", false, nil
	}
	// Only use this helper when the caller has surrounding context that makes any
	// mapped DocID acceptable. Shared genesis field blocks can map to more than one DocID.
	return docIDs[0], true, nil
}

func DeleteBlockDocIDMappings(
	ctx context.Context,
	store corekv.ReaderWriter,
	collectionShortID uint32,
	publicDocID string,
) error {
	if publicDocID == "" {
		return nil
	}

	blockPrefix := keys.NewBlockCIDToDocIDKey(0, "", "").ToString() + "/"
	iter, err := store.Iterator(ctx, corekv.IterOptions{Prefix: []byte(blockPrefix)})
	if err != nil {
		return err
	}

	mappingKeys := make([][]byte, 0)
	for {
		hasNext, err := iter.Next()
		if err != nil {
			return stderrors.Join(err, iter.Close())
		}
		if !hasNext {
			break
		}

		keyCollectionShortID, docID, err := blockMappingKeyDocID(iter.Key()[len(blockPrefix):])
		if err != nil {
			return stderrors.Join(err, iter.Close())
		}
		if keyCollectionShortID != collectionShortID || docID != publicDocID {
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

func docIDFromBlockMappingSuffix(data []byte, collectionShortID uint32) (string, error) {
	if collectionShortID == 0 {
		rest, _, err := keys.DecodeCollectionShortIDPrefix(data)
		if err != nil {
			return "", err
		}
		if len(rest) == 0 || rest[0] != '/' {
			return "", keys.ErrInvalidKey
		}
		return string(rest[1:]), nil
	}
	return string(data), nil
}

func blockMappingKeyDocID(data []byte) (uint32, string, error) {
	blockEnd := bytes.IndexByte(data, '/')
	if blockEnd < 0 {
		return 0, "", keys.ErrInvalidKey
	}
	rest, collectionShortID, err := keys.DecodeCollectionShortIDPrefix(data[blockEnd+1:])
	if err != nil {
		return 0, "", err
	}
	if len(rest) == 0 || rest[0] != '/' {
		return 0, "", keys.ErrInvalidKey
	}
	return collectionShortID, string(rest[1:]), nil
}

func DeleteDocIDMappings(
	ctx context.Context,
	store corekv.ReaderWriter,
	collectionShortID uint32,
	shortDocID uint32,
) error {
	if collectionShortID == 0 || shortDocID == 0 {
		return nil
	}

	publicDocID, found, err := GetDocIDFromStore(ctx, store, collectionShortID, shortDocID)
	if err != nil {
		return err
	}

	if err := deleteKeyIfExists(ctx, store, keys.NewShortIDToDocIDKey(collectionShortID, shortDocID).Bytes()); err != nil {
		return err
	}
	if err := DeleteNodeDocIDAliasesForShortDocID(ctx, store, collectionShortID, shortDocID); err != nil {
		return err
	}
	if found {
		return DeleteBlockDocIDMappings(ctx, store, collectionShortID, publicDocID)
	}
	return nil
}

func DeleteNodeDocIDAliasesForShortDocID(
	ctx context.Context,
	store corekv.ReaderWriter,
	collectionShortID uint32,
	shortDocID uint32,
) error {
	if collectionShortID == 0 || shortDocID == 0 {
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
		localDocID, err := keys.DecodeLocalDocID(value)
		if err != nil {
			return stderrors.Join(err, iter.Close())
		}
		if localDocID.CollectionShortID == collectionShortID && localDocID.DocShortID == shortDocID {
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
