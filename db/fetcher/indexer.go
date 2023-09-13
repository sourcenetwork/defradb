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
	"github.com/sourcenetwork/defradb/connor"
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
	docFetcher        Fetcher
	col               *client.CollectionDescription
	txn               datastore.Txn
	indexFilter       *mapper.Filter
	docFilter         *mapper.Filter
	doc               *encodedDocument
	mapping           *core.DocumentMapping
	index             client.IndexDescription
	indexedField      client.FieldDescription
	docFields         []client.FieldDescription
	indexIter         indexIterator
	indexDataStoreKey core.IndexDataStoreKey
	execInfo          ExecInfo
}

var _ Fetcher = (*IndexFetcher)(nil)

func NewIndexFetcher(
	docFetcher Fetcher,
	indexedFieldDesc client.FieldDescription,
	indexFilter *mapper.Filter,
) *IndexFetcher {
	return &IndexFetcher{
		docFetcher:   docFetcher,
		indexedField: indexedFieldDesc,
		indexFilter:  indexFilter,
	}
}

type indexIterator interface {
	Init(context.Context, datastore.DSReaderWriter) error
	Next() (core.IndexDataStoreKey, bool, error)
	Close() error
}

type queryResultIterator struct {
	resultIter query.Results
}

func (i queryResultIterator) Next() (core.IndexDataStoreKey, bool, error) {
	res, hasVal := i.resultIter.NextSync()
	if res.Error != nil {
		return core.IndexDataStoreKey{}, false, res.Error
	}
	if !hasVal {
		return core.IndexDataStoreKey{}, false, nil
	}
	key, err := core.NewIndexDataStoreKey(res.Key)
	if err != nil {
		return core.IndexDataStoreKey{}, false, err
	}
	return key, true, nil
}

func (i queryResultIterator) Close() error {
	return i.resultIter.Close()
}

type eqIndexIterator struct {
	queryResultIterator
	indexKey  core.IndexDataStoreKey
	filterVal []byte
	execInfo  *ExecInfo
}

func (i *eqIndexIterator) Init(ctx context.Context, store datastore.DSReaderWriter) error {
	i.indexKey.FieldValues = [][]byte{i.filterVal}
	resultIter, err := store.Query(ctx, query.Query{
		Prefix:   i.indexKey.ToString(),
		KeysOnly: true,
	})
	if err != nil {
		return err
	}
	i.resultIter = resultIter
	return nil
}

func (i *eqIndexIterator) Next() (core.IndexDataStoreKey, bool, error) {
	key, hasValue, err := i.queryResultIterator.Next()
	if hasValue {
		i.execInfo.IndexesFetched++
	}
	return key, hasValue, err
}

type inIndexIterator struct {
	eqIndexIterator
	filterValues [][]byte
	nextValIndex int
	ctx          context.Context
	store        datastore.DSReaderWriter
	hasIterator  bool
}

func newInIndexIterator(
	indexKey core.IndexDataStoreKey,
	filterValues [][]byte,
	execInfo *ExecInfo,
) *inIndexIterator {
	return &inIndexIterator{
		eqIndexIterator: eqIndexIterator{
			indexKey: indexKey,
			execInfo: execInfo,
		},
		filterValues: filterValues,
	}
}

func (i *inIndexIterator) nextIterator() (bool, error) {
	if i.nextValIndex > 0 {
		err := i.eqIndexIterator.Close()
		if err != nil {
			return false, err
		}
	}

	if i.nextValIndex >= len(i.filterValues) {
		return false, nil
	}

	i.filterVal = i.filterValues[i.nextValIndex]
	i.eqIndexIterator.Init(i.ctx, i.store)
	i.nextValIndex++
	return true, nil
}

func (i *inIndexIterator) Init(ctx context.Context, store datastore.DSReaderWriter) error {
	i.ctx = ctx
	i.store = store
	var err error
	i.hasIterator, err = i.nextIterator()
	return err
}

func (i *inIndexIterator) Next() (core.IndexDataStoreKey, bool, error) {
	for i.hasIterator {
		key, hasValue, err := i.eqIndexIterator.Next()
		if err != nil {
			return core.IndexDataStoreKey{}, false, err
		}
		if !hasValue {
			i.hasIterator, err = i.nextIterator()
			if err != nil {
				return core.IndexDataStoreKey{}, false, err
			}
			continue
		}
		return key, true, nil
	}
	return core.IndexDataStoreKey{}, false, nil
}

func (i *inIndexIterator) Close() error {
	return nil
}

type filteredIndexIterator struct {
	queryResultIterator
	indexKey core.IndexDataStoreKey
	filter   query.Filter
	execInfo *ExecInfo
}

type fetchCounterFilterDecorator struct {
	filter   query.Filter
	execInfo *ExecInfo
}

func (f *fetchCounterFilterDecorator) Filter(e query.Entry) bool {
	f.execInfo.IndexesFetched++
	return f.filter.Filter(e)
}

