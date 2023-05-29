// Copyright 2022 Democratized Data Foundation
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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/datastore/iterable"
	"github.com/sourcenetwork/defradb/db/base"
)

// Fetcher is the interface for collecting documents from the underlying data store.
// It handles all the key/value scanning, aggregation, and document encoding.
type Fetcher interface {
	Init(col *client.CollectionDescription, fields []*client.FieldDescription, reverse bool, showDeleted bool) error
	Start(ctx context.Context, txn datastore.Txn, spans core.Spans) error
	FetchNext(ctx context.Context) (EncodedDocument, error)
	FetchNextDecoded(ctx context.Context) (*client.Document, error)
	FetchNextDoc(ctx context.Context, mapping *core.DocumentMapping) ([]byte, core.Doc, error)
	Close() error
}

var (
	_ Fetcher = (*DocumentFetcher)(nil)
)

// DocumentFetcher is a utility to incrementally fetch all the documents.
type DocumentFetcher struct {
	col     *client.CollectionDescription
	reverse bool

	txn          datastore.Txn
	spans        core.Spans
	order        []dsq.Order
	curSpanIndex int

	schemaFields map[uint32]client.FieldDescription
	fields       []*client.FieldDescription

	doc         *encodedDocument
	decodedDoc  *client.Document
	initialized bool

	kv                *core.KeyValue
	kvIter            iterable.Iterator
	kvResultsIter     dsq.Results
	kvEnd             bool
	isReadingDocument bool

	// Since deleted documents are stored under a different instance type than active documents,
	// we use a parallel fetcher to be able to return the documents in the expected order.
	// That being lexicographically ordered dockeys.
	deletedDocFetcher *DocumentFetcher
}

// Init implements DocumentFetcher.
func (df *DocumentFetcher) Init(
	col *client.CollectionDescription,
	fields []*client.FieldDescription,
	reverse bool,
	showDeleted bool,
) error {
	if col.Schema.IsEmpty() {
		return client.NewErrUninitializeProperty("DocumentFetcher", "Schema")
	}

	err := df.init(col, fields, reverse)
	if err != nil {
		return err
	}

	if showDeleted {
		if df.deletedDocFetcher == nil {
			df.deletedDocFetcher = new(DocumentFetcher)
		}
		return df.deletedDocFetcher.init(col, fields, reverse)
	}

	return nil
}

func (df *DocumentFetcher) init(
	col *client.CollectionDescription,
	fields []*client.FieldDescription,
	reverse bool,
) error {
	df.col = col
	df.fields = fields
	df.reverse = reverse
	df.initialized = true
	df.isReadingDocument = false
	df.doc = new(encodedDocument)

	if df.kvResultsIter != nil {
		if err := df.kvResultsIter.Close(); err != nil {
			return err
		}
	}
	df.kvResultsIter = nil
	if df.kvIter != nil {
		if err := df.kvIter.Close(); err != nil {
			return err
		}
	}
	df.kvIter = nil

	df.schemaFields = make(map[uint32]client.FieldDescription)
	for _, field := range col.Schema.Fields {
		df.schemaFields[uint32(field.ID)] = field
	}
	return nil
}

func (df *DocumentFetcher) Start(ctx context.Context, txn datastore.Txn, spans core.Spans) error {
	err := df.start(ctx, txn, spans, false)
	if err != nil {
		return err
	}

	if df.deletedDocFetcher != nil {
		return df.deletedDocFetcher.start(ctx, txn, spans, true)
	}

	return nil
}

// Start implements DocumentFetcher.
func (df *DocumentFetcher) start(ctx context.Context, txn datastore.Txn, spans core.Spans, withDeleted bool) error {
	if df.col == nil {
		return client.NewErrUninitializeProperty("DocumentFetcher", "CollectionDescription")
	}
	if df.doc == nil {
		return client.NewErrUninitializeProperty("DocumentFetcher", "Document")
	}

	if !spans.HasValue { // no specified spans so create a prefix scan key for the entire collection
		start := base.MakeCollectionKey(*df.col)
		if withDeleted {
			start = start.WithDeletedFlag()
		} else {
			start = start.WithValueFlag()
		}
		df.spans = core.NewSpans(core.NewSpan(start, start.PrefixEnd()))
	} else {
		valueSpans := make([]core.Span, len(spans.Value))
		for i, span := range spans.Value {
			// We can only handle value keys, so here we ensure we only read value keys
			if withDeleted {
				valueSpans[i] = core.NewSpan(span.Start().WithDeletedFlag(), span.End().WithDeletedFlag())
			} else {
				valueSpans[i] = core.NewSpan(span.Start().WithValueFlag(), span.End().WithValueFlag())
			}
		}

		spans := core.MergeAscending(valueSpans)
		if df.reverse {
			for i, j := 0, len(spans)-1; i < j; i, j = i+1, j-1 {
				spans[i], spans[j] = spans[j], spans[i]
			}
		}
		df.spans = core.NewSpans(spans...)
	}

	df.curSpanIndex = -1
	df.txn = txn

	if df.reverse {
		df.order = []dsq.Order{dsq.OrderByKeyDescending{}}
	} else {
		df.order = []dsq.Order{dsq.OrderByKey{}}
	}

	_, err := df.startNextSpan(ctx)
	return err
}

