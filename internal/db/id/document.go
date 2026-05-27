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
	"context"
	stderrors "errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/ipfs/go-cid"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/keys"
)

const shortDocIDWidth = 16
const genesisDocIDPrefix = "genesis:"

func NewGenesisDocID(docID string) string {
	return genesisDocIDPrefix + docID
}

func IsGenesisDocID(docID string) bool {
	return strings.HasPrefix(docID, genesisDocIDPrefix)
}

func UnwrapGenesisDocID(docID string) string {
	return strings.TrimPrefix(docID, genesisDocIDPrefix)
}

func FormatShortDocID(seq uint64) string {
	return fmt.Sprintf("%016x", seq)
}

func ParseShortDocID(shortDocID string) (uint64, error) {
	if len(shortDocID) == 0 || len(shortDocID) > shortDocIDWidth {
		return 0, strconv.ErrSyntax
	}
	return strconv.ParseUint(shortDocID, 16, 64)
}

func SetDocIDMapping(
	ctx context.Context,
	collectionShortID uint32,
	shortDocID string,
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
		keys.NewDocIDToShortIDKey(collectionShortID, docID).Bytes(),
		[]byte(shortDocID),
	); err != nil {
		return err
	}

	if err := txn.Systemstore().Set(
		ctx,
		keys.NewNodeDocIDToShortIDKey(docID).Bytes(),
		[]byte(shortDocID),
	); err != nil {
		return err
	}

	if err := txn.Systemstore().Set(
		ctx,
		keys.NewNodeShortIDToDocIDKey(shortDocID).Bytes(),
		[]byte(docID),
	); err != nil {
		return err
	}

	docIDIndexKey := keys.NewIndexDataStoreKey(
		collectionShortID,
		keys.DocIDIndexID,
		[]keys.IndexedField{
			{Value: client.NewNormalString(docID)},
			{Value: client.NewNormalString(shortDocID)},
		},
	)
	return txn.Datastore().Set(ctx, &docIDIndexKey, []byte{})
}

func SetDocIDAlias(ctx context.Context, shortDocID string, docID string) error {
	txn := datastore.CtxMustGetTxn(ctx)
	return txn.Systemstore().Set(
		ctx,
		keys.NewNodeDocIDToShortIDKey(docID).Bytes(),
		[]byte(shortDocID),
	)
}

func GetPublicDocID(
	ctx context.Context,
	collectionShortID uint32,
	shortDocID string,
) (string, bool, error) {
	txn := datastore.CtxMustGetTxn(ctx)
	return GetPublicDocIDFromStore(ctx, txn.Systemstore(), collectionShortID, shortDocID)
}

func GetPublicDocIDFromStore(
	ctx context.Context,
	store corekv.Reader,
	collectionShortID uint32,
	shortDocID string,
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
) (string, bool, error) {
	txn := datastore.CtxMustGetTxn(ctx)
	return GetShortDocIDFromStore(ctx, txn.Systemstore(), collectionShortID, docID)
}

func GetShortDocIDFromStore(
	ctx context.Context,
	store corekv.Reader,
	collectionShortID uint32,
	docID string,
) (string, bool, error) {
	value, err := store.Get(ctx, keys.NewDocIDToShortIDKey(collectionShortID, docID).Bytes())
	if errors.Is(err, corekv.ErrNotFound) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return string(value), true, nil
}

func GetNodeShortDocID(ctx context.Context, docID string) (string, bool, error) {
	txn := datastore.CtxMustGetTxn(ctx)
	return GetNodeShortDocIDFromStore(ctx, txn.Systemstore(), docID)
}

func ResolveShortDocID(
	ctx context.Context,
	collectionShortID uint32,
	docID string,
) (string, bool, error) {
	shortDocID, found, err := GetShortDocID(ctx, collectionShortID, docID)
	if err != nil || found {
		return shortDocID, found, err
	}

	shortDocID, found, err = GetNodeShortDocID(ctx, docID)
	if err != nil || !found {
		return "", false, err
	}
	return shortDocID, true, nil
}

func GetNodeShortDocIDFromStore(ctx context.Context, store corekv.Reader, docID string) (string, bool, error) {
	value, err := store.Get(ctx, keys.NewNodeDocIDToShortIDKey(docID).Bytes())
	if errors.Is(err, corekv.ErrNotFound) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return string(value), true, nil
}

func GetNodePublicDocID(ctx context.Context, docID string) (string, bool, error) {
	txn := datastore.CtxMustGetTxn(ctx)
	return GetNodePublicDocIDFromStore(ctx, txn.Systemstore(), docID)
}

