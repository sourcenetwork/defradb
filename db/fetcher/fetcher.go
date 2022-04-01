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
	"context"
	"errors"
	"fmt"

	dsq "github.com/ipfs/go-datastore/query"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/query/graphql/parser"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/base"
)

type fetcherState int

const (
	fetcherFilterGather = iota
	fetcherValueGather
	fetcherSeeking
)

// Fetcher is the interface for collecting documents
// from the underlying data store. It handles all
// the key/value scanning, aggregation, and document
// encoding.
type Fetcher interface {
	Init(col *client.CollectionDescription, index *client.IndexDescription, filter *parser.Filter, reqfields []client.FieldDescription, reverse bool) error
	Start(ctx context.Context, txn datastore.Txn, spans core.Spans) error
	FetchNext(ctx context.Context) (*encodedDocument, error)
	FetchNextDecoded(ctx context.Context) (*client.Document, error)
	FetchNextMap(ctx context.Context) ([]byte, map[string]interface{}, error)
	Close() error
}

var (
	_ Fetcher = (*DocumentFetcher)(nil)
)

type DocumentFetcher struct {
	col     *client.CollectionDescription
	index   *client.IndexDescription
	reverse bool

	txn          datastore.Txn
	spans        core.Spans
	order        []dsq.Order
	uniqueSpans  map[core.Span]struct{} // nolint:structcheck,unused
	curSpanIndex int

	filter *parser.Filter

	schemaFields    map[uint32]client.FieldDescription
	reqFields       map[string]struct{}
	filterFields    map[string]struct{}
	numReqFields    int
	numFilterFields int

	doc         *encodedDocument
	decodedDoc  *client.Document
	initialized bool

	kv     *core.KeyValue
	kvIter dsq.Results // kv val iter that follows the keyIter

	kvEnd bool

	isReadingDocument bool
	state             fetcherState
}

