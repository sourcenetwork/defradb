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
	"context"
	"strings"
	"time"
	"errors"

	ds "github.com/ipfs/go-datastore"
	"golang.org/x/exp/slices"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/connor"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"

	"github.com/ipfs/go-datastore/query"
)

const (
	opEq       = "_eq"
	opGt       = "_gt"
	opGe       = "_ge"
	opLt       = "_lt"
	opLe       = "_le"
	opNe       = "_ne"
	opIn       = "_in"
	opNin      = "_nin"
	opLike     = "_like"
	opNlike    = "_nlike"
	opILike    = "_ilike"
	opNILike   = "_nilike"
	compOpAny  = "_any"
	compOpAll  = "_all"
	compOpNone = "_none"
	// it's just there for composite indexes. We construct a slice of value matchers with
	// every matcher being responsible for a corresponding field in the index to match.
	// For some fields there might not be any criteria to match. For examples if you have
	// composite index of /name/age/email/ and in the filter you specify only "name" and "email".
	// Then the "__any" matcher will be used for "age".
	opAny = "__any"
)

// indexIterator is an iterator over index keys.
// It is used to iterate over the index keys that match a specific condition.
// For example, iteration over condition _eq and _gt will have completely different logic.
type indexIterator interface {
	Init(context.Context, datastore.DSReaderWriter) error
	Next() (indexIterResult, error)
	Close() error
}

type indexIterResult struct {
	key      keys.IndexDataStoreKey
	foundKey bool
	value    []byte
}

// indexPrefixIterator is an iterator over index keys with a specific prefix.
type indexPrefixIterator struct {
	indexDesc     client.IndexDescription
	indexedFields []client.FieldDefinition
	indexKey      keys.IndexDataStoreKey
	matchers      []valueMatcher
	execInfo      *ExecInfo
	resultIter    query.Results
	ctx           context.Context
	store         datastore.DSReaderWriter
}

var _ indexIterator = (*indexPrefixIterator)(nil)

func (iter *indexPrefixIterator) Init(ctx context.Context, store datastore.DSReaderWriter) error {
	iter.ctx = ctx
	iter.store = store
	if iter.resultIter != nil {
		if err := iter.resultIter.Close(); err != nil {
			return err
		}
	}
	iter.resultIter = nil
	return nil
}

func (iter *indexPrefixIterator) checkResultIterator() error {
	if iter.resultIter == nil {
		resultIter, err := iter.store.Query(iter.ctx, query.Query{
			Prefix: iter.indexKey.ToString(),
		})
		if err != nil {
			return err
		}
		iter.resultIter = resultIter
	}
	return nil
}

func (iter *indexPrefixIterator) nextResult() (indexIterResult, error) {
	res, hasVal := iter.resultIter.NextSync()
	if res.Error != nil {
		return indexIterResult{}, res.Error
	}
	if !hasVal {
		return indexIterResult{}, nil
	}
	key, err := keys.DecodeIndexDataStoreKey([]byte(res.Key), &iter.indexDesc, iter.indexedFields)
	if err != nil {
		return indexIterResult{}, err
	}

	return indexIterResult{key: key, value: res.Value, foundKey: true}, nil
}

func (iter *indexPrefixIterator) Next() (indexIterResult, error) {
	if err := iter.checkResultIterator(); err != nil {
		return indexIterResult{}, err
	}

	for {
		res, err := iter.nextResult()
		if err != nil || !res.foundKey {
			return res, err
		}
		iter.execInfo.IndexesFetched++
		didMatch, err := executeValueMatchers(iter.matchers, res.key.Fields)
		if err != nil {
			return indexIterResult{}, err
		}
		if didMatch {
			return res, nil
		}
	}
}

func (iter *indexPrefixIterator) Close() error {
	if iter.resultIter == nil {
		return nil
	}
	return iter.resultIter.Close()
}

type eqSingleIndexIterator struct {
	indexKey keys.IndexDataStoreKey
	execInfo *ExecInfo

	ctx   context.Context
	store datastore.DSReaderWriter
}

var _ indexIterator = (*eqSingleIndexIterator)(nil)