func (i *filteredIndexIterator) Init(ctx context.Context, store datastore.DSReaderWriter) error {
	iter, err := store.Query(ctx, query.Query{
		Prefix:   i.indexKey.ToString(),
		KeysOnly: true,
		Filters: []query.Filter{
			&fetchCounterFilterDecorator{filter: i.filter, execInfo: i.execInfo},
		},
	})
	if err != nil {
		return err
	}
	i.resultIter = iter

	return nil
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

func (f *IndexFetcher) createIndexIterator(indexFilter *mapper.Filter) (indexIterator, error) {
	var op string
	var filterVal any
	for _, indexFilterCond := range indexFilter.Conditions {
		condMap := indexFilterCond.(map[connor.FilterKey]any)
		var key connor.FilterKey
		for key, filterVal = range condMap {
			break
		}
		opKey := key.(*mapper.Operator)
		op = opKey.Operation
		break
	}

	if op == opEq || op == opGt || op == opGe || op == opLt || op == opLe || op == opNe {
		writableValue := client.NewCBORValue(client.LWW_REGISTER, filterVal)

		valueBytes, err := writableValue.Bytes()
		if err != nil {
			return nil, err
		}

		if op == opEq {
			return &eqIndexIterator{
				indexKey:  f.indexDataStoreKey,
				filterVal: valueBytes,
				execInfo:  &f.execInfo,
			}, nil
		} else if op == opGt {
			return &filteredIndexIterator{
				indexKey: f.indexDataStoreKey,
				filter:   &gtIndexCmp{value: valueBytes},
				execInfo: &f.execInfo,
			}, nil
		} else if op == opGe {
			return &filteredIndexIterator{
				indexKey: f.indexDataStoreKey,
				filter:   &geIndexCmp{value: valueBytes},
				execInfo: &f.execInfo,
			}, nil
		} else if op == opLt {
			return &filteredIndexIterator{
				indexKey: f.indexDataStoreKey,
				filter:   &ltIndexCmp{value: valueBytes},
				execInfo: &f.execInfo,
			}, nil
		} else if op == opLe {
			return &filteredIndexIterator{
				indexKey: f.indexDataStoreKey,
				filter:   &leIndexCmp{value: valueBytes},
				execInfo: &f.execInfo,
			}, nil
		} else if op == opNe {
			return &filteredIndexIterator{
				indexKey: f.indexDataStoreKey,
				filter:   &neIndexCmp{value: valueBytes},
				execInfo: &f.execInfo,
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
			return newInIndexIterator(f.indexDataStoreKey, valArr, &f.execInfo), nil
		} else {
			return &filteredIndexIterator{
				indexKey: f.indexDataStoreKey,
				filter:   newNinIndexCmp(valArr, false),
				execInfo: &f.execInfo,
			}, nil
		}
	} else if op == opLike {
		return &filteredIndexIterator{
			indexKey: f.indexDataStoreKey,
			filter:   newLikeIndexCmp(filterVal.(string), true),
			execInfo: &f.execInfo,
		}, nil
	} else if op == opNlike {
		return &filteredIndexIterator{
			indexKey: f.indexDataStoreKey,
			filter:   newLikeIndexCmp(filterVal.(string), false),
			execInfo: &f.execInfo,
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
	f.docFilter = filter
	f.doc = &encodedDocument{}
	f.mapping = docMapper
	f.txn = txn

	for _, index := range col.Indexes {
		if index.Fields[0].Name == f.indexedField.Name {
			f.index = index
			break
		}
	}

	f.indexDataStoreKey.CollectionID = f.col.ID
	f.indexDataStoreKey.IndexID = f.index.ID

	for i := range fields {
		if fields[i].Name == f.indexedField.Name {
			f.docFields = append(fields[:i], fields[i+1:]...)
			break
		}
	}

	iter, err := f.createIndexIterator(f.indexFilter)
	if err != nil {
		return err
	}
	f.indexIter = iter

	if f.docFetcher != nil && len(f.docFields) > 0 {
		err = f.docFetcher.Init(ctx, f.txn, f.col, f.docFields, f.docFilter, f.mapping, false, false)
	}

	return err
}

func (f *IndexFetcher) Start(ctx context.Context, spans core.Spans) error {
	err := f.indexIter.Init(ctx, f.txn.Datastore())
	if err != nil {
		return err
	}
	return nil
}

func (f *IndexFetcher) FetchNext(ctx context.Context) (EncodedDocument, ExecInfo, error) {
	f.execInfo.Reset()
	for {
		f.doc.Reset()

		indexKey, hasValue, err := f.indexIter.Next()
		if err != nil {
			return nil, ExecInfo{}, err
		}

		if !hasValue {
			return nil, f.execInfo, nil
		}

		property := &encProperty{
			Desc: f.indexedField,
			Raw:  indexKey.FieldValues[0],
		}

		f.doc.key = indexKey.FieldValues[1]
		f.doc.properties[f.indexedField] = property
		f.execInfo.FieldsFetched++

		if f.docFetcher != nil && len(f.docFields) > 0 {
			targetKey := base.MakeDocKey(*f.col, string(f.doc.key))
			spans := core.NewSpans(core.NewSpan(targetKey, targetKey.PrefixEnd()))
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
			f.execInfo.Add(execInfo)
			if encDoc == nil {
				continue
			}
			f.doc.MergeProperties(encDoc)
		} else {
			f.execInfo.DocsFetched++
		}
		return f.doc, f.execInfo, nil
	}
}

func (f *IndexFetcher) Close() error {
	if f.indexIter != nil {
		return f.indexIter.Close()
	}
	return nil
}
