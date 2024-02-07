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
	"strings"

	"github.com/bits-and-blooms/bitset"
	dsq "github.com/ipfs/go-datastore/query"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/datastore/iterable"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/planner/mapper"
	"github.com/sourcenetwork/defradb/request/graphql/parser"
)

// ExecInfo contains statistics about the fetcher execution.
type ExecInfo struct {
	// Number of documents fetched.
	DocsFetched uint64
	// Number of fields fetched.
	FieldsFetched uint64
	// Number of indexes fetched.
	IndexesFetched uint64
}

// Add adds the other ExecInfo to the current ExecInfo.
func (s *ExecInfo) Add(other ExecInfo) {
	s.DocsFetched += other.DocsFetched
	s.FieldsFetched += other.FieldsFetched
	s.IndexesFetched += other.IndexesFetched
}

// Reset resets the ExecInfo.
func (s *ExecInfo) Reset() {
	s.DocsFetched = 0
	s.FieldsFetched = 0
	s.IndexesFetched = 0
}

// Fetcher is the interface for collecting documents from the underlying data store.
// It handles all the key/value scanning, aggregation, and document encoding.
type Fetcher interface {
	Init(
		ctx context.Context,
		txn datastore.Txn,
		acp immutable.Option[acp.ACPModule],
		col client.Collection,
		fields []client.FieldDescription,
		filter *mapper.Filter,
		docmapper *core.DocumentMapping,
		reverse bool,
		showDeleted bool,
	) error
	Start(ctx context.Context, spans core.Spans) error
	FetchNext(ctx context.Context) (EncodedDocument, ExecInfo, error)
	Close() error
}

// keyValue is a KV store response containing the resulting core.Key and byte array value.
type keyValue struct {
	Key   core.DataStoreKey
	Value []byte
}

var (
	_ Fetcher = (*DocumentFetcher)(nil)
)

// DocumentFetcher is a utility to incrementally fetch all the documents.
type DocumentFetcher struct {
	acp         immutable.Option[acp.ACPModule]
	col         client.Collection
	reverse     bool
	deletedDocs bool

	txn          datastore.Txn
	spans        core.Spans
	order        []dsq.Order
	curSpanIndex int

	filter                *mapper.Filter
	passedPermissionCheck bool // have valid permission to access
	ranFilter             bool // did we run the filter
	passedFilter          bool // did we pass the filter

	filterFields map[uint32]client.FieldDescription
	selectFields map[uint32]client.FieldDescription

	// static bitset to which stores the IDs of fields
	// needed for filtering.
	//
	// This is compared against the encdoc.filterSet which
	// is a dynamic bitset, that gets updated as fields are
	// added to the encdoc, and cleared on reset.
	//
	// We compare the two bitsets to determine if we've collected
	// all the necessary fields to run the filter.
	//
	// This is *much* more effecient for comparison then most (any?)
	// other approach.
	//
	// When proper seek() is added, this will also be responsible
	// for effectiently finding the next field to seek to.
	filterSet *bitset.BitSet

	doc     *encodedDocument
	mapping *core.DocumentMapping

	initialized bool

	kv                *keyValue
	kvIter            iterable.Iterator
	kvResultsIter     dsq.Results
	kvEnd             bool
	isReadingDocument bool

	// Since deleted documents are stored under a different instance type than active documents,
	// we use a parallel fetcher to be able to return the documents in the expected order.
	// That being lexicographically ordered docIDs.
	deletedDocFetcher *DocumentFetcher

	execInfo ExecInfo
}

// Init implements DocumentFetcher.
func (df *DocumentFetcher) Init(
	ctx context.Context,
	txn datastore.Txn,
	acp immutable.Option[acp.ACPModule],
	col client.Collection,
	fields []client.FieldDescription,
	filter *mapper.Filter,
	docmapper *core.DocumentMapping,
	reverse bool,
	showDeleted bool,
) error {
	df.txn = txn

	err := df.init(acp, col, fields, filter, docmapper, reverse)
	if err != nil {
		return err
	}

	if showDeleted {
		if df.deletedDocFetcher == nil {
			df.deletedDocFetcher = new(DocumentFetcher)
			df.deletedDocFetcher.txn = txn
		}
		return df.deletedDocFetcher.init(acp, col, fields, filter, docmapper, reverse)
	}

	return nil
}

