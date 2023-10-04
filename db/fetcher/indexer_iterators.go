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

// indexIterator is an iterator over index keys.
// It is used to iterate over the index keys that match a specific condition.
// For example, iteration over condition _eq and _gt will have completely different logic.
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
	err := i.eqIndexIterator.Init(i.ctx, i.store)
	if err != nil {
		return false, err
	}
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

func createIndexIterator(
	indexDataStoreKey core.IndexDataStoreKey,
	indexFilter *mapper.Filter,
	execInfo *ExecInfo,
) (indexIterator, error) {
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
				indexKey:  indexDataStoreKey,
				filterVal: valueBytes,
				execInfo:  execInfo,
			}, nil
		} else if op == opGt {
			return &filteredIndexIterator{
				indexKey: indexDataStoreKey,
				filter:   &gtIndexCmp{value: valueBytes},
				execInfo: execInfo,
			}, nil
		} else if op == opGe {
			return &filteredIndexIterator{
				indexKey: indexDataStoreKey,
				filter:   &geIndexCmp{value: valueBytes},
				execInfo: execInfo,
			}, nil
		} else if op == opLt {
			return &filteredIndexIterator{
				indexKey: indexDataStoreKey,
				filter:   &ltIndexCmp{value: valueBytes},
				execInfo: execInfo,
			}, nil
		} else if op == opLe {
			return &filteredIndexIterator{
				indexKey: indexDataStoreKey,
				filter:   &leIndexCmp{value: valueBytes},
				execInfo: execInfo,
			}, nil
		} else if op == opNe {
			return &filteredIndexIterator{
				indexKey: indexDataStoreKey,
				filter:   &neIndexCmp{value: valueBytes},
				execInfo: execInfo,
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
			return newInIndexIterator(indexDataStoreKey, valArr, execInfo), nil
		} else {
			return &filteredIndexIterator{
				indexKey: indexDataStoreKey,
				filter:   newNinIndexCmp(valArr, false),
				execInfo: execInfo,
			}, nil
		}
	} else if op == opLike {
		return &filteredIndexIterator{
			indexKey: indexDataStoreKey,
			filter:   newLikeIndexCmp(filterVal.(string), true),
			execInfo: execInfo,
		}, nil
	} else if op == opNlike {
		return &filteredIndexIterator{
			indexKey: indexDataStoreKey,
			filter:   newLikeIndexCmp(filterVal.(string), false),
			execInfo: execInfo,
		}, nil
	}

	return nil, errors.New("invalid index filter condition")
}
