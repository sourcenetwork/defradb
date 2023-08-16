// Copyright 2023 Democratized Data Foundation
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
	"errors"
	"strings"

	"github.com/fxamacker/cbor/v2"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/planner/mapper"

	"github.com/ipfs/go-datastore/query"
)

const (
	opEq    = "_eq"
	opGt    = "_gt"
	opGe    = "_ge"
	opLt    = "_lt"
	opLe    = "_le"
	opNe    = "_ne"
	opIn    = "_in"
	opNin   = "_nin"
	opLike  = "_like"
	opNlike = "_nlike"
)

type IndexFetcher struct {
	docFetcher         Fetcher
	col                *client.CollectionDescription
	txn                datastore.Txn
	filter             *mapper.Filter
	doc                *encodedDocument
	mapping            *core.DocumentMapping
	index              client.IndexDescription
	indexedField       client.FieldDescription
	docFields          []client.FieldDescription
	indexQuery         query.Results
	indexDataStoreKey  core.IndexDataStoreKey
	indexFilterCond    any
	indexQueryProvider filteredIndexQueryProvider
}

var _ Fetcher = (*IndexFetcher)(nil)

func NewIndexFetcher(
	docFetcher Fetcher,
	indexedFieldDesc client.FieldDescription,
	indexDesc client.IndexDescription,
	filterCond any,
) *IndexFetcher {
	return &IndexFetcher{
		docFetcher:      docFetcher,
		indexedField:    indexedFieldDesc,
		index:           indexDesc,
		indexFilterCond: filterCond,
	}
}

type filteredIndexQueryProvider interface {
	Get(context.Context, datastore.Txn) (query.Results, error)
}

type eqIndexQueryProvider struct {
	indexKey  core.IndexDataStoreKey
	filterVal []byte
}

func (i *eqIndexQueryProvider) Get(ctx context.Context, txn datastore.Txn) (query.Results, error) {
	if len(i.indexKey.FieldValues) != 0 {
		return nil, nil
	}

	i.indexKey.FieldValues = [][]byte{i.filterVal}
	return txn.Datastore().Query(ctx, query.Query{
		Prefix:   i.indexKey.ToString(),
		KeysOnly: true,
	})
}

type gtIndexCmp struct {
	value []byte
}

func (cmp *gtIndexCmp) Filter(e query.Entry) bool {
	indexKey, err := core.NewIndexDataStoreKey(e.Key)
	if err != nil {
		return false
	}
	res := bytes.Compare(indexKey.FieldValues[0], cmp.value)
	return res > 0
}

type geIndexCmp struct {
	value []byte
}

func (cmp *geIndexCmp) Filter(e query.Entry) bool {
	indexKey, err := core.NewIndexDataStoreKey(e.Key)
	if err != nil {
		return false
	}
	res := bytes.Compare(indexKey.FieldValues[0], cmp.value)
	return res > 0 || res == 0
}

type ltIndexCmp struct {
	value []byte
}

func (cmp *ltIndexCmp) Filter(e query.Entry) bool {
	indexKey, err := core.NewIndexDataStoreKey(e.Key)
	if err != nil {
		return false
	}
	res := bytes.Compare(indexKey.FieldValues[0], cmp.value)
	return res < 0
}

type leIndexCmp struct {
	value []byte
}

func (cmp *leIndexCmp) Filter(e query.Entry) bool {
	indexKey, err := core.NewIndexDataStoreKey(e.Key)
	if err != nil {
		return false
	}
	res := bytes.Compare(indexKey.FieldValues[0], cmp.value)
	return res < 0 || res == 0
}

type neIndexCmp struct {
	value []byte
}

func (cmp *neIndexCmp) Filter(e query.Entry) bool {
	indexKey, err := core.NewIndexDataStoreKey(e.Key)
	if err != nil {
		return false
	}
	return !bytes.Equal(indexKey.FieldValues[0], cmp.value)
}

type arrIndexCmp struct {
	values map[string]bool
	isIn   bool
}

func newNinIndexCmp(values [][]byte, isIn bool) *arrIndexCmp {
	valuesMap := make(map[string]bool)
	for _, v := range values {
		valuesMap[string(v)] = true
	}
	return &arrIndexCmp{values: valuesMap, isIn: isIn}
}