func (df *DocumentFetcher) startNextSpan(ctx context.Context) (bool, error) {
	nextSpanIndex := df.curSpanIndex + 1
	if nextSpanIndex >= len(df.spans.Value) {
		return false, nil
	}

	var err error
	if df.kvIter == nil {
		df.kvIter, err = df.txn.Datastore().GetIterator(dsq.Query{
			Orders: df.order,
		})
	}
	if err != nil {
		return false, err
	}

	if df.kvResultsIter != nil {
		err = df.kvResultsIter.Close()
		if err != nil {
			return false, err
		}
	}

	span := df.spans.Value[nextSpanIndex]
	df.kvResultsIter, err = df.kvIter.IteratePrefix(ctx, span.Start().ToDS(), span.End().ToDS())
	if err != nil {
		return false, err
	}
	df.curSpanIndex = nextSpanIndex

	_, err = df.nextKey(ctx)
	return err == nil, err
}

func (df *DocumentFetcher) KVEnd() bool {
	return df.kvEnd
}

func (df *DocumentFetcher) KV() *core.KeyValue {
	return df.kv
}

func (df *DocumentFetcher) NextKey(ctx context.Context) (docDone bool, err error) {
	return df.nextKey(ctx)
}

func (df *DocumentFetcher) NextKV() (iterDone bool, kv *core.KeyValue, err error) {
	return df.nextKV()
}

func (df *DocumentFetcher) ProcessKV(kv *core.KeyValue) error {
	return df.processKV(kv)
}

// nextKey gets the next kv. It sets both kv and kvEnd internally.
// It returns true if the current doc is completed
func (df *DocumentFetcher) nextKey(ctx context.Context) (spanDone bool, err error) {
	// get the next kv from nextKV()
	spanDone, df.kv, err = df.nextKV()
	// handle any internal errors
	if err != nil {
		return false, err
	}
	if df.kv != nil && (df.kv.Key.InstanceType != core.ValueKey && df.kv.Key.InstanceType != core.DeletedKey) {
		// We can only ready value values, if we escape the collection's value keys
		// then we must be done and can stop reading
		spanDone = true
	}

	df.kvEnd = spanDone
	if df.kvEnd {
		_, err := df.startNextSpan(ctx)
		if err != nil {
			return false, err
		}
		return true, nil
	}

	// check if we've crossed document boundries
	if df.doc.Key() != nil && df.kv.Key.DocKey != string(df.doc.Key()) {
		df.isReadingDocument = false
		return true, nil
	}
	return false, nil
}

// nextKV is a lower-level utility compared to nextKey. The differences are as follows:
// - It directly interacts with the KVIterator.
// - Returns true if the entire iterator/span is exhausted
// - Returns a kv pair instead of internally updating
func (df *DocumentFetcher) nextKV() (iterDone bool, kv *core.KeyValue, err error) {
	res, available := df.kvResultsIter.NextSync()
	if !available {
		return true, nil, nil
	}
	err = res.Error
	if err != nil {
		return true, nil, err
	}

	dsKey, err := core.NewDataStoreKey(res.Key)
	if err != nil {
		return true, nil, err
	}

	kv = &core.KeyValue{
		Key:   dsKey,
		Value: res.Value,
	}
	return false, kv, nil
}

// processKV continuously processes the key value pairs we've received
// and step by step constructs the current encoded document
func (df *DocumentFetcher) processKV(kv *core.KeyValue) error {
	// skip MerkleCRDT meta-data priority key-value pair
	// implement here <--
	// instance := kv.Key.Name()
	// if instance != "v" {
	// 	return nil
	// }
	if df.doc == nil {
		return client.NewErrUninitializeProperty("DocumentFetcher", "Document")
	}

	if !df.isReadingDocument {
		df.isReadingDocument = true
		df.doc.Reset([]byte(kv.Key.DocKey))
	}

	// we have to skip the object marker
	if bytes.Equal(df.kv.Value, []byte{base.ObjectMarker}) {
		return nil
	}

	// extract the FieldID and update the encoded doc properties map
	fieldID, err := kv.Key.FieldID()
	if err != nil {
		return err
	}
	fieldDesc, exists := df.schemaFields[fieldID]
	if !exists {
		return NewErrFieldIdNotFound(fieldID)
	}

	// @todo: Secondary Index might not have encoded FieldIDs
	// @body: Need to generalized the processKV, and overall Fetcher architecture
	// to better handle dynamic use cases beyond primary indexes. If a
	// secondary index is provided, we need to extract the indexed/implicit fields
	// from the KV pair.
	df.doc.Properties[fieldDesc] = &encProperty{
		Desc: fieldDesc,
		Raw:  kv.Value,
	}
	// @todo: Extract Index implicit/stored keys
	return nil
}