func GetNodePublicDocIDFromStore(ctx context.Context, store corekv.Reader, docID string) (string, bool, error) {
	shortDocID, found, err := GetNodeShortDocIDFromStore(ctx, store, docID)
	if err != nil || !found {
		return "", found, err
	}

	value, err := store.Get(ctx, keys.NewNodeShortIDToDocIDKey(shortDocID).Bytes())
	if err == nil {
		return string(value), true, nil
	}
	if !errors.Is(err, corekv.ErrNotFound) {
		return "", false, err
	}

	prefix := keys.DOC_ID_INDEX + "/"
	iter, err := store.Iterator(ctx, corekv.IterOptions{Prefix: []byte(prefix)})
	if err != nil {
		return "", false, err
	}

	for {
		hasNext, err := iter.Next()
		if err != nil {
			return "", false, stderrors.Join(err, iter.Close())
		}
		if !hasNext {
			break
		}

		parts := strings.Split(strings.TrimPrefix(string(iter.Key()), prefix), "/")
		if len(parts) != 3 || parts[1] != keys.SHORT_ID_TO_DOC_ID || parts[2] != shortDocID {
			continue
		}

		value, err := iter.Value()
		if err != nil {
			return "", false, stderrors.Join(err, iter.Close())
		}
		if err := iter.Close(); err != nil {
			return "", false, err
		}
		return string(value), true, nil
	}
	if err := iter.Close(); err != nil {
		return "", false, err
	}
	return "", false, nil
}

func GetNodeDocIDAliasesForShortDocID(
	ctx context.Context,
	store corekv.Reader,
	shortDocID string,
) ([]string, error) {
	if shortDocID == "" {
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
		if string(value) == shortDocID {
			aliases = append(aliases, strings.TrimPrefix(string(iter.Key()), prefix))
		}
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return aliases, nil
}

func SetGenesisFieldDocIDMapping(
	ctx context.Context,
	collectionShortID uint32,
	fieldCID cid.Cid,
	docID string,
) error {
	if !fieldCID.Defined() || docID == "" {
		return nil
	}

	txn := datastore.CtxMustGetTxn(ctx)
	fieldCIDStr := fieldCID.String()
	if err := txn.Systemstore().Set(
		ctx,
		keys.NewGenesisFieldToDocIDKey(collectionShortID, fieldCIDStr, docID).Bytes(),
		[]byte{},
	); err != nil {
		return err
	}

	return txn.Systemstore().Set(
		ctx,
		keys.NewDocIDToGenesisFieldKey(collectionShortID, docID, fieldCIDStr).Bytes(),
		[]byte{},
	)
}

func GetPublicDocIDsForGenesisFieldFromStore(
	ctx context.Context,
	store corekv.Reader,
	collectionShortID uint32,
	fieldCID cid.Cid,
) ([]string, error) {
	if !fieldCID.Defined() {
		return nil, nil
	}

	prefix := keys.NewGenesisFieldToDocIDKey(collectionShortID, fieldCID.String(), "").ToString() + "/"
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
		docID := strings.TrimPrefix(string(iter.Key()), prefix)
		if docID != "" {
			docIDs = append(docIDs, docID)
		}
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return docIDs, nil
}

func GetPublicDocIDForGenesisFieldFromStore(
	ctx context.Context,
	store corekv.Reader,
	collectionShortID uint32,
	fieldCID cid.Cid,
) (string, bool, error) {
	docIDs, err := GetPublicDocIDsForGenesisFieldFromStore(ctx, store, collectionShortID, fieldCID)
	if err != nil {
		return "", false, err
	}
	if len(docIDs) == 0 {
		return "", false, nil
	}
	return docIDs[0], true, nil
}

func DeleteGenesisFieldDocIDMappings(
	ctx context.Context,
	store corekv.ReaderWriter,
	collectionShortID uint32,
	publicDocID string,
) error {
	if publicDocID == "" {
		return nil
	}

	docFieldPrefix := keys.NewDocIDToGenesisFieldKey(collectionShortID, publicDocID, "").ToString() + "/"
	iter, err := store.Iterator(ctx, corekv.IterOptions{Prefix: []byte(docFieldPrefix)})
	if err != nil {
		return err
	}

	type mappingKey struct {
		docToField []byte
		fieldToDoc []byte
	}
	mappingKeys := make([]mappingKey, 0)
	for {
		hasNext, err := iter.Next()
		if err != nil {
			return stderrors.Join(err, iter.Close())
		}
		if !hasNext {
			break
		}

		fieldCID := strings.TrimPrefix(string(iter.Key()), docFieldPrefix)
		if fieldCID == "" {
			continue
		}
		mappingKeys = append(mappingKeys, mappingKey{
			docToField: append([]byte(nil), iter.Key()...),
			fieldToDoc: keys.NewGenesisFieldToDocIDKey(
				collectionShortID,
				fieldCID,
				publicDocID,
			).Bytes(),
		})
	}
	if err := iter.Close(); err != nil {
		return err
	}

	for _, key := range mappingKeys {
		if err := store.Delete(ctx, key.docToField); err != nil && !errors.Is(err, corekv.ErrNotFound) {
			return err
		}
		if err := store.Delete(ctx, key.fieldToDoc); err != nil && !errors.Is(err, corekv.ErrNotFound) {
			return err
		}
	}
	return nil
}

func DeleteNodeDocIDAliasesForShortDocID(
	ctx context.Context,
	store corekv.ReaderWriter,
	shortDocID string,
) error {
	if shortDocID == "" {
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
		if string(value) == shortDocID {
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
