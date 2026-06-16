// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package fetcher

import (
	"bytes"
	"context"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/base"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// documentFetcher is the type responsible for fetching documents from the datastore.
//
// It does not filter the data in any way.
type documentFetcher struct {
	ctx context.Context

	// The set of fields to fetch, mapped by field ID.
	fieldsByID map[uint32]client.CollectionFieldDescription
	// The status to assign fetched documents.
	status client.DocumentStatus
	// Statistics on the actions of this instance.
	execInfo *ExecInfo
	// The iterable results that documents will be fetched from.
	iter corekv.Iterator

	// The most recently yielded item from kvResultsIter.
	currentKV keyValue
	// nextKV may hold a datastore key value retrieved from kvResultsIter
	// that was not yet ready to be yielded from the instance.
	//
	// When the next document is requested, this value should be yielded
	// before resuming iteration through the kvResultsIter.
	nextKV immutable.Option[keyValue]

	// keysOnly indicates that we only need keys(DocID) and not values.
	keysOnly bool
}

var _ fetcher = (*documentFetcher)(nil)

func newDocumentFetcher(
	ctx context.Context,
	txn datastore.Txn,
	fieldsByID map[uint32]client.CollectionFieldDescription,
	prefix keys.DataStoreKey,
	status client.DocumentStatus,
	execInfo *ExecInfo,
) (*documentFetcher, error) {
	switch status {
	case client.Active:
		prefix = prefix.WithValueFlag()
	case client.Deleted:
		prefix = prefix.WithDeletedFlag()
	}

	iterOptions := datastore.IterOptions{
		Start: prefix,
		End:   prefix.PrefixEnd(),
	}

	keysOnly := len(fieldsByID) == 0
	if keysOnly {
		iterOptions.KeysOnly = true
	}

	iter, err := txn.Datastore().Iterator(ctx, iterOptions)
	if err != nil {
		return nil, NewErrCreateDocIterator(err)
	}

	return &documentFetcher{
		ctx:        ctx,
		fieldsByID: fieldsByID,
		iter:       iter,
		status:     status,
		execInfo:   execInfo,
		keysOnly:   keysOnly,
	}, nil
}

// keyValue is a KV store response containing the resulting core.DataStoreKey and byte array value.
type keyValue struct {
	Key   keys.DataStoreKey
	Value []byte
}

func (f *documentFetcher) NextDoc() (immutable.Option[string], error) {
	if f.nextKV.HasValue() {
		kv := f.nextKV.Value()
		f.nextKV = immutable.None[keyValue]()

		if kv.Key.CollectionShortID != 0 && kv.Key.DocShortID != 0 {
			f.currentKV = kv
			f.execInfo.DocsFetched++
			docID, err := f.publicDocID(kv.Key.CollectionShortID, kv.Key.DocShortID)
			if err != nil {
				return immutable.None[string](), err
			}
			return immutable.Some(docID), nil
		}
	}

	for {
		hasValue, err := f.iter.Next()
		if err != nil {
			return immutable.None[string](), NewErrIterateDocuments(err)
		}
		if !hasValue {
			return immutable.None[string](), nil
		}

		dsKey, err := keys.NewDataStoreKey(string(f.iter.Key()))
		if err != nil {
			return immutable.None[string](), NewErrParseDocumentKey(err)
		}
		if dsKey.CollectionShortID == 0 || dsKey.DocShortID == 0 {
			continue
		}

		var value []byte
		if !f.keysOnly {
			value, err = f.iter.Value()
			if err != nil {
				return immutable.None[string](), NewErrGetDocumentValue(err)
			}
		}

		previousKV := f.currentKV
		f.currentKV = keyValue{
			Key:   dsKey,
			Value: value,
		}

		if dsKey.DocShortID != previousKV.Key.DocShortID {
			break
		}
	}

	f.execInfo.DocsFetched++

	docID, err := f.publicDocID(f.currentKV.Key.CollectionShortID, f.currentKV.Key.DocShortID)
	if err != nil {
		return immutable.None[string](), err
	}
	return immutable.Some(docID), nil
}

func (f *documentFetcher) GetFields() (immutable.Option[EncodedDocument], error) {
	if f.currentKV.Key.CollectionShortID == 0 || f.currentKV.Key.DocShortID == 0 {
		return immutable.None[EncodedDocument](), nil
	}

	doc := encodedDocument{}
	docID, err := f.publicDocID(f.currentKV.Key.CollectionShortID, f.currentKV.Key.DocShortID)
	if err != nil {
		return immutable.None[EncodedDocument](), err
	}
	doc.id = []byte(docID)
	doc.status = f.status
	doc.properties = map[client.CollectionFieldDescription]*encProperty{}

	err = f.appendKV(&doc, f.currentKV)
	if err != nil {
		return immutable.None[EncodedDocument](), err
	}

	for {
		hasValue, err := f.iter.Next()
		if err != nil {
			return immutable.None[EncodedDocument](), NewErrIterateDocFields(err)
		}
		if !hasValue {
			break
		}

		dsKey, err := keys.NewDataStoreKey(string(f.iter.Key()))
		if err != nil {
			return immutable.None[EncodedDocument](), NewErrParseFieldKey(err)
		}
		if dsKey.CollectionShortID == 0 || dsKey.DocShortID == 0 {
			continue
		}

		var value []byte
		if !f.keysOnly {
			value, err = f.iter.Value()
			if err != nil {
				return immutable.None[EncodedDocument](), NewErrGetFieldValue(err)
			}
		}

		kv := keyValue{
			Key:   dsKey,
			Value: value,
		}

		if dsKey.DocShortID != f.currentKV.Key.DocShortID {
			f.nextKV = immutable.Some(kv)
			break
		}

		err = f.appendKV(&doc, kv)
		if err != nil {
			return immutable.None[EncodedDocument](), err
		}
	}

	return immutable.Some[EncodedDocument](&doc), nil
}

func (f *documentFetcher) publicDocID(collectionShortID uint32, shortDocID uint64) (string, error) {
	docID, found, err := id.GetPublicDocID(f.ctx, collectionShortID, shortDocID)
	if err != nil {
		return "", err
	}
	if found {
		return docID, nil
	}
	return "", nil
}

func (f *documentFetcher) appendKV(doc *encodedDocument, kv keyValue) error {
	if kv.Key.FieldID == keys.DATASTORE_DOC_VERSION_FIELD_ID {
		doc.collectionVersionID = string(kv.Value)
		return nil
	}

	if bytes.Equal(kv.Value, []byte{base.ObjectMarker}) ||
		bytes.Equal(kv.Value, []byte{base.DeletedObjectMarker}) {
		return nil
	}
	if kv.Key.FieldID == "" {
		return nil
	}

	fieldID, err := kv.Key.FieldIDAsUint()
	if err != nil {
		return err
	}

	// we count the fields fetched here instead of after checking if the field was requested
	// because we need to count all fields fetched to see more accurate picture of the performance
	// of the query
	f.execInfo.FieldsFetched++

	fieldDesc, ok := f.fieldsByID[fieldID]
	if !ok {
		return nil
	}

	doc.properties[fieldDesc] = &encProperty{
		Desc: fieldDesc,
		Raw:  kv.Value,
	}

	return nil
}

func (f *documentFetcher) Close() error {
	return f.iter.Close()
}