// FetchNext returns a raw binary encoded document. It iterates over all the relevant
// keypairs from the underlying store and constructs the document.
func (df *DocumentFetcher) FetchNext(ctx context.Context) (EncodedDocument, error) {
	if df.kvEnd {
		return nil, nil
	}

	if df.kv == nil {
		return nil, client.NewErrUninitializeProperty("DocumentFetcher", "kv")
	}
	// save the DocKey of the current kv pair so we can track when we cross the doc pair boundries
	// keyparts := df.kv.Key.List()
	// key := keyparts[len(keyparts)-2]

	// iterate until we have collected all the necessary kv pairs for the doc
	// we'll know when were done when either
	// A) Reach the end of the iterator
	for {
		err := df.processKV(df.kv)
		if err != nil {
			return nil, err
		}

		end, err := df.nextKey(ctx)
		if err != nil {
			return nil, err
		}
		if end {
			return df.doc, nil
		}

		// // crossed document kv boundary?
		// // if so, return document
		// newkeyparts := df.kv.Key.List()
		// newKey := newkeyparts[len(newkeyparts)-2]
		// if newKey != key {
		// 	return df.doc, nil
		// }
	}
}

// FetchNextDecoded implements DocumentFetcher
func (df *DocumentFetcher) FetchNextDecoded(ctx context.Context) (*client.Document, error) {
	encdoc, err := df.FetchNext(ctx)
	if err != nil {
		return nil, err
	}
	if encdoc == nil {
		return nil, nil
	}

	df.decodedDoc, err = encdoc.Decode()
	if err != nil {
		return nil, err
	}

	return df.decodedDoc, nil
}

// FetchNextDoc returns the next document as a core.Doc.
// The first return value is the parsed document key.
func (df *DocumentFetcher) FetchNextDoc(
	ctx context.Context,
	mapping *core.DocumentMapping,
) ([]byte, core.Doc, error) {
	var err error
	var encdoc EncodedDocument
	var status client.DocumentStatus

	// If the deletedDocFetcher isn't nil, this means that the user requested to include the deleted documents
	// in the query. To keep the active and deleted docs in lexicographic order of dockeys, we use the two distinct
	// fetchers and fetch the one that has the next lowest (or highest if requested in reverse order) dockey value.
	ddf := df.deletedDocFetcher
	if ddf != nil {
		// If we've reached the end of the deleted docs, we can skip to getting the next active docs.
		if !ddf.kvEnd {
			if df.reverse {
				if df.kvEnd || ddf.kv.Key.DocKey > df.kv.Key.DocKey {
					encdoc, err = ddf.FetchNext(ctx)
					if err != nil {
						return nil, core.Doc{}, err
					}
					status = client.Deleted
				}
			} else {
				if df.kvEnd || ddf.kv.Key.DocKey < df.kv.Key.DocKey {
					encdoc, err = ddf.FetchNext(ctx)
					if err != nil {
						return nil, core.Doc{}, err
					}
					status = client.Deleted
				}
			}
		}
	}

	// At this point id encdoc is nil, it means that the next document to be
	// returned will be from the active ones.
	if encdoc == nil {
		encdoc, err = df.FetchNext(ctx)
		if err != nil {
			return nil, core.Doc{}, err
		}
		if encdoc == nil {
			return nil, core.Doc{}, nil
		}
		status = client.Active
	}

	doc, err := encdoc.DecodeToDoc(mapping)
	if err != nil {
		return nil, core.Doc{}, err
	}
	doc.Status = status
	return encdoc.Key(), doc, err
}

// Close closes the DocumentFetcher.
func (df *DocumentFetcher) Close() error {
	if df.kvIter == nil {
		return nil
	}

	err := df.kvIter.Close()
	if err != nil {
		return err
	}

	if df.kvResultsIter == nil {
		return nil
	}

	err = df.kvResultsIter.Close()
	if err != nil {
		return err
	}

	if df.deletedDocFetcher != nil {
		return df.deletedDocFetcher.Close()
	}

	return nil
}