func (df *DocumentFetcher) init(
	acp immutable.Option[acp.ACPModule],
	col client.Collection,
	fields []client.FieldDescription,
	filter *mapper.Filter,
	docMapper *core.DocumentMapping,
	reverse bool,
) error {
	df.acp = acp
	df.col = col
	df.reverse = reverse
	df.initialized = true
	df.filter = filter
	df.isReadingDocument = false
	df.doc = new(encodedDocument)
	df.mapping = docMapper

	if df.filter != nil && docMapper == nil {
		return ErrMissingMapper
	}

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

	df.selectFields = make(map[uint32]client.FieldDescription, len(fields))
	// if we haven't been told to get specific fields
	// get them all
	var targetFields []client.FieldDescription
	if len(fields) == 0 {
		targetFields = df.col.Schema().Fields
	} else {
		targetFields = fields
	}

	for _, field := range targetFields {
		df.selectFields[uint32(field.ID)] = field
	}

	if df.filter != nil {
		conditions := df.filter.ToMap(df.mapping)
		parsedfilterFields, err := parser.ParseFilterFieldsForDescription(conditions, df.col.Schema())
		if err != nil {
			return err
		}
		df.filterFields = make(map[uint32]client.FieldDescription, len(parsedfilterFields))
		df.filterSet = bitset.New(uint(len(col.Schema().Fields)))
		for _, field := range parsedfilterFields {
			df.filterFields[uint32(field.ID)] = field
			df.filterSet.Set(uint(field.ID))
		}
	}

	return nil
}

func (df *DocumentFetcher) Start(ctx context.Context, spans core.Spans) error {
	err := df.start(ctx, spans, false)
	if err != nil {
		return err
	}

	if df.deletedDocFetcher != nil {
		return df.deletedDocFetcher.start(ctx, spans, true)
	}

	return nil
}

