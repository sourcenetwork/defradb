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

type errorCheckingFilter struct {
	matcher indexMatcher
	err     error
}

func (f *errorCheckingFilter) Filter(e query.Entry) bool {
	if f.err != nil {
		return false
	}
	indexKey, err := core.NewIndexDataStoreKey(e.Key)
	if err != nil {
		f.err = err
		return false
	}
	res, err := f.matcher.Match(indexKey)
	if err != nil {
		f.err = err
		return false
	}
	return res
}

// execInfoIndexMatcherDecorator is a decorator for indexMatcher that counts the number 
// of indexes fetched on every call to Match.
type execInfoIndexMatcherDecorator struct {
	matcher  indexMatcher
	execInfo *ExecInfo
}

func (d *execInfoIndexMatcherDecorator) Match(key core.IndexDataStoreKey) (bool, error) {
	d.execInfo.IndexesFetched++
	return d.matcher.Match(key)
}

type scanningIndexIterator struct {
	queryResultIterator
	indexKey core.IndexDataStoreKey
	matcher  indexMatcher
	filter   errorCheckingFilter
	execInfo *ExecInfo
}

func (i *scanningIndexIterator) Init(ctx context.Context, store datastore.DSReaderWriter) error {
	i.filter.matcher = &execInfoIndexMatcherDecorator{matcher: i.matcher, execInfo: i.execInfo}

	iter, err := store.Query(ctx, query.Query{
		Prefix:   i.indexKey.ToString(),
		KeysOnly: true,
		Filters:  []query.Filter{&i.filter},
	})
	if err != nil {
		return err
	}
	i.resultIter = iter

	return nil
}

func (i *scanningIndexIterator) Next() (core.IndexDataStoreKey, bool, error) {
	key, hasValue, err := i.queryResultIterator.Next()
	if i.filter.err != nil {
		return core.IndexDataStoreKey{}, false, i.filter.err
	}
	return key, hasValue, err
}

// checks if the stored index value satisfies the condition
type indexMatcher interface {
	Match(core.IndexDataStoreKey) (bool, error)
}

// indexByteValuesMatcher is a filter that compares the index value with a given value.
// It uses bytes.Compare to compare the values and evaluate the result with evalFunc.
type indexByteValuesMatcher struct {
	value []byte
	// evalFunc receives a result of bytes.Compare
	evalFunc func(int) bool
}

func (m *indexByteValuesMatcher) Match(key core.IndexDataStoreKey) (bool, error) {
	res := bytes.Compare(key.FieldValues[0], m.value)
	return m.evalFunc(res), nil
}

// matcher if _ne condition is met
type neIndexMatcher struct {
	value []byte
}

func (m *neIndexMatcher) Match(key core.IndexDataStoreKey) (bool, error) {
	return !bytes.Equal(key.FieldValues[0], m.value), nil
}

// checks if the index value is or is not in the given array
type indexInArrayMatcher struct {
	values map[string]bool
	isIn   bool
}

func newNinIndexCmp(values [][]byte, isIn bool) *indexInArrayMatcher {
	valuesMap := make(map[string]bool)
	for _, v := range values {
		valuesMap[string(v)] = true
	}
	return &indexInArrayMatcher{values: valuesMap, isIn: isIn}
}

func (m *indexInArrayMatcher) Match(key core.IndexDataStoreKey) (bool, error) {
	_, found := m.values[string(key.FieldValues[0])]
	return found == m.isIn, nil
}

// checks if the index value satisfies the LIKE condition
type indexLikeMatcher struct {
	hasPrefix   bool
	hasSuffix   bool
	startAndEnd []string
	isLike      bool
	value       string
}

func newLikeIndexCmp(filterValue string, isLike bool) *indexLikeMatcher {
	matcher := &indexLikeMatcher{
		isLike: isLike,
	}
	if len(filterValue) >= 2 {
		if filterValue[0] == '%' {
			matcher.hasPrefix = true
			filterValue = strings.TrimPrefix(filterValue, "%")
		}
		if filterValue[len(filterValue)-1] == '%' {
			matcher.hasSuffix = true
			filterValue = strings.TrimSuffix(filterValue, "%")
		}
		if !matcher.hasPrefix && !matcher.hasSuffix {
			matcher.startAndEnd = strings.Split(filterValue, "%")
		}
	}
	matcher.value = filterValue

	return matcher
}