func (iter *eqSingleIndexIterator) Init(ctx context.Context, store datastore.DSReaderWriter) error {
	iter.ctx = ctx
	iter.store = store
	return nil
}

func (iter *eqSingleIndexIterator) Next() (indexIterResult, error) {
	if iter.store == nil {
		return indexIterResult{}, nil
	}
	val, err := iter.store.Get(iter.ctx, iter.indexKey.ToDS())
	if err != nil {
		if errors.Is(err, ds.ErrNotFound) {
			return indexIterResult{key: iter.indexKey}, nil
		}
		return indexIterResult{}, err
	}
	iter.store = nil
	iter.execInfo.IndexesFetched++
	return indexIterResult{key: iter.indexKey, value: val, foundKey: true}, nil
}

func (iter *eqSingleIndexIterator) Close() error {
	return nil
}

type inIndexIterator struct {
	indexIterator
	inValues     []client.NormalValue
	nextValIndex int
	ctx          context.Context
	store        datastore.DSReaderWriter
	hasIterator  bool
}

var _ indexIterator = (*inIndexIterator)(nil)

func (iter *inIndexIterator) nextIterator() (bool, error) {
	if iter.nextValIndex > 0 {
		err := iter.indexIterator.Close()
		if err != nil {
			return false, err
		}
	}

	if iter.nextValIndex >= len(iter.inValues) {
		return false, nil
	}

	switch fieldIter := iter.indexIterator.(type) {
	case *indexPrefixIterator:
		fieldIter.indexKey.Fields[0].Value = iter.inValues[iter.nextValIndex]
	case *eqSingleIndexIterator:
		fieldIter.indexKey.Fields[0].Value = iter.inValues[iter.nextValIndex]
	}
	err := iter.indexIterator.Init(iter.ctx, iter.store)
	if err != nil {
		return false, err
	}
	iter.nextValIndex++
	return true, nil
}

func (iter *inIndexIterator) Init(ctx context.Context, store datastore.DSReaderWriter) error {
	iter.ctx = ctx
	iter.store = store
	var err error
	iter.hasIterator, err = iter.nextIterator()
	return err
}

func (iter *inIndexIterator) Next() (indexIterResult, error) {
	for iter.hasIterator {
		res, err := iter.indexIterator.Next()
		if err != nil {
			return indexIterResult{}, err
		}
		if !res.foundKey {
			iter.hasIterator, err = iter.nextIterator()
			if err != nil {
				return indexIterResult{}, err
			}
			continue
		}
		return res, nil
	}
	return indexIterResult{}, nil
}

func (iter *inIndexIterator) Close() error {
	return nil
}

// arrayIndexIterator is an iterator indexed array elements.
// It keeps track of the already fetched documents to avoid duplicates.
type arrayIndexIterator struct {
	inner indexIterator

	fetchedDocs map[string]struct{}

	ctx   context.Context
	store datastore.DSReaderWriter
}

var _ indexIterator = (*arrayIndexIterator)(nil)

func (iter *arrayIndexIterator) Init(ctx context.Context, store datastore.DSReaderWriter) error {
	iter.ctx = ctx
	iter.store = store
	iter.fetchedDocs = make(map[string]struct{})
	return iter.inner.Init(ctx, store)
}

func (iter *arrayIndexIterator) Next() (indexIterResult, error) {
	for {
		res, err := iter.inner.Next()
		if err != nil {
			return indexIterResult{}, err
		}
		if !res.foundKey {
			return res, nil
		}
		var docID string
		if len(res.value) > 0 {
			docID = string(res.value)
		} else {
			lastField := &res.key.Fields[len(res.key.Fields)-1]
			var ok bool
			docID, ok = lastField.Value.String()
			if !ok {
				return indexIterResult{}, NewErrUnexpectedTypeValue[string](lastField.Value)
			}
		}
		if _, ok := iter.fetchedDocs[docID]; ok {
			continue
		}
		iter.fetchedDocs[docID] = struct{}{}
		return res, nil
	}
}

func (iter *arrayIndexIterator) Close() error {
	return iter.inner.Close()
}