// Init implements DocumentFetcher
func (df *DocumentFetcher) Init(col *client.CollectionDescription, index *client.IndexDescription, filter *parser.Filter, reqFields []client.FieldDescription, reverse bool) error {
	fmt.Println("fetcher init")
	if col.Schema.IsEmpty() {
		return errors.New("DocumentFetcher must be given a schema")
	}

	df.col = col
	df.index = index
	df.reverse = reverse
	df.reqFields = make(map[string]struct{})
	df.doc = new(encodedDocument)
	for _, f := range reqFields {
		// @todo: Sanity check, make sure fid is in schema
		if f.ID == 0 {
			continue // skip _key
		}
		df.reqFields[f.ID.String()] = struct{}{}
		fmt.Printf("adding %s %v to requested fields...\n", f.Name, f.ID)
	}
	df.numReqFields = len(df.reqFields)

	// parse filter fields
	if filter != nil {
		df.filter = filter
		df.filterFields = make(map[string]struct{})
		for k, _ := range df.filter.Conditions {
			field, ok := df.col.GetField(k)
			if !ok {
				// we have an error, filter field not part of description
				return fmt.Errorf("invalid filter field in conditions map: %v", k)
			}

			if field.ID == 0 {
				continue // skip _key
			}

			if field.IsObject() { // skip objects they are handled higherup as a typeJoin
				continue
			}
			df.filterFields[field.ID.String()] = struct{}{}
		}
	}
	df.numFilterFields = len(df.filterFields)

	df.initialized = true
	df.isReadingDocument = false

	// if df.kvIter != nil {
	// 	if err := df.kvResultsIter.Close(); err != nil {
	// 		return err
	// 	}
	// }
	// df.kvResultsIter = nil

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

// newFilterFetcher instantiates a new DocumentFetcher that will retrieve only the fields
// needed for filtering
// func (df *DocumentFetcher) newFilterFetcher(filter *parser.Filter) (*DocumentFetcher, error) {
// 	df.filter = filter
// 	filterFetcher := new(DocumentFetcher)
// 	filterfields := make([]client.FieldDescription, 0, len(filter.Conditions))

// 	for k, _ := range df.filter.Conditions {
// 		field, ok := df.col.GetField(k)
// 		if !ok {
// 			// we have an error, filter field not part of description
// 			return nil, fmt.Errorf("invalid filter field in conditions map: %v", k)
// 		}
// 		filterfields = append(filterfields, field)
// 	}
// 	filterFetcher.Init(df.col, df.index, nil, filterfields, df.reverse)
// 	// df.filterFetcher.doc = df.doc // re-use the same doc for both fetchers
// 	return filterFetcher, nil
// }

// Start implements DocumentFetcher
func (df *DocumentFetcher) Start(ctx context.Context, txn datastore.Txn, spans core.Spans) error {
	if df.col == nil {
		return errors.New("DocumentFetcher cannot be started without a CollectionDescription")
	}
	if df.doc == nil {
		return errors.New("DocumentFetcher cannot be started without an initialized document object")
	}
	if df.index == nil {
		return errors.New("DocumentFetcher cannot be started without a IndexDescription")
	}
	//@todo: Handle fields Description
	// check spans
	numspans := len(spans)
	var uniqueSpans core.Spans
	if numspans == 0 { // no specified spans so create a prefix scan key for the entire collection/index
		start := base.MakeIndexPrefixKey(*df.col, df.index)
		uniqueSpans = core.Spans{core.NewSpan(start, start.PrefixEnd())}
	} else {
		uniqueSpans = spans.MergeAscending()
		if df.reverse {
			for i, j := 0, len(uniqueSpans)-1; i < j; i, j = i+1, j-1 {
				uniqueSpans[i], uniqueSpans[j] = uniqueSpans[j], uniqueSpans[i]
			}
		}
	}

	df.spans = uniqueSpans
	df.curSpanIndex = -1
	df.txn = txn

	if df.reverse {
		df.order = []dsq.Order{dsq.OrderByKeyDescending{}}
	} else {
		df.order = []dsq.Order{dsq.OrderByKey{}}
	}

	if df.filter != nil {
		df.state = fetcherFilterGather // initial state for a fetcher with a filter is FilterGather
	} else {
		df.state = fetcherValueGather
	}

	_, err := df.startNextSpan(ctx)
	return err
}

func (df *DocumentFetcher) startNextSpan(ctx context.Context) (bool, error) {
	nextSpanIndex := df.curSpanIndex + 1
	if nextSpanIndex >= len(df.spans) {
		return false, nil
	}

	var err error
	// if df.kvIter == nil {
	// 	df.kvIter, err = df.txn.Datastore().GetIterator(dsq.Query{
	// 		KeysOnly: true,
	// 		Orders:   df.order,
	// 	})
	// }
	// if err != nil {
	// 	return false, err
	// }

	// if df.keyIter != nil {
	// 	err = df.keyIter.Close()
	// 	if err != nil {
	// 		return false, err
	// 	}
	// }

	if df.kvIter != nil {
		err = df.kvIter.Close()
		if err != nil {
			return false, err
		}
	}

	span := df.spans[nextSpanIndex]

	// df.kvIter, err = df.kvIter.IteratePrefix(ctx, span.Start().ToDS(), span.End().ToDS())

	df.kvIter, err = df.txn.Datastore().Query(ctx, dsq.Query{
		KeysOnly: true,
		Orders:   df.order,
		Prefix:   span.Start().ToDS().String(),
	})
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
//
// Basically,
// Initial call starts with keyOnly iterator
// gets first key
// skip value instance types
// if we need to filter
//		and we havent got all the filter fields
//			iterate until we do
func (df *DocumentFetcher) nextKey(ctx context.Context) (docDone bool, err error) {
	// get the next kv from nextKV()
	for {
		fmt.Println("nextKey loop")
		docDone, df.kv, err = df.nextKV()
		// handle any internal errors
		if err != nil {
			fmt.Println("err7")
			return false, err
		}
		df.kvEnd = docDone
		if df.kvEnd {
			fmt.Println("hi9")
			_, err := df.startNextSpan(ctx)
			if err != nil {
				fmt.Println("err8")
				return false, err
			}
			return true, nil
		}

		// skip if we are iterating through a non value kv pair
		fmt.Println("check skipping instance")
		if df.kv.Key.InstanceType != "v" {
			fmt.Println("skipping non instance:", df.kv.Key.FieldId, df.kv.Key.InstanceType)
			continue
		}
		fmt.Println("didnt skip!")

		// check if we've crossed document boundries
		if df.doc.Key != nil && df.kv.Key.DocKey != string(df.doc.Key) {
			df.isReadingDocument = false
			return true, nil
		}

		// we are either gathering filter fields, requested fields
		// or continue iterating
		switch df.state {
		case fetcherFilterGather:
			if df.IsFilterFieldKey(df.kv.Key) && !df.hasFetchedField(df.kv.Key) {
				df.kv.Value, err = df.kv.Res.ValueCopy(nil)
				if err != nil {
					return false, err
				}
			}

		case fetcherValueGather:
			fmt.Println("checking requeted field for", df.kv.Key.FieldId)
			if df.IsReqFieldKey(df.kv.Key) && !df.hasFetchedField(df.kv.Key) {
				fmt.Println("found requested field for", df.kv.Key.FieldId)
				fmt.Printf("%+v\n", df.kv)
				df.kv.Value, err = df.kv.Res.ValueCopy(nil)
				if err != nil {
					fmt.Println("err6")
					return false, err
				}
				fmt.Println("got value:", df.kv.Value)
			}
		}

		// spew.Dump("COREKEY:", df.kv.Key)
		// fmt.Println("DATASTORE KEY:", df.kv.Key.ToDS())
		// res, err := df.txn.Datastore().Get(ctx, df.kv.Key.ToDS())
		// if err != nil {
		// 	// panic(err)
		// 	return false, err
		// }

		// if df.filter != nil {
		// 	// FetchDoc
		// }

		// check if we need to filter & our key is in the filter set
		// BRANCHY AF - need refactor, POC
		// if df.numFilterFields > 0 {
		// 	if len(df.doc.Properties) < df.numFilterFields {
		// 		fid := df.kv.Key.FieldId
		// 		if _, ok := df.filterFields[fid]; ok {
		// 			df.kv.Value = df.kv.Res.ValueCopy(nil) // lazy load value
		// 		} else {
		// 			continue
		// 		}
		// 	} else {

		// 	}
		// }

		// skip object markers
		// if bytes.Equal(df.kv.Value, []byte{base.ObjectMarker}) {
		// 	continue
		// }

		return false, nil
	}
}

// func (df *DocumentFetcher) hasFetchedField(key core.DataStoreKey) bool {
// 	f, exists := df.schemaFields[key.Fie]
// }

func (df *DocumentFetcher) hasFetchedField(key core.DataStoreKey) bool {
	fid, err := key.FieldID()
	if err != nil {
		panic(err)
	}
	_, exists := df.doc.Properties[client.FieldID(fid)]
	return exists
}

func (df *DocumentFetcher) IsReqFieldKey(key core.DataStoreKey) bool {
	_, exists := df.reqFields[key.FieldId]
	return exists
}

func (df *DocumentFetcher) IsFilterFieldKey(key core.DataStoreKey) bool {
	_, exists := df.filterFields[key.FieldId]
	return exists
}

// func (df *DocumentFetcher) resolveFilterFields(ctx context.Context)

// nextKV is a lower-level utility compared to nextKey. The differences are as follows:
// - It directly interacts with the KVIterator.
// - Returns true if the entire iterator/span is exhausted
// - Returns a kv pair instead of internally updating
func (df *DocumentFetcher) nextKV() (iterDone bool, kv *core.KeyValue, err error) {
	fmt.Println("next sync...")
	res, available := df.kvIter.NextSync()
	if !available {
		return true, nil, nil
	}
	err = res.Error
	if err != nil {
		return true, nil, err
	}

	fmt.Printf("VALUE: %+v\n", res)

	kv = &core.KeyValue{
		Res: res,
		Key: core.NewDataStoreKey(res.Key),
		// Value: res.Value,
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
		return errors.New("Failed to process KV, uninitialized document object")
	}

	if !df.isReadingDocument {
		df.isReadingDocument = true
		df.doc.Reset()
		df.doc.Key = []byte(kv.Key.DocKey)
	}

	// skip if theres no value
	if kv.Value == nil {
		return nil
	}

	// extract the FieldID and update the encoded doc properties map
	fieldID, err := kv.Key.FieldID()
	if err != nil {
		return err
	}
	fieldDesc, exists := df.schemaFields[fieldID]
	if !exists {
		return errors.New("Found field with no matching FieldDescription")
	}

	// @todo: Secondary Index might not have encoded FieldIDs
	// @body: Need to generalized the processKV, and overall Fetcher architecture
	// to better handle dynamic use cases beyond primary indexes. If a
	// secondary index is provided, we need to extract the indexed/implicit fields
	// from the KV pair.
	df.doc.Properties[fieldDesc.ID] = &encProperty{
		Desc: fieldDesc,
		Raw:  kv.Value,
	}
	// @todo: Extract Index implicit/stored keys
	return nil
}

// FetchNext returns a raw binary encoded document. It iterates over all the relevant
// keypairs from the underlying store and constructs the document.
func (df *DocumentFetcher) FetchNext(ctx context.Context) (*encodedDocument, error) {
	if df.kvEnd {
		return nil, nil
	}

	if df.kv == nil {
		return nil, errors.New("Failed to get document, fetcher hasn't been initalized or started")
	}
	// save the DocKey of the current kv pair so we can track when we cross the doc pair boundries
	// keyparts := df.kv.Key.List()
	// key := keyparts[len(keyparts)-2]

	// iterate until we have collected all the necessary kv pairs for the doc
	// we'll know when were done when either
	// A) Reach the end of the iterator
	var end bool
	var err error
	for {
		if end && df.state == fetcherValueGather {
			return df.doc, nil
		}

		if df.state != fetcherSeeking {
			err = df.processKV(df.kv)
			if err != nil {
				return nil, err
			}
		}

		if df.filter != nil && df.state == fetcherFilterGather {
			fmt.Println("CHECKING FILTER FIELD STATE")
			// have we got all the filter fields
			fmt.Printf("numFilterFields: %v, num retrieved properties: %v\n", df.numFilterFields, len(df.doc.Properties))
			if df.numFilterFields != len(df.doc.Properties) && !end {
				end, err = df.nextKey(ctx)
				if err != nil {
					return nil, err
				}
				continue
			}

			// if so, get doc as map, run filter
			doc, err := df.doc.DecodeToMap()
			if err != nil {
				return nil, err
			}

			passed, err := parser.RunFilter(doc, df.filter, parser.EvalContext{})
			if err != nil {
				return nil, err
			}

			if !passed {
				df.state = fetcherSeeking
				break
			} else {
				// seek backwards to start of doc this doc
				fmt.Println(df.kv)
				targetKey := core.DataStoreKey{
					CollectionId: df.kv.Key.CollectionId,
					IndexId:      df.kv.Key.IndexId,
					DocKey:       df.kv.Key.DocKey,
					// FieldId:      "3",
				}
				fmt.Println("Seeking to:", targetKey.ToString())
				df.kvIter.Seek(targetKey.ToString())
				df.state = fetcherValueGather
			}

			// if it passes:
			// - if so, seek back to start of doc boundry
			// - set state to ValueGather
			// else seek till next doc boundry

		}

		if end && df.state == fetcherSeeking {
			df.state = fetcherFilterGather
		} else if end {
			return df.doc, nil
		}

		end, err = df.nextKey(ctx)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
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

// FetchNextMap returns the next document as a map[string]interface{}
// The first return value is the parsed document key
func (df *DocumentFetcher) FetchNextMap(ctx context.Context) ([]byte, map[string]interface{}, error) {
	encdoc, err := df.FetchNext(ctx)
	if err != nil {
		fmt.Println("err3")
		return nil, nil, err
	}
	if encdoc == nil {
		return nil, nil, nil
	}

	doc, err := encdoc.DecodeToMap()
	if err != nil {
		fmt.Println("err4")
		return nil, nil, err
	}
	return encdoc.Key, doc, err
}

func (df *DocumentFetcher) Close() error {
	if df.kvIter == nil {
		return nil
	}

	return df.kvIter.Close()
}
