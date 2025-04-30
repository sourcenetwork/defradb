// Copyright 2025 Democratized Data Foundation
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

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/connor"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/planner/filter"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
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
	opNot      = "_not"
	// it's just there for composite indexes. We construct a slice of value matchers with
	// every matcher being responsible for a corresponding field in the index to match.
	// For some fields there might not be any criteria to match. For examples if you have
	// composite index of /name/age/email/ and in the filter you specify only "name" and "email".
	// Then the "__any" matcher will be used for "age".
	opAny = "__any"
)

func isArrayCondition(op string) bool {
	return op == compOpAny || op == compOpAll || op == compOpNone
}

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
	resultIter    corekv.Iterator
	ctx           context.Context
	store         datastore.DSReaderWriter
	reverse       bool
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
		iterator, err := iter.store.Iterator(iter.ctx, corekv.IterOptions{
			Prefix:  iter.indexKey.Bytes(),
			Reverse: iter.reverse,
		})
		if err != nil {
			return err
		}
		iter.resultIter = iterator
	}
	return nil
}

func (iter *indexPrefixIterator) nextResult() (indexIterResult, error) {
	hasValue, err := iter.resultIter.Next()
	if err != nil || !hasValue {
		return indexIterResult{}, err
	}

	key, err := keys.DecodeIndexDataStoreKey(iter.resultIter.Key(), &iter.indexDesc, iter.indexedFields)
	if err != nil {
		return indexIterResult{}, err
	}

	value, err := iter.resultIter.Value()
	if err != nil {
		return indexIterResult{}, err
	}

	return indexIterResult{key: key, value: value, foundKey: true}, nil
}