type jsonIndexIterator struct {
	inner indexIterator

	fetchedDocs map[string]struct{}
	jsonPath    []string

	ctx   context.Context
	store datastore.DSReaderWriter
}

var _ indexIterator = (*jsonIndexIterator)(nil)

func (iter *jsonIndexIterator) Init(ctx context.Context, store datastore.DSReaderWriter) error {
	iter.ctx = ctx
	iter.store = store
	iter.fetchedDocs = make(map[string]struct{})
	return iter.inner.Init(ctx, store)
}

func (iter *jsonIndexIterator) Next() (indexIterResult, error) {
	for {
		res, err := iter.inner.Next()
		if err != nil {
			return indexIterResult{}, err
		}
		if !res.foundKey {
			return res, nil
		}
		var docID string
		if len(res.value) > 0 {
			docID = string(res.value)
		} else {
			lastField := &res.key.Fields[len(res.key.Fields)-1]
			var ok bool
			docID, ok = lastField.Value.String()
			if !ok {
				return indexIterResult{}, NewErrUnexpectedTypeValue[string](lastField.Value)
			}
		}
		if _, ok := iter.fetchedDocs[docID]; ok {
			continue
		}
		iter.fetchedDocs[docID] = struct{}{}
		return res, nil
	}
}

func (iter *jsonIndexIterator) Close() error {
	return iter.inner.Close()
}

// newPrefixIteratorFromConditions creates a new eqPrefixIndexIterator for fetching indexed data.
// It can modify the input matchers slice.
func (f *IndexFetcher) newPrefixIteratorFromConditions(
	fieldConditions []fieldFilterCond,
	matchers []valueMatcher,
) (*indexPrefixIterator, error) {
	keyFieldValues := make([]client.NormalValue, 0, len(fieldConditions))
	for i := range fieldConditions {
		c := &fieldConditions[i]
		if c.op != opEq || c.arrOp == compOpNone {
			// prefix can be created only for subsequent _eq conditions
			// if we encounter any other condition, we built the longest prefix we could
			break
		}

		keyFieldValues = append(keyFieldValues, fieldConditions[i].val)
	}

	// iterators for _eq filter already iterate over keys with first field value
	// matching the filter value, so we can skip the first matcher
	if len(matchers) > 1 {
		matchers[0] = &anyMatcher{}
	}

	key := f.newIndexDataStoreKeyWithValues(keyFieldValues)

	return f.newPrefixIterator(key, matchers, &f.execInfo), nil
}

func (f *IndexFetcher) newPrefixIterator(
	indexKey keys.IndexDataStoreKey,
	matchers []valueMatcher,
	execInfo *ExecInfo,
) *indexPrefixIterator {
	return &indexPrefixIterator{
		indexDesc:     f.indexDesc,
		indexedFields: f.indexedFields,
		indexKey:      indexKey,
		matchers:      matchers,
		execInfo:      execInfo,
	}
}

// newInIndexIterator creates a new inIndexIterator for fetching indexed data.
// It can modify the input matchers slice.
func (f *IndexFetcher) newInIndexIterator(
	fieldConditions []fieldFilterCond,
	matchers []valueMatcher,
) (*inIndexIterator, error) {
	if !fieldConditions[0].val.IsArray() {
		return nil, ErrInvalidInOperatorValue
	}
	inValues, err := client.ToArrayOfNormalValues(fieldConditions[0].val)
	if err != nil {
		return nil, err
	}

	// iterators for _in filter already iterate over keys with first field value
	// matching the filter value, so we can skip the first matcher
	if len(matchers) > 1 {
		matchers[0] = &anyMatcher{}
	}

	var iter indexIterator
	if isUniqueFetchByFullKey(&f.indexDesc, fieldConditions) {
		keyFieldValues := make([]client.NormalValue, len(fieldConditions))
		for i := range fieldConditions {
			keyFieldValues[i] = fieldConditions[i].val
		}

		key := f.newIndexDataStoreKeyWithValues(keyFieldValues)

		iter = &eqSingleIndexIterator{
			indexKey: key,
			execInfo: &f.execInfo,
		}
	} else {
		indexKey := f.newIndexDataStoreKey()
		indexKey.Fields = []keys.IndexedField{{Descending: f.indexDesc.Fields[0].Descending}}

		iter = f.newPrefixIterator(indexKey, matchers, &f.execInfo)
	}
	return &inIndexIterator{
		indexIterator: iter,
		inValues:      inValues,
	}, nil
}

