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

	dsq "github.com/ipfs/go-datastore/query"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore/iterable"
	"github.com/sourcenetwork/defradb/internal/db/base"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// document is the type responsible for fetching documents from the datastore.
//
// It does not filter the data in any way.
type document struct {
	// The set of fields to fetch, mapped by field ID.
	fieldsByID map[uint32]client.FieldDefinition
	// The status to assign fetched documents.
	status client.DocumentStatus
	// Statistics on the actions of this instance.
	execInfo *ExecInfo
	// The iterable results that documents will be fetched from.
	kvResultsIter dsq.Results

	// The most recently yielded item from kvResultsIter.
	currentKV keyValue
	// nextKV may hold a datastore key value retrieved from kvResultsIter
	// that was not yet ready to be yielded from the instance.
	//
	// When the next document is requested, this value should be yielded
	// before resuming iteration through the kvResultsIter.
	nextKV immutable.Option[keyValue]
}

var _ fetcher = (*document)(nil)

func newDocumentFetcher(
	ctx context.Context,
	fieldsByID map[uint32]client.FieldDefinition,
	kvIter iterable.Iterator,
	prefix keys.DataStoreKey,
	status client.DocumentStatus,
	execInfo *ExecInfo,
) (*document, error) {
	if status == client.Active {
		prefix = prefix.WithValueFlag()
	} else if status == client.Deleted {
		prefix = prefix.WithDeletedFlag()
	}

	kvResultsIter, err := kvIter.IteratePrefix(ctx, prefix.ToDS(), prefix.PrefixEnd().ToDS())
	if err != nil {
		return nil, err
	}

	return &document{
		fieldsByID:    fieldsByID,
		kvResultsIter: kvResultsIter,
		status:        status,
		execInfo:      execInfo,
	}, nil
}

// keyValue is a KV store response containing the resulting core.DataStoreKey and byte array value.
type keyValue struct {
	Key   keys.DataStoreKey
	Value []byte
}

func (f *document) NextDoc() (immutable.Option[string], error) {
	if f.nextKV.HasValue() {
		docID := f.nextKV.Value().Key.DocID
		f.currentKV = f.nextKV.Value()

		f.nextKV = immutable.None[keyValue]()
		f.execInfo.DocsFetched++

		return immutable.Some(docID), nil
	}

	for {
		res, ok := f.kvResultsIter.NextSync()
		if res.Error != nil {
			return immutable.None[string](), res.Error
		}
		if !ok {
			return immutable.None[string](), nil
		}

		dsKey, err := keys.NewDataStoreKey(res.Key)
		if err != nil {
			return immutable.None[string](), err
		}

		if dsKey.DocID != f.currentKV.Key.DocID {
			f.currentKV = keyValue{
				Key:   dsKey,
				Value: res.Value,
			}
			break
		}
	}

	f.execInfo.DocsFetched++

	return immutable.Some(f.currentKV.Key.DocID), nil
}

func (f *document) GetFields() (immutable.Option[EncodedDocument], error) {
	doc := encodedDocument{}
	doc.id = []byte(f.currentKV.Key.DocID)
	doc.status = f.status
	doc.properties = map[client.FieldDefinition]*encProperty{}

	err := f.appendKv(&doc, f.currentKV)
	if err != nil {
		return immutable.None[EncodedDocument](), err
	}

	for {
		res, ok := f.kvResultsIter.NextSync()
		if !ok {
			break
		}

		dsKey, err := keys.NewDataStoreKey(res.Key)
		if err != nil {
			return immutable.None[EncodedDocument](), err
		}

		kv := keyValue{
			Key:   dsKey,
			Value: res.Value,
		}

		if dsKey.DocID != f.currentKV.Key.DocID {
			f.nextKV = immutable.Some(kv)
			break
		}

		err = f.appendKv(&doc, kv)
		if err != nil {
			return immutable.None[EncodedDocument](), err
		}
	}

	return immutable.Some[EncodedDocument](&doc), nil
}

func (f *document) appendKv(doc *encodedDocument, kv keyValue) error {
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

	fieldDesc, ok := f.fieldsByID[fieldID]
	if !ok {
		return nil
	}

	f.execInfo.FieldsFetched++

	doc.properties[fieldDesc] = &encProperty{
		Desc: fieldDesc,
		Raw:  kv.Value,
	}

	return nil
}

func (f *document) Close() error {
	return f.kvResultsIter.Close()
}