func (cmp *arrIndexCmp) Filter(e query.Entry) bool {
	indexKey, err := core.NewIndexDataStoreKey(e.Key)
	if err != nil {
		return false
	}
	_, found := cmp.values[string(indexKey.FieldValues[0])]
	return found == cmp.isIn
}

type cmpIndexQueryProvider struct {
	indexKey core.IndexDataStoreKey
	filter   query.Filter
}

func (i *cmpIndexQueryProvider) Get(ctx context.Context, txn datastore.Txn) (query.Results, error) {
	return txn.Datastore().Query(ctx, query.Query{
		Prefix:   i.indexKey.ToString(),
		KeysOnly: true,
		Filters:  []query.Filter{i.filter},
	})
}

type likeIndexCmp struct {
	filterValue string
	hasPrefix   bool
	hasSuffix   bool
	startAndEnd []string
	isLike      bool
}

func newLikeIndexCmp(filterValue string, isLike bool) *likeIndexCmp {
	cmp := &likeIndexCmp{
		filterValue: filterValue,
		isLike:      isLike,
	}
	if len(cmp.filterValue) >= 2 {
		if cmp.filterValue[0] == '%' {
			cmp.hasPrefix = true
			cmp.filterValue = strings.TrimPrefix(cmp.filterValue, "%")
		}
		if cmp.filterValue[len(cmp.filterValue)-1] == '%' {
			cmp.hasSuffix = true
			cmp.filterValue = strings.TrimSuffix(cmp.filterValue, "%")
		}
		if !cmp.hasPrefix && !cmp.hasSuffix {
			cmp.startAndEnd = strings.Split(cmp.filterValue, "%")
		}
	}

	return cmp
}

func (cmp *likeIndexCmp) Filter(e query.Entry) bool {
	indexKey, err := core.NewIndexDataStoreKey(e.Key)
	if err != nil {
		return false
	}

	var value string
	err = cbor.Unmarshal(indexKey.FieldValues[0], &value)
	if err != nil {
		return false
	}

	return cmp.doesMatch(value) == cmp.isLike
}

func (cmp *likeIndexCmp) doesMatch(value string) bool {
	switch {
	case cmp.hasPrefix && cmp.hasSuffix:
		return strings.Contains(value, cmp.filterValue)
	case cmp.hasPrefix:
		return strings.HasSuffix(value, cmp.filterValue)
	case cmp.hasSuffix:
		return strings.HasPrefix(value, cmp.filterValue)
	case len(cmp.startAndEnd) == 2:
		return strings.HasPrefix(value, cmp.startAndEnd[0]) &&
			strings.HasSuffix(value, cmp.startAndEnd[1])
	default:
		return cmp.filterValue == value
	}
}

func (f *IndexFetcher) createFilteredIndexQueryProvider(
	indexFilterCond any,
) (filteredIndexQueryProvider, error) {
	condMap, ok := indexFilterCond.(map[string]any)
	if !ok {
		return nil, errors.New("invalid index filter condition")
	}
	var op string
	var filterVal any
	for op, filterVal = range condMap {
		break
	}

	if op == opEq || op == opGt || op == opGe || op == opLt || op == opLe || op == opNe {
		writableValue := client.NewCBORValue(client.LWW_REGISTER, filterVal)

		valueBytes, err := writableValue.Bytes()
		if err != nil {
			return nil, err
		}

		if op == opEq {
			return &eqIndexQueryProvider{
				indexKey:  f.indexDataStoreKey,
				filterVal: valueBytes,
			}, nil
		} else if op == opGt {
			return &cmpIndexQueryProvider{
				indexKey: f.indexDataStoreKey,
				filter:   &gtIndexCmp{value: valueBytes},
			}, nil
		} else if op == opGe {
			return &cmpIndexQueryProvider{
				indexKey: f.indexDataStoreKey,
				filter:   &geIndexCmp{value: valueBytes},
			}, nil
		} else if op == opLt {
			return &cmpIndexQueryProvider{
				indexKey: f.indexDataStoreKey,
				filter:   &ltIndexCmp{value: valueBytes},
			}, nil
		} else if op == opLe {
			return &cmpIndexQueryProvider{
				indexKey: f.indexDataStoreKey,
				filter:   &leIndexCmp{value: valueBytes},
			}, nil
		} else if op == opNe {
			return &cmpIndexQueryProvider{
				indexKey: f.indexDataStoreKey,
				filter:   &neIndexCmp{value: valueBytes},
			}, nil
		}
	} else if op == opIn || op == opNin {
		inArr, ok := filterVal.([]any)
		if !ok {
			return nil, errors.New("invalid _in/_nin value")
		}
		valArr := make([][]byte, 0, len(inArr))
		for _, v := range inArr {
			writableValue := client.NewCBORValue(client.LWW_REGISTER, v)
			valueBytes, err := writableValue.Bytes()
			if err != nil {
				return nil, err
			}
			valArr = append(valArr, valueBytes)
		}
		if op == opIn {
			return &cmpIndexQueryProvider{
				indexKey: f.indexDataStoreKey,
				filter:   newNinIndexCmp(valArr, true),
			}, nil
		} else {
			return &cmpIndexQueryProvider{
				indexKey: f.indexDataStoreKey,
				filter:   newNinIndexCmp(valArr, false),
			}, nil
		}
	} else if op == opLike {
		return &cmpIndexQueryProvider{
			indexKey: f.indexDataStoreKey,
			filter:   newLikeIndexCmp(filterVal.(string), true),
		}, nil
	} else if op == opNlike {
		return &cmpIndexQueryProvider{
			indexKey: f.indexDataStoreKey,
			filter:   newLikeIndexCmp(filterVal.(string), false),
		}, nil
	}

	return nil, errors.New("invalid index filter condition")
}