func (f *IndexFetcher) newIndexDataStoreKey() keys.IndexDataStoreKey {
	return keys.IndexDataStoreKey{CollectionID: f.col.ID(), IndexID: f.indexDesc.ID}
}

func (f *IndexFetcher) newIndexDataStoreKeyWithValues(values []client.NormalValue) keys.IndexDataStoreKey {
	fields := make([]keys.IndexedField, len(values))
	for i := range values {
		fields[i].Value = values[i]
		fields[i].Descending = f.indexDesc.Fields[i].Descending
	}
	return keys.NewIndexDataStoreKey(f.col.ID(), f.indexDesc.ID, fields)
}

func (f *IndexFetcher) createIndexIterator() (indexIterator, error) {
	fieldConditions, err := f.determineFieldFilterConditions()
	if err != nil {
		return nil, err
	}

	// this can happen if a query contains an empty condition like User(filter: {name: {}})
	if len(fieldConditions) == 0 {
		return nil, nil
	}

	matchers, err := createValueMatchers(fieldConditions)
	if err != nil {
		return nil, err
	}

	hasArray := false
	for i := range fieldConditions {
		if len(fieldConditions[i].arrOp) > 0 {
			hasArray = true
			if fieldConditions[i].arrOp == compOpNone {
				matchers[i] = &invertedMatcher{matcher: matchers[i]}
			}
		}
	}

	hasJSON := false
	for i := range fieldConditions {
		if fieldConditions[i].kind == client.FieldKind_NILLABLE_JSON {
			hasJSON = true
			break
		}
	}

	var iter indexIterator

	if fieldConditions[0].op == opEq {
		if isUniqueFetchByFullKey(&f.indexDesc, fieldConditions) {
			keyFieldValues := make([]client.NormalValue, len(fieldConditions))
			for i := range fieldConditions {
				keyFieldValues[i] = fieldConditions[i].val
			}

			key := f.newIndexDataStoreKeyWithValues(keyFieldValues)
			iter = &eqSingleIndexIterator{indexKey: key, execInfo: &f.execInfo}
		} else {
			iter, err = f.newPrefixIteratorFromConditions(fieldConditions, matchers)
		}
	} else if fieldConditions[0].op == opIn && fieldConditions[0].arrOp != compOpNone {
		iter, err = f.newInIndexIterator(fieldConditions, matchers)
	} else {
		key := f.newIndexDataStoreKey()
		// TODO: can we test fieldConditions[not 0]?
		if fieldConditions[0].kind == client.FieldKind_NILLABLE_JSON {
			key.Fields = []keys.IndexedField{{
				Descending: f.indexDesc.Fields[0].Descending,
				Value:      client.NewNormalJSON(client.MakeVoidJSON(fieldConditions[0].jsonPath)),
			}}
		}
		iter, err = f.newPrefixIterator(key, matchers, &f.execInfo), nil
	}

	if err != nil {
		return nil, err
	}

	if iter == nil {
		return nil, NewErrInvalidFilterOperator(fieldConditions[0].op)
	}

	if hasJSON {
		iter = &jsonIndexIterator{inner: iter, jsonPath: fieldConditions[0].jsonPath}
	} else if hasArray {
		iter = &arrayIndexIterator{inner: iter}
	}

	return iter, nil
}

type fieldFilterCond struct {
	op       string
	arrOp    string
	jsonPath []string
	val      client.NormalValue
	kind     client.FieldKind
}

