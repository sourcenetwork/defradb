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
	"github.com/sourcenetwork/defradb/internal/keys"
)

// documentFetcher is the type responsible for fetching documents from the datastore.
//
// It does not filter the data in any way.
type documentFetcher struct {
	// The set of fields to fetch, mapped by field ID.
	fieldsByID map[uint32]client.FieldDefinition
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
}

var _ fetcher = (*documentFetcher)(nil)

func newDocumentFetcher(
	ctx context.Context,
	txn datastore.Txn,
	fieldsByID map[uint32]client.FieldDefinition,
	prefix keys.DataStoreKey,
	status client.DocumentStatus,
	execInfo *ExecInfo,
) (*documentFetcher, error) {
	if status == client.Active {
		prefix = prefix.WithValueFlag()
	} else if status == client.Deleted {
		prefix = prefix.WithDeletedFlag()
	}

	iter, err := txn.Datastore().Iterator(ctx, corekv.IterOptions{
		Start: prefix.Bytes(),
		End:   prefix.PrefixEnd().Bytes(),
	})
	if err != nil {
		return nil, err
	}

	return &documentFetcher{
		fieldsByID: fieldsByID,
		iter:       iter,
		status:     status,
		execInfo:   execInfo,
	}, nil
}

// keyValue is a KV store response containing the resulting core.DataStoreKey and byte array value.
type keyValue struct {
	Key   keys.DataStoreKey
	Value []byte
}

func (f *documentFetcher) NextDoc() (immutable.Option[string], error) {
	if f.nextKV.HasValue() {
		docID := f.nextKV.Value().Key.DocID
		f.currentKV = f.nextKV.Value()

		f.nextKV = immutable.None[keyValue]()
		f.execInfo.DocsFetched++

		return immutable.Some(docID), nil
	}

	for {
		hasValue, err := f.iter.Next()
		if err != nil || !hasValue {
			return immutable.None[string](), err
		}

		dsKey, err := keys.NewDataStoreKey(string(f.iter.Key()))
		if err != nil {
			return immutable.None[string](), err
		}

		value, err := f.iter.Value()
		if err != nil {
			return immutable.None[string](), err
		}

		previousKV := f.currentKV
		f.currentKV = keyValue{
			Key:   dsKey,
			Value: value,
		}

		if dsKey.DocID != previousKV.Key.DocID {
			break
		}
	}

	f.execInfo.DocsFetched++

	return immutable.Some(f.currentKV.Key.DocID), nil
}

func (f *documentFetcher) GetFields() (immutable.Option[EncodedDocument], error) {
	doc := encodedDocument{}
	doc.id = []byte(f.currentKV.Key.DocID)
	doc.status = f.status
	doc.properties = map[client.FieldDefinition]*encProperty{}

	err := f.appendKV(&doc, f.currentKV)
	if err != nil {
		return immutable.None[EncodedDocument](), err
	}

	for {
		hasValue, err := f.iter.Next()
		if err != nil {
			return immutable.None[EncodedDocument](), err
		}
		if !hasValue {
			break
		}

		dsKey, err := keys.NewDataStoreKey(string(f.iter.Key()))
		if err != nil {
			return immutable.None[EncodedDocument](), err
		}

		value, err := f.iter.Value()
		if err != nil {
			return immutable.None[EncodedDocument](), err
		}

		kv := keyValue{
			Key:   dsKey,
			Value: value,
		}

		if dsKey.DocID != f.currentKV.Key.DocID {
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

func (f *documentFetcher) appendKV(doc *encodedDocument, kv keyValue) error {
	if kv.Key.FieldID == keys.DATASTORE_DOC_VERSION_FIELD_ID {
		doc.schemaVersionID = string(kv.Value)
		return nil
	}

	// we have to skip the object marker
	if bytes.Equal(kv.Value, []byte{base.ObjectMarker}) {
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