func (m *indexLikeMatcher) Match(key core.IndexDataStoreKey) (bool, error) {
	var currentVal string
	err := cbor.Unmarshal(key.FieldValues[0], &currentVal)
	if err != nil {
		return false, err
	}

	return m.doesMatch(currentVal) == m.isLike, nil
}

func (m *indexLikeMatcher) doesMatch(currentVal string) bool {
	switch {
	case m.hasPrefix && m.hasSuffix:
		return strings.Contains(currentVal, m.value)
	case m.hasPrefix:
		return strings.HasSuffix(currentVal, m.value)
	case m.hasSuffix:
		return strings.HasPrefix(currentVal, m.value)
	// there might be 2 ends only for LIKE with 1 % in the middle "ab%cd"
	case len(m.startAndEnd) == 2:
		return strings.HasPrefix(currentVal, m.startAndEnd[0]) &&
			strings.HasSuffix(currentVal, m.startAndEnd[1])
	default:
		return m.value == currentVal
	}
}

func createIndexIterator(
	indexDataStoreKey core.IndexDataStoreKey,
	indexFilterConditions *mapper.Filter,
	execInfo *ExecInfo,
) (indexIterator, error) {
	var op string
	var filterVal any
	for _, indexFilterCond := range indexFilterConditions.Conditions {
		condMap := indexFilterCond.(map[connor.FilterKey]any)
		var key connor.FilterKey
		for key, filterVal = range condMap {
			break
		}
		opKey := key.(*mapper.Operator)
		op = opKey.Operation
		break
	}

	switch op {
	case opEq, opGt, opGe, opLt, opLe, opNe:
		writableValue := client.NewCBORValue(client.LWW_REGISTER, filterVal)

		valueBytes, err := writableValue.Bytes()
		if err != nil {
			return nil, err
		}

		switch op {
		case opEq:
			return &eqIndexIterator{
				indexKey:  indexDataStoreKey,
				filterVal: valueBytes,
				execInfo:  execInfo,
			}, nil
		case opGt:
			return &scanningIndexIterator{
				indexKey: indexDataStoreKey,
				matcher: &indexByteValuesMatcher{
					value:    valueBytes,
					evalFunc: func(res int) bool { return res > 0 },
				},
				execInfo: execInfo,
			}, nil
		case opGe:
			return &scanningIndexIterator{
				indexKey: indexDataStoreKey,
				matcher: &indexByteValuesMatcher{
					value:    valueBytes,
					evalFunc: func(res int) bool { return res > 0 || res == 0 },
				},
				execInfo: execInfo,
			}, nil
		case opLt:
			return &scanningIndexIterator{
				indexKey: indexDataStoreKey,
				matcher: &indexByteValuesMatcher{
					value:    valueBytes,
					evalFunc: func(res int) bool { return res < 0 },
				},
				execInfo: execInfo,
			}, nil
		case opLe:
			return &scanningIndexIterator{
				indexKey: indexDataStoreKey,
				matcher: &indexByteValuesMatcher{
					value:    valueBytes,
					evalFunc: func(res int) bool { return res < 0 || res == 0 },
				},
				execInfo: execInfo,
			}, nil
		case opNe:
			return &scanningIndexIterator{
				indexKey: indexDataStoreKey,
				matcher: &neIndexMatcher{
					value: valueBytes,
				},
				execInfo: execInfo,
			}, nil
		}
	case opIn, opNin:
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
			return &scanningIndexIterator{
				indexKey: indexDataStoreKey,
				matcher:  newNinIndexCmp(valArr, false),
				execInfo: execInfo,
			}, nil
		}
	case opLike:
		return &scanningIndexIterator{
			indexKey: indexDataStoreKey,
			matcher:  newLikeIndexCmp(filterVal.(string), true),
			execInfo: execInfo,
		}, nil
	case opNlike:
		return &scanningIndexIterator{
			indexKey: indexDataStoreKey,
			matcher:  newLikeIndexCmp(filterVal.(string), false),
			execInfo: execInfo,
		}, nil
	}

	return nil, errors.New("invalid index filter condition")
}