// determineFieldFilterConditions determines the conditions and their corresponding operation
// for each indexed field.
// It returns a slice of fieldFilterCond, where each element corresponds to a field in the index.
func (f *IndexFetcher) determineFieldFilterConditions() ([]fieldFilterCond, error) {
	result := make([]fieldFilterCond, 0, len(f.indexedFields))
	for i := range f.indexedFields {
		fieldInd := f.mapping.FirstIndexOfName(f.indexedFields[i].Name)
		found := false
		// iterate through conditions and find the one that matches the current field
		for filterKey, indexFilterCond := range f.indexFilter.Conditions {
			propKey, ok := filterKey.(*mapper.PropertyIndex)
			if !ok || fieldInd != propKey.Index {
				continue
			}

			found = true

			fieldDef := f.indexedFields[slices.IndexFunc(f.indexedFields, func(f client.FieldDefinition) bool {
				return int(f.ID) == fieldInd
			})]

			condMap := indexFilterCond.(map[connor.FilterKey]any)

			jsonPath := []string{}
			if fieldDef.Kind == client.FieldKind_NILLABLE_JSON {

			jsonPathLoop:
				for {
					for key, filterVal := range condMap {
						prop, ok := key.(*mapper.ObjectProperty)
						if !ok {
							break jsonPathLoop
						}
						jsonPath = append(jsonPath, prop.Name)
						condMap = filterVal.(map[connor.FilterKey]any)
					}
				}
			}

			for key, filterVal := range condMap {
				cond := fieldFilterCond{
					op:       key.(*mapper.Operator).Operation,
					jsonPath: jsonPath,
					kind:     f.indexedFields[i].Kind,
				}

				var err error
				if len(jsonPath) > 0 {
					var jsonVal client.JSON
					if cond.op == compOpAny || cond.op == compOpAll || cond.op == compOpNone {
						subCondMap := filterVal.(map[connor.FilterKey]any)
						for subKey, subVal := range subCondMap {
							// TODO: check what happens with _any: {_eq: [1, 2]}
							cond.arrOp = cond.op
							cond.op = subKey.(*mapper.Operator).Operation
							jsonVal, err = client.NewJSONWithPath(subVal, jsonPath)
							// the sub condition is supposed to have only 1 record
							break
						}
					} else {
						jsonVal, err = client.NewJSONWithPath(filterVal, jsonPath)
					}
					if err == nil {
						cond.val = client.NewNormalJSON(jsonVal)
					}
				} else if filterVal == nil {
					cond.val, err = client.NewNormalNil(cond.kind)
				} else if !f.indexedFields[i].Kind.IsArray() {
					cond.val, err = client.NewNormalValue(filterVal)
				} else {
					subCondMap := filterVal.(map[connor.FilterKey]any)
					for subKey, subVal := range subCondMap {
						if subVal == nil {
							arrKind := cond.kind.(client.ScalarArrayKind)
							cond.val, err = client.NewNormalNil(arrKind.SubKind())
						} else {
							cond.val, err = client.NewNormalValue(subVal)
						}
						cond.arrOp = cond.op
						cond.op = subKey.(*mapper.Operator).Operation
						// the sub condition is supposed to have only 1 record
						break
					}
				}

				if err != nil {
					return nil, err
				}
				result = append(result, cond)
				break
			}
			break
		}
		if !found {
			result = append(result, fieldFilterCond{
				op:   opAny,
				val:  client.NormalVoid{},
				kind: f.indexedFields[i].Kind,
			})
		}
	}
	return result, nil
}

// isUniqueFetchByFullKey checks if the only index key can be fetched by the full index key.
//
// This method ignores the first condition (unless it's nil) because it's expected to be called only
// when the first field is used as a prefix in the index key. So we only check if the
// rest of the conditions are _eq.
func isUniqueFetchByFullKey(indexDesc *client.IndexDescription, conditions []fieldFilterCond) bool {
	// we need to check length of conditions because full key fetch is only possible
	// if all fields of the index are specified in the filter
	res := indexDesc.Unique && len(conditions) == len(indexDesc.Fields)

	// first condition is not required to be _eq, but if is, val must be not nil
	res = res && (conditions[0].op != opEq || !conditions[0].val.IsNil())

	// for the rest it must be _eq and val must be not nil
	for i := 1; i < len(conditions); i++ {
		res = res && (conditions[i].op == opEq && !conditions[i].val.IsNil())
	}
	return res
}