func (f *IndexFetcher) Init(
	ctx context.Context,
	txn datastore.Txn,
	col *client.CollectionDescription,
	fields []client.FieldDescription,
	filter *mapper.Filter,
	docMapper *core.DocumentMapping,
	reverse bool,
	showDeleted bool,
) error {
	f.col = col
	f.filter = filter
	f.doc = &encodedDocument{}
	f.mapping = docMapper
	f.txn = txn

	f.indexDataStoreKey.CollectionID = f.col.ID
	f.indexDataStoreKey.IndexID = f.index.ID

	for i := range fields {
		if fields[i].Name == f.indexedField.Name {
			f.docFields = append(fields[:i], fields[i+1:]...)
		}
	}

	queryProvider, err := f.createFilteredIndexQueryProvider(f.indexFilterCond)
	if err != nil {
		return err
	}
	f.indexQueryProvider = queryProvider

	return nil
}

func (f *IndexFetcher) Start(ctx context.Context, spans core.Spans) error {
	var err error
	f.indexQuery, err = f.indexQueryProvider.Get(ctx, f.txn)
	if err != nil {
		return err
	}
	return nil
}

func (f *IndexFetcher) FetchNext(ctx context.Context) (EncodedDocument, ExecInfo, error) {
	f.doc.Reset()

	res, hasValue := f.indexQuery.NextSync()
	if !hasValue || res.Error != nil {
		return nil, ExecInfo{}, res.Error
	}

	indexKey, err := core.NewIndexDataStoreKey(res.Key)
	if err != nil {
		return nil, ExecInfo{}, err
	}
	property := &encProperty{
		Desc: f.indexedField,
		Raw:  indexKey.FieldValues[0],
	}

	f.doc.key = indexKey.FieldValues[1]
	f.doc.properties[f.indexedField] = property

	var resultExecInfo ExecInfo
	if f.docFetcher != nil {
		targetKey := base.MakeDocKey(*f.col, string(f.doc.key))
		spans := core.NewSpans(core.NewSpan(targetKey, targetKey.PrefixEnd()))
		err = f.docFetcher.Init(ctx, f.txn, f.col, f.docFields, f.filter, f.mapping, false, false)
		if err != nil {
			return nil, ExecInfo{}, err
		}
		err = f.docFetcher.Start(ctx, spans)
		if err != nil {
			return nil, ExecInfo{}, err
		}
		encDoc, execInfo, err := f.docFetcher.FetchNext(ctx)
		if err != nil {
			return nil, ExecInfo{}, err
		}
		err = f.docFetcher.Close()
		if err != nil {
			return nil, ExecInfo{}, err
		}
		resultExecInfo.Add(execInfo)
		f.doc.MergeProperties(encDoc)
	}
	return f.doc, resultExecInfo, nil
}

func (f *IndexFetcher) Close() error {
	if f.indexQuery != nil {
		return f.indexQuery.Close()
	}
	return nil
}