// Start implements DocumentFetcher.
func (df *DocumentFetcher) start(ctx context.Context, spans core.Spans, withDeleted bool) error {
	if df.col == nil {
		return client.NewErrUninitializeProperty("DocumentFetcher", "CollectionDescription")
	}
	if df.doc == nil {
		return client.NewErrUninitializeProperty("DocumentFetcher", "Document")
	}

	df.deletedDocs = withDeleted

	if !spans.HasValue { // no specified spans so create a prefix scan key for the entire collection
		start := base.MakeDataStoreKeyWithCollectionDescription(df.col.Description())
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

	_, _, err = df.nextKey(ctx, false)
	return err == nil, err
}

// nextKey gets the next kv. It sets both kv and kvEnd internally.
// It returns true if the current doc is completed.
// The first call to nextKey CANNOT have seekNext be true (ErrFailedToSeek)
func (df *DocumentFetcher) nextKey(ctx context.Context, seekNext bool) (spanDone bool, docDone bool, err error) {
	// safety against seekNext on first call
	if seekNext && df.kv == nil {
		return false, false, ErrFailedToSeek
	}

	if seekNext {
		curKey := df.kv.Key
		curKey.FieldId = "" // clear field so prefixEnd applies to docID
		seekKey := curKey.PrefixEnd().ToString()
		spanDone, df.kv, err = df.seekKV(seekKey)
		// handle any internal errors
		if err != nil {
			return false, false, err
		}
	} else {
		spanDone, df.kv, err = df.nextKV()
		// handle any internal errors
		if err != nil {
			return false, false, err
		}
	}

	if df.kv != nil && (df.kv.Key.InstanceType != core.ValueKey && df.kv.Key.InstanceType != core.DeletedKey) {
		// We can only ready value values, if we escape the collection's value keys
		// then we must be done and can stop reading
		spanDone = true
	}

	df.kvEnd = spanDone
	if df.kvEnd {
		err = df.kvResultsIter.Close()
		if err != nil {
			return false, false, err
		}
		moreSpans, err := df.startNextSpan(ctx)
		if err != nil {
			return false, false, err
		}
		df.isReadingDocument = false
		return !moreSpans, true, nil
	}

	// check if we've crossed document boundries
	if (df.doc.id != nil && df.kv.Key.DocID != string(df.doc.id)) || seekNext {
		df.isReadingDocument = false
		return false, true, nil
	}
	return false, false, nil
}

// nextKV is a lower-level utility compared to nextKey. The differences are as follows:
// - It directly interacts with the KVIterator.
// - Returns true if the entire iterator/span is exhausted
// - Returns a kv pair instead of internally updating
func (df *DocumentFetcher) nextKV() (iterDone bool, kv *keyValue, err error) {
	done, dsKey, res, err := df.nextKVRaw()
	if done || err != nil {
		return done, nil, err
	}

	kv = &keyValue{
		Key:   dsKey,
		Value: res.Value,
	}
	return false, kv, nil
}

// seekKV will seek through results/iterator until it reaches
// the target key, or if the target key doesn't exist, the
// next smallest key that is greater than the target.
func (df *DocumentFetcher) seekKV(key string) (bool, *keyValue, error) {
	// make sure the current kv is *before* the target key
	switch strings.Compare(df.kv.Key.ToString(), key) {
	case 0:
		// equal, we should just return the kv state
		return df.kvEnd, df.kv, nil
	case 1:
		// greater, error
		return false, nil, NewErrFailedToSeek(key, nil)
	}

	for {
		done, dsKey, res, err := df.nextKVRaw()
		if done || err != nil {
			return done, nil, err
		}

		switch strings.Compare(dsKey.ToString(), key) {
		case -1:
			// before, so lets seek again
			continue
		case 0, 1:
			// equal or greater (first), return a formatted kv
			kv := &keyValue{
				Key:   dsKey,
				Value: res.Value, // @todo make lazy
			}
			return false, kv, nil
		}
	}
}

// nextKV is a lower-level utility compared to nextKey. The differences are as follows:
// - It directly interacts with the KVIterator.
// - Returns true if the entire iterator/span is exhausted
// - Returns a kv pair instead of internally updating
func (df *DocumentFetcher) nextKVRaw() (bool, core.DataStoreKey, dsq.Result, error) {
	res, available := df.kvResultsIter.NextSync()
	if !available {
		return true, core.DataStoreKey{}, res, nil
	}
	err := res.Error
	if err != nil {
		return true, core.DataStoreKey{}, res, err
	}

	dsKey, err := core.NewDataStoreKey(res.Key)
	if err != nil {
		return true, core.DataStoreKey{}, res, err
	}

	return false, dsKey, res, nil
}

// processKV continuously processes the key value pairs we've received
// and step by step constructs the current encoded document
func (df *DocumentFetcher) processKV(kv *keyValue) error {
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
		df.doc.Reset()

		// re-init doc state
		if df.filterSet != nil {
			df.doc.filterSet = bitset.New(df.filterSet.Len())
			if df.filterSet.Test(0) {
				df.doc.filterSet.Set(0) // mark docID as set
			}
		}
		df.doc.id = []byte(kv.Key.DocID)
		df.passedPermissionCheck = false
		df.passedFilter = false
		df.ranFilter = false

		if df.deletedDocs {
			df.doc.status = client.Deleted
		} else {
			df.doc.status = client.Active
		}
	}

	if kv.Key.FieldId == core.DATASTORE_DOC_VERSION_FIELD_ID {
		df.doc.schemaVersionID = string(kv.Value)
		return nil
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
	fieldDesc, exists := df.selectFields[fieldID]
	if !exists {
		fieldDesc, exists = df.filterFields[fieldID]
		if !exists {
			return nil // if we can't find this field in our sets, just ignore it
		}
	}

	ufid := uint(fieldID)

	property := &encProperty{
		Desc: fieldDesc,
		Raw:  kv.Value,
	}

	if df.filterSet != nil && df.filterSet.Test(ufid) {
		df.doc.filterSet.Set(ufid)
		property.IsFilter = true
	}

	df.execInfo.FieldsFetched++

	df.doc.properties[fieldDesc] = property

	return nil
}

// FetchNext returns a raw binary encoded document. It iterates over all the relevant
// keypairs from the underlying store and constructs the document.
func (df *DocumentFetcher) FetchNext(ctx context.Context) (EncodedDocument, ExecInfo, error) {
	var resultExecInfo ExecInfo

	// If the deletedDocFetcher isn't nil, this means that the user requested to include the deleted documents
	// in the query. To keep the active and deleted docs in lexicographic order of docIDs, we use the two distinct
	// fetchers and fetch the one that has the next lowest (or highest if requested in reverse order) docID value.
	ddf := df.deletedDocFetcher
	if ddf != nil {
		// If we've reached the end of the deleted docs, we can skip to getting the next active docs.
		if !ddf.kvEnd {
			if df.kvEnd ||
				(df.reverse && ddf.kv.Key.DocID > df.kv.Key.DocID) ||
				(!df.reverse && ddf.kv.Key.DocID < df.kv.Key.DocID) {
				encdoc, execInfo, err := ddf.FetchNext(ctx)
				resultExecInfo.Add(execInfo)

				if err != nil {
					return nil, ExecInfo{}, err
				}
				if encdoc != nil {
					return encdoc, resultExecInfo, nil
				}
			}
		}
	}

	encdoc, execInfo, err := df.fetchNext(ctx)
	resultExecInfo.Add(execInfo)

	if err != nil {
		return nil, ExecInfo{}, err
	}

	return encdoc, resultExecInfo, err
}

func (df *DocumentFetcher) fetchNext(ctx context.Context) (EncodedDocument, ExecInfo, error) {
	if df.kvEnd {
		return nil, ExecInfo{}, nil
	}

	if df.kv == nil {
		return nil, ExecInfo{}, client.NewErrUninitializeProperty("DocumentFetcher", "kv")
	}

	prevExecInfo := df.execInfo
	defer func() { df.execInfo.Add(prevExecInfo) }()
	df.execInfo.Reset()
	// iterate until we have collected all the necessary kv pairs for the doc
	// we'll know when were done when either
	// A) Reach the end of the iterator
	for {
		if err := df.processKV(df.kv); err != nil {
			return nil, ExecInfo{}, err
		}

		if df.filter != nil {
			// only run filter if we've collected all the fields
			// required for filtering. This is tracked by the bitsets.
			if df.filterSet.Equal(df.doc.filterSet) {
				filterDoc, err := DecodeToDoc(df.doc, df.mapping, true)
				if err != nil {
					return nil, ExecInfo{}, err
				}

				df.ranFilter = true
				df.passedFilter, err = mapper.RunFilter(filterDoc, df.filter)
				if err != nil {
					return nil, ExecInfo{}, err
				}
			}
		}

		// Check if can access document with current permissions/signature.
		if !df.passedPermissionCheck {
			if err := df.runDocReadPermissionCheck(ctx); err != nil {
				return nil, ExecInfo{}, err
			}
		}

		// if we don't pass the filter (ran and pass)
		// theres no point in collecting other select fields
		// so we seek to the next doc
		spansDone, docDone, err := df.nextKey(ctx, !df.passedPermissionCheck || !df.passedFilter && df.ranFilter)

		if err != nil {
			return nil, ExecInfo{}, err
		}

		if !docDone {
			continue
		}

		df.execInfo.DocsFetched++

		if df.passedPermissionCheck {
			if df.filter != nil {
				// if we passed, return
				if df.passedFilter {
					return df.doc, df.execInfo, nil
				} else if !df.ranFilter { // if we didn't run, run it
					decodedDoc, err := DecodeToDoc(df.doc, df.mapping, false)
					if err != nil {
						return nil, ExecInfo{}, err
					}
					df.passedFilter, err = mapper.RunFilter(decodedDoc, df.filter)
					if err != nil {
						return nil, ExecInfo{}, err
					}
					if df.passedFilter {
						return df.doc, df.execInfo, nil
					}
				}
			} else {
				return df.doc, df.execInfo, nil
			}
		}

		if spansDone {
			return nil, df.execInfo, nil
		}
	}
}

// Close closes the DocumentFetcher.
func (df *DocumentFetcher) Close() error {
	if df.kvIter != nil {
		err := df.kvIter.Close()
		if err != nil {
			return err
		}
	}

	if df.kvResultsIter != nil {
		err := df.kvResultsIter.Close()
		if err != nil {
			return err
		}
	}

	if df.deletedDocFetcher != nil {
		return df.deletedDocFetcher.Close()
	}

	return nil
}