func (iter *indexPrefixIterator) Next() (indexIterResult, error) {
	err := iter.checkResultIterator()
	if err != nil {
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

func (f *indexFetcher) newPrefixIterator(
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

func (f *indexPrefixIterator) Reverse(reverse bool) *indexPrefixIterator {
	f.reverse = reverse
	return f
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
	val, err := iter.store.Get(iter.ctx, iter.indexKey.Bytes())
	if err != nil {
		if errors.Is(err, corekv.ErrNotFound) {
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

// memorizingIndexIterator is an iterator for set of indexes that belong to the same document
// It keeps track of the already fetched documents to avoid duplicates.
type memorizingIndexIterator struct {
	inner indexIterator

	fetchedDocs map[string]struct{}

	ctx   context.Context
	store datastore.DSReaderWriter
}

var _ indexIterator = (*memorizingIndexIterator)(nil)

func (iter *memorizingIndexIterator) Init(ctx context.Context, store datastore.DSReaderWriter) error {
	iter.ctx = ctx
	iter.store = store
	iter.fetchedDocs = make(map[string]struct{})
	return iter.inner.Init(ctx, store)
}

func (iter *memorizingIndexIterator) Next() (indexIterResult, error) {
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

func (iter *memorizingIndexIterator) Close() error {
	return iter.inner.Close()
}

// newPrefixIteratorFromConditions creates a new indexPrefixIterator for fetching indexed data.
// It can modify the input matchers slice.
func (f *indexFetcher) newPrefixIteratorFromConditions(
	fieldConditions []fieldFilterCond,
	matchers []valueMatcher,
) (*indexPrefixIterator, error) {
	keyFieldValues := make([]client.NormalValue, 0, len(fieldConditions))
	for i := range fieldConditions {
		c := &fieldConditions[i]
		// prefix can be created only for subsequent _eq conditions. So we build the longest possible
		// prefix until we hit a condition that is not _eq.
		// The exception is when _eq is nested in _none.
		if c.op != opEq || c.arrOp == compOpNone {
			// if the field where we interrupt building of prefix is JSON, we still want to make sure
			// that the JSON path is included in the key
			if len(c.jsonPath) > 0 {
				jsonVal, _ := fieldConditions[i].val.JSON()
				keyFieldValues = append(keyFieldValues, client.NewNormalJSON(client.MakeVoidJSON(jsonVal.GetPath())))
			}
			break
		}

		keyFieldValues = append(keyFieldValues, fieldConditions[i].val)
	}

	// iterators for _eq filter already iterate over keys with first field value
	// matching the filter value, so we can skip the first matcher
	if len(matchers) > 1 {
		matchers[0] = &anyMatcher{}
	}

	key, err := f.newIndexDataStoreKeyWithValues(keyFieldValues)
	if err != nil {
		return nil, err
	}
	iter := f.newPrefixIterator(key, matchers, f.execInfo)
	ordered, reverse := f.isIndexOrdered()
	if ordered {
		iter.Reverse(reverse)
	}
	return iter, nil
}

// newInIndexIterator creates a new inIndexIterator for fetching indexed data.
// It can modify the input matchers slice.
func (f *indexFetcher) newInIndexIterator(
	fieldConditions []fieldFilterCond,
	matchers []valueMatcher,
) (*inIndexIterator, error) {
	inValues, err := client.ToArrayOfNormalValues(fieldConditions[0].val)
	if err != nil {
		return nil, NewErrInvalidInOperatorValue(err)
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

		key, err := f.newIndexDataStoreKeyWithValues(keyFieldValues)
		if err != nil {
			return nil, err
		}
		iter = &eqSingleIndexIterator{indexKey: key, execInfo: f.execInfo}
	} else {
		indexKey, err := f.newIndexDataStoreKey()
		if err != nil {
			return nil, err
		}
		indexKey.Fields = []keys.IndexedField{{Descending: f.indexDesc.Fields[0].Descending}}

		iter = f.newPrefixIterator(indexKey, matchers, f.execInfo)
	}
	return &inIndexIterator{indexIterator: iter, inValues: inValues}, nil
}

func (f *indexFetcher) newIndexDataStoreKey() (keys.IndexDataStoreKey, error) {
	shortID, err := id.GetShortCollectionID(f.ctx, f.txn, f.col.Description().CollectionID)
	if err != nil {
		return keys.IndexDataStoreKey{}, err
	}

	return keys.IndexDataStoreKey{CollectionShortID: shortID, IndexID: f.indexDesc.ID}, nil
}

func (f *indexFetcher) newIndexDataStoreKeyWithValues(values []client.NormalValue) (keys.IndexDataStoreKey, error) {
	fields := make([]keys.IndexedField, len(values))
	for i := range values {
		fields[i].Value = values[i]
		fields[i].Descending = f.indexDesc.Fields[i].Descending
	}

	shortID, err := id.GetShortCollectionID(f.ctx, f.txn, f.col.Description().CollectionID)
	if err != nil {
		return keys.IndexDataStoreKey{}, err
	}

	return keys.NewIndexDataStoreKey(shortID, f.indexDesc.ID, fields), nil
}

func (f *indexFetcher) tryCreateOrderedIndexIterator() (indexIterator, error) {
	ordered, reverse := f.isIndexOrdered()
	if ordered {
		key, err := f.newIndexDataStoreKey()
		if err != nil {
			return nil, err
		}
		iter := f.newPrefixIterator(key, nil, f.execInfo).Reverse(reverse)
		return iter, nil
	}
	return nil, nil
}

func (f *indexFetcher) isIndexOrdered() (bool, bool) {
	if len(f.ordering) > 0 {
		orderField, found := f.mapping.TryToFindNameFromIndex(f.ordering[0].FieldIndexes[0])
		if found {
			indexField := f.indexDesc.Fields[0]
			if indexField.Name == orderField {
				return true, f.ordering[0].Direction == mapper.DESC && !indexField.Descending
			}
		}
	}
	return false, false
}

func (f *indexFetcher) createIndexIterator() (indexIterator, error) {
	fieldConditions, err := f.determineFieldFilterConditions()
	if err != nil {
		return nil, err
	}

	// fieldConditions might be empty if a query contains an empty condition like User(filter: {name: {}})
	if len(fieldConditions) == 0 {
		return f.tryCreateOrderedIndexIterator()
	}

	matchers, err := createValueMatchers(fieldConditions)
	if err != nil {
		return nil, err
	}

	var iter indexIterator
	if fieldConditions[0].op == opEq {
		if isUniqueFetchByFullKey(&f.indexDesc, fieldConditions) {
			keyFieldValues := make([]client.NormalValue, len(fieldConditions))
			for i := range fieldConditions {
				keyFieldValues[i] = fieldConditions[i].val
			}

			key, err := f.newIndexDataStoreKeyWithValues(keyFieldValues)
			if err != nil {
				return nil, err
			}
			iter = &eqSingleIndexIterator{indexKey: key, execInfo: f.execInfo}
		} else {
			iter, err = f.newPrefixIteratorFromConditions(fieldConditions, matchers)
			if err != nil {
				return nil, err
			}
		}
	} else if fieldConditions[0].op == opIn && fieldConditions[0].arrOp != compOpNone {
		iter, err = f.newInIndexIterator(fieldConditions, matchers)
		if err != nil {
			return nil, err
		}
	} else {
		iter, err = f.newPrefixIteratorFromConditions(fieldConditions, matchers)
		if err != nil {
			return nil, err
		}
	}

	if err != nil {
		return nil, err
	}

	if iter == nil {
		return nil, NewErrInvalidFilterOperator(fieldConditions[0].op)
	}

	if doConditionsHaveArrayOrJSON(fieldConditions) {
		iter = &memorizingIndexIterator{inner: iter}
	}

	return iter, nil
}

func doConditionsHaveArrayOrJSON(conditions []fieldFilterCond) bool {
	hasArray := false
	hasJSON := false
	for i := range conditions {
		hasJSON = hasJSON || conditions[i].kind == client.FieldKind_NILLABLE_JSON
		hasArray = hasArray || conditions[i].kind.IsArray()
	}
	return hasArray || hasJSON
}

type fieldFilterCond struct {
	op       string
	arrOp    string
	jsonPath client.JSONPath
	val      client.NormalValue
	kind     client.FieldKind
}

// determineFieldFilterConditions determines the conditions and their corresponding operation
// for each indexed field.
// It returns a slice of fieldFilterCond, where each element corresponds to a field in the index.
func (f *indexFetcher) determineFieldFilterConditions() ([]fieldFilterCond, error) {
	if f.indexFilter == nil {
		return nil, nil
	}

	result := make([]fieldFilterCond, 0, len(f.indexedFields))
	// we process first the conditions that match composite index fields starting from the first one
	for i := range f.indexDesc.Fields {
		indexedField := f.indexedFields[i]
		fieldInd := f.mapping.FirstIndexOfName(indexedField.Name)
		var err error

		filter.TraverseProperties(
			f.indexFilter.Conditions,
			func(prop *mapper.PropertyIndex, condMap map[connor.FilterKey]any) bool {
				if fieldInd != prop.Index {
					return true
				}

				var jsonPath client.JSONPath
				condMap, jsonPath = getNestedOperatorConditionIfJSON(indexedField, condMap)

				for key, filterVal := range condMap {
					op := key.(*mapper.Operator).Operation

					// if the array condition is _none it doesn't make sense to use index  because
					// values picked by the index is random guessing. For example if we have doc1
					// with array of [3, 5, 1] and doc2 with [7, 4, 8] the index first fetches
					// value 1 of doc1, let it go through the filter and then fetches value 3 of doc1
					// again, skips it (because it cached doc1 id) and fetches value 4 of doc2, and
					// so on until it exhaust all prefixes in ascending order.
					// It might be even less effective than just scanning all documents.
					if op == compOpNone {
						return true
					}

					cond, err := makeFieldFilterCondition(op, jsonPath, indexedField, filterVal)

					if err != nil {
						return false
					}

					result = append(result, cond)
					break
				}
				return false
			},
			// if the filter contains _not operator, we ignore the entire branch because in this
			// case index will do more harm. For example if we have _not: {_eq: 5} and the index
			// fetches value 5, it will skip all documents with value 5, but we need to return them.
			opNot,
		)

		// if after traversing the filter for the first field we didn't find any condition that can
		// be used with the index, we return nil indicating that the index can't be used.
		if len(result) == 0 {
			return nil, err
		}

		// if traversing for the current (not first) field of the composite index didn't find any
		// condition, we add a dummy that will match any value for this field.
		if len(result) == i {
			result = append(result, fieldFilterCond{
				op:   opAny,
				val:  client.NormalVoid{},
				kind: indexedField.Kind,
			})
		}
	}
	return result, nil
}

// makeFieldFilterCondition creates a fieldFilterCond based on the given operator and filter value on
// the given indexed field.
// If jsonPath is not empty, it means that the indexed field is a JSON field and the filter value
// should be treated as a JSON value.
func makeFieldFilterCondition(
	op string,
	jsonPath client.JSONPath,
	indexedField client.FieldDefinition,
	filterVal any,
) (fieldFilterCond, error) {
	cond := fieldFilterCond{
		op:       op,
		jsonPath: jsonPath,
		kind:     indexedField.Kind,
	}

	var err error
	if len(jsonPath) > 0 {
		err = setJSONFilterCondition(&cond, filterVal, jsonPath)
	} else if filterVal == nil {
		cond.val, err = client.NewNormalNil(cond.kind)
	} else if !indexedField.Kind.IsArray() {
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
	return cond, err
}

// getNestedOperatorConditionIfJSON traverses the filter map if the indexed field is JSON to find the
// nested operator condition and returns it along with the JSON path to the nested field.
// If the indexed field is not JSON, it returns the original condition map.
func getNestedOperatorConditionIfJSON(
	indexedField client.FieldDefinition,
	condMap map[connor.FilterKey]any,
) (map[connor.FilterKey]any, client.JSONPath) {
	if indexedField.Kind != client.FieldKind_NILLABLE_JSON {
		return condMap, client.JSONPath{}
	}
	var jsonPath client.JSONPath
	for {
		for key, filterVal := range condMap {
			prop, ok := key.(*mapper.ObjectProperty)
			if !ok {
				// if filter contains an array condition, we need to append index 0 to the json path
				// to limit the search only to array elements
				op, ok := key.(*mapper.Operator)
				if ok && isArrayCondition(op.Operation) {
					jsonPath = jsonPath.AppendIndex(0)
				}
				return condMap, jsonPath
			}
			jsonPath = jsonPath.AppendProperty(prop.Name)
			// if key is ObjectProperty it's safe to cast filterVal to map[connor.FilterKey]any
			// containing either another nested ObjectProperty or Operator
			condMap = filterVal.(map[connor.FilterKey]any)
		}
	}
}

// setJSONFilterCondition sets up the given condition struct based on the filter value and JSON path so that
// it can be used to fetch the indexed data.
func setJSONFilterCondition(cond *fieldFilterCond, filterVal any, jsonPath client.JSONPath) error {
	if isArrayCondition(cond.op) {
		subCondMap := filterVal.(map[connor.FilterKey]any)
		for subKey, subVal := range subCondMap {
			cond.arrOp = cond.op
			cond.op = subKey.(*mapper.Operator).Operation
			jsonVal, err := client.NewJSONWithPath(subVal, jsonPath)
			if err != nil {
				return err
			}
			cond.val = client.NewNormalJSON(jsonVal)
			// the array sub condition (_any, _all or _none) is supposed to have only 1 record
			break
		}
	} else if cond.op == opIn {
		// values in _in operator should not be considered as array elements just because they happened
		// to be written as an array in the filter. We need to convert them to normal values and
		// treat them individually.
		var jsonVals []client.JSON
		if anyArr, ok := filterVal.([]any); ok {
			// if filter value is []any we convert each value separately because JSON might have
			// array elements of different types. That's why we can't just pass it directly to
			// client.ToArrayOfNormalValues
			jsonVals = make([]client.JSON, 0, len(anyArr))
			for _, val := range anyArr {
				jsonVal, err := client.NewJSONWithPath(val, jsonPath)
				if err != nil {
					return err
				}
				jsonVals = append(jsonVals, jsonVal)
			}
		} else {
			normValue, err := client.NewNormalValue(filterVal)
			if err != nil {
				return err
			}
			normArr, err := client.ToArrayOfNormalValues(normValue)
			if err != nil {
				return err
			}
			jsonVals = make([]client.JSON, 0, len(normArr))
			for _, val := range normArr {
				jsonVal, err := client.NewJSONWithPath(val.Unwrap(), jsonPath)
				if err != nil {
					return err
				}
				jsonVals = append(jsonVals, jsonVal)
			}
		}
		normJSONs, err := client.NewNormalValue(jsonVals)
		if err != nil {
			return err
		}
		cond.val = normJSONs
	} else {
		jsonVal, err := client.NewJSONWithPath(filterVal, jsonPath)
		if err != nil {
			return err
		}
		cond.val = client.NewNormalJSON(jsonVal)
	}
	return nil
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
