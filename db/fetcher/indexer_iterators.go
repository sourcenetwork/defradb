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
	"cmp"
	"context"
	"errors"
	"strings"

	ds "github.com/ipfs/go-datastore"

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
	// it's just there for composite indexes. We construct a slice of value matchers with
	// every matcher being responsible for a corresponding field in the index to match.
	// For some fields there might not be any criteria to match. For examples if you have
	// composite index of /name/age/email/ and in the filter you specify only "name" and "email".
	// Then the "_any" matcher will be used for "age".
	opAny = "_any"
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
	key      core.IndexDataStoreKey
	foundKey bool
	value    []byte
}

type queryResultIterator struct {
	resultIter    query.Results
	indexDesc     client.IndexDescription
	indexedFields []client.FieldDescription
}

func (iter *queryResultIterator) Next() (indexIterResult, error) {
	res, hasVal := iter.resultIter.NextSync()
	if res.Error != nil {
		return indexIterResult{}, res.Error
	}
	if !hasVal {
		return indexIterResult{}, nil
	}
	key, err := core.DecodeIndexDataStoreKey([]byte(res.Key), &iter.indexDesc, iter.indexedFields)
	if err != nil {
		return indexIterResult{}, err
	}

	return indexIterResult{key: key, value: res.Value, foundKey: true}, nil
}

func (iter *queryResultIterator) Close() error {
	return iter.resultIter.Close()
}

type eqPrefixIndexIterator struct {
	queryResultIterator
	indexKey core.IndexDataStoreKey
	execInfo *ExecInfo
	matchers []valueMatcher
}

func (iter *eqPrefixIndexIterator) Init(ctx context.Context, store datastore.DSReaderWriter) error {
	resultIter, err := store.Query(ctx, query.Query{
		Prefix: iter.indexKey.ToString(),
	})
	if err != nil {
		return err
	}
	iter.resultIter = resultIter
	return nil
}

func (iter *eqPrefixIndexIterator) Next() (indexIterResult, error) {
	for {
		res, err := iter.queryResultIterator.Next()
		if err != nil || !res.foundKey {
			return res, err
		}
		iter.execInfo.IndexesFetched++
		doesMatch, err := executeValueMatchers(iter.matchers, res.key.Fields())
		if err != nil {
			return indexIterResult{}, err
		}
		if !doesMatch {
			continue
		}
		return res, err
	}
}

type eqSingleIndexIterator struct {
	indexKey core.IndexDataStoreKey
	execInfo *ExecInfo

	ctx   context.Context
	store datastore.DSReaderWriter
}

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

func (i *eqSingleIndexIterator) Close() error {
	return nil
}

type inIndexIterator struct {
	indexIterator
	inValues     []any
	nextValIndex int
	ctx          context.Context
	store        datastore.DSReaderWriter
	hasIterator  bool
}

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

	var err error
	switch fieldIter := iter.indexIterator.(type) {
	case *eqPrefixIndexIterator:
		err = fieldIter.indexKey.SetField(0, core.IndexedField{
			Value:      iter.inValues[iter.nextValIndex],
			Descending: fieldIter.indexKey.Field(0).Descending,
		})
	case *eqSingleIndexIterator:
		err = fieldIter.indexKey.SetField(0, core.IndexedField{
			Value:      iter.inValues[iter.nextValIndex],
			Descending: fieldIter.indexKey.Field(0).Descending,
		})
	}
	if err != nil {
		return false, err
	}
	err = iter.indexIterator.Init(iter.ctx, iter.store)
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

func executeValueMatchers(matchers []valueMatcher, fields []core.IndexedField) (bool, error) {
	for i := range matchers {
		res, err := matchers[i].Match(fields[i].Value)
		if err != nil {
			return false, err
		}
		if !res {
			return false, nil
		}
	}
	return true, nil
}

type scanningIndexIterator struct {
	queryResultIterator
	indexKey core.IndexDataStoreKey
	matchers []valueMatcher
	execInfo *ExecInfo
}

func (iter *scanningIndexIterator) Init(ctx context.Context, store datastore.DSReaderWriter) error {
	resultIter, err := store.Query(ctx, query.Query{
		Prefix: iter.indexKey.ToString(),
	})
	if err != nil {
		return err
	}
	iter.resultIter = resultIter

	return nil
}

func (iter *scanningIndexIterator) Next() (indexIterResult, error) {
	for {
		res, err := iter.queryResultIterator.Next()
		if err != nil || !res.foundKey {
			return indexIterResult{}, err
		}
		iter.execInfo.IndexesFetched++

		didMatch, err := executeValueMatchers(iter.matchers, res.key.Fields())

		if didMatch {
			return res, err
		}
	}
}

// checks if the value satisfies the condition
type valueMatcher interface {
	Match(any) (bool, error)
}

type intMatcher struct {
	value    int64
	evalFunc func(int64, int64) bool
}

func (m *intMatcher) Match(value any) (bool, error) {
	intVal, ok := value.(int64)
	if !ok {
		return false, NewErrUnexpectedTypeValue[int64](value)
	}
	return m.evalFunc(intVal, m.value), nil
}

type floatMatcher struct {
	value    float64
	evalFunc func(float64, float64) bool
}

func (m *floatMatcher) Match(value any) (bool, error) {
	floatVal, ok := value.(float64)
	if !ok {
		return false, NewErrUnexpectedTypeValue[float64](value)
	}
	return m.evalFunc(m.value, floatVal), nil
}

type stringMatcher struct {
	value    string
	evalFunc func(string, string) bool
}

func (m *stringMatcher) Match(value any) (bool, error) {
	stringVal, ok := value.(string)
	if !ok {
		return false, NewErrUnexpectedTypeValue[string](value)
	}
	return m.evalFunc(m.value, stringVal), nil
}

type nilMatcher struct{}

func (m *nilMatcher) Match(value any) (bool, error) {
	return value == nil, nil
}

// checks if the index value is or is not in the given array
type indexInArrayMatcher struct {
	inValues []any
	isIn     bool
}

func newNinIndexCmp(values []any, kind client.FieldKind, isIn bool) (*indexInArrayMatcher, error) {
	normalizeValueFunc := getNormalizeValueFunc(kind)
	for i := range values {
		normalized, err := normalizeValueFunc(values[i])
		if err != nil {
			return nil, err
		}
		values[i] = normalized
	}
	return &indexInArrayMatcher{inValues: values, isIn: isIn}, nil
}

func getNormalizeValueFunc(kind client.FieldKind) func(any) (any, error) {
	switch kind {
	case client.FieldKind_NILLABLE_INT:
		return func(value any) (any, error) {
			if v, ok := value.(int64); ok {
				return v, nil
			}
			if v, ok := value.(int32); ok {
				return int64(v), nil
			}
			return nil, ErrInvalidInOperatorValue
		}
	case client.FieldKind_NILLABLE_FLOAT:
		return func(value any) (any, error) {
			if v, ok := value.(float64); ok {
				return v, nil
			}
			if v, ok := value.(float32); ok {
				return float64(v), nil
			}
			return nil, ErrInvalidInOperatorValue
		}
	case client.FieldKind_NILLABLE_STRING:
		return func(value any) (any, error) {
			if v, ok := value.(string); ok {
				return v, nil
			}
			return nil, ErrInvalidInOperatorValue
		}
	}
	return nil
}

func (m *indexInArrayMatcher) Match(value any) (bool, error) {
	for _, inVal := range m.inValues {
		if inVal == value {
			return m.isIn, nil
		}
	}
	return !m.isIn, nil
}

// checks if the index value satisfies the LIKE condition
type indexLikeMatcher struct {
	hasPrefix   bool
	hasSuffix   bool
	startAndEnd []string
	isLike      bool
	value       string
}

func newLikeIndexCmp(filterValue string, isLike bool) (*indexLikeMatcher, error) {
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

	return matcher, nil
}

func (m *indexLikeMatcher) Match(value any) (bool, error) {
	currentVal, ok := value.(string)
	if !ok {
		return false, NewErrUnexpectedTypeValue[string](currentVal)
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

type anyMatcher struct{}

func (m *anyMatcher) Match(any) (bool, error) { return true, nil }

// newPrefixIndexIterator creates a new eqPrefixIndexIterator for fetching indexed data.
// It can modify the input matchers slice.
func (f *IndexFetcher) newPrefixIndexIterator(
	fieldConditions []fieldFilterCond,
	matchers []valueMatcher,
) (*eqPrefixIndexIterator, error) {
	keyFieldValues := make([]any, 0, len(fieldConditions))
	for i := range fieldConditions {
		if fieldConditions[i].op != opEq {
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

	key, err := f.newIndexDataStoreKeyWithValues(keyFieldValues)
	if err != nil {
		return nil, err
	}

	return &eqPrefixIndexIterator{
		queryResultIterator: f.newQueryResultIterator(),
		indexKey:            key,
		execInfo:            &f.execInfo,
		matchers:            matchers,
	}, nil
}

func (f *IndexFetcher) newQueryResultIterator() queryResultIterator {
	return queryResultIterator{indexDesc: f.indexDesc, indexedFields: f.indexedFields}
}

// newInIndexIterator creates a new inIndexIterator for fetching indexed data.
// It can modify the input matchers slice.
func (f *IndexFetcher) newInIndexIterator(
	fieldConditions []fieldFilterCond,
	matchers []valueMatcher,
) (*inIndexIterator, error) {
	inArr, ok := fieldConditions[0].val.([]any)
	if !ok {
		return nil, ErrInvalidInOperatorValue
	}
	inValues := make([]any, 0, len(inArr))
	for _, v := range inArr {
		fieldVal := client.NewFieldValue(client.NONE_CRDT, v, f.indexedFields[0].Kind)
		inValues = append(inValues, fieldVal.Value())
	}

	// iterators for _in filter already iterate over keys with first field value
	// matching the filter value, so we can skip the first matcher
	if len(matchers) > 1 {
		matchers[0] = &anyMatcher{}
	}

	var iter indexIterator
	if isUniqueFetchByFullKey(&f.indexDesc, fieldConditions) {
		keyFieldValues := make([]any, len(fieldConditions))
		for i := range fieldConditions {
			keyFieldValues[i] = fieldConditions[i].val
		}

		key, err := f.newIndexDataStoreKeyWithValues(keyFieldValues)
		if err != nil {
			return nil, err
		}

		iter = &eqSingleIndexIterator{
			indexKey: key,
			execInfo: &f.execInfo,
		}
	} else {
		indexKey := f.newIndexDataStoreKey()
		indexKey.SetFields([]core.IndexedField{{Descending: f.indexDesc.Fields[0].Descending}})

		iter = &eqPrefixIndexIterator{
			queryResultIterator: f.newQueryResultIterator(),
			indexKey:            indexKey,
			execInfo:            &f.execInfo,
			matchers:            matchers,
		}
	}
	return &inIndexIterator{
		indexIterator: iter,
		inValues:      inValues,
	}, nil
}

func (f *IndexFetcher) newIndexDataStoreKey() core.IndexDataStoreKey {
	key := core.IndexDataStoreKey{CollectionID: f.col.ID(), IndexID: f.indexDesc.ID}
	return key
}

func (f *IndexFetcher) newIndexDataStoreKeyWithValues(values []any) (core.IndexDataStoreKey, error) {
	fields := make([]core.IndexedField, len(values))
	for i := range values {
		fields[i].Value = values[i]
		fields[i].Descending = f.indexDesc.Fields[i].Descending
	}
	key, err := core.NewIndexDataStoreKey(f.col.ID(), f.indexDesc.ID, fields)
	return key, err
}

func (f *IndexFetcher) createIndexIterator() (indexIterator, error) {
	fieldConditions := f.determineFieldFilterConditions()

	matchers, err := createValueMatchers(fieldConditions)
	if err != nil {
		return nil, err
	}

	switch fieldConditions[0].op {
	case opEq:
		if isUniqueFetchByFullKey(&f.indexDesc, fieldConditions) {
			keyFieldValues := make([]any, len(fieldConditions))
			for i := range fieldConditions {
				keyFieldValues[i] = fieldConditions[i].val
			}

			key, err := f.newIndexDataStoreKeyWithValues(keyFieldValues)
			if err != nil {
				return nil, err
			}

			return &eqSingleIndexIterator{
				indexKey: key,
				execInfo: &f.execInfo,
			}, nil
		} else {
			return f.newPrefixIndexIterator(fieldConditions, matchers)
		}
	case opIn:
		return f.newInIndexIterator(fieldConditions, matchers)
	case opGt, opGe, opLt, opLe, opNe, opNin, opLike, opNlike:
		return &scanningIndexIterator{
			queryResultIterator: f.newQueryResultIterator(),
			indexKey:            f.newIndexDataStoreKey(),
			matchers:            matchers,
			execInfo:            &f.execInfo,
		}, nil
	}

	return nil, NewErrInvalidFilterOperator(fieldConditions[0].op)
}

func createValueMatcher(condition *fieldFilterCond) (valueMatcher, error) {
	if condition.op == "" {
		return &anyMatcher{}, nil
	}

	if client.IsNillableKind(condition.kind) && condition.val == nil {
		return &nilMatcher{}, nil
	}

	switch condition.op {
	case opEq, opGt, opGe, opLt, opLe, opNe:
		switch condition.kind {
		case client.FieldKind_NILLABLE_INT:
			var intVal int64
			switch v := condition.val.(type) {
			case int64:
				intVal = v
			case int32:
				intVal = int64(v)
			case int:
				intVal = int64(v)
			default:
				return nil, NewErrUnexpectedTypeValue[int64](condition.val)
			}
			return &intMatcher{value: intVal, evalFunc: getCompareValsFunc[int64](condition.op)}, nil
		case client.FieldKind_NILLABLE_FLOAT:
			floatVal, ok := condition.val.(float64)
			if !ok {
				return nil, NewErrUnexpectedTypeValue[float64](condition.val)
			}
			return &floatMatcher{value: floatVal, evalFunc: getCompareValsFunc[float64](condition.op)}, nil
		case client.FieldKind_DocID, client.FieldKind_NILLABLE_STRING:
			strVal, ok := condition.val.(string)
			if !ok {
				return nil, NewErrUnexpectedTypeValue[string](condition.val)
			}
			return &stringMatcher{value: strVal, evalFunc: getCompareValsFunc[string](condition.op)}, nil
		}
	case opIn, opNin:
		inArr, ok := condition.val.([]any)
		if !ok {
			return nil, ErrInvalidInOperatorValue
		}
		return newNinIndexCmp(inArr, condition.kind, condition.op == opIn)
	case opLike, opNlike:
		strVal, ok := condition.val.(string)
		if !ok {
			return nil, NewErrUnexpectedTypeValue[string](condition.val)
		}
		return newLikeIndexCmp(strVal, condition.op == opLike)
	case opAny:
		return &anyMatcher{}, nil
	}

	return nil, NewErrInvalidFilterOperator(condition.op)
}

func createValueMatchers(conditions []fieldFilterCond) ([]valueMatcher, error) {
	matchers := make([]valueMatcher, 0, len(conditions))
	for i := range conditions {
		m, err := createValueMatcher(&conditions[i])
		if err != nil {
			return nil, err
		}
		matchers = append(matchers, m)
	}
	return matchers, nil
}

type fieldFilterCond struct {
	op   string
	val  any
	kind client.FieldKind
}

// determineFieldFilterConditions determines the conditions and their corresponding operation
// for each indexed field.
// It returns a slice of fieldFilterCond, where each element corresponds to a field in the index.
func (f *IndexFetcher) determineFieldFilterConditions() []fieldFilterCond {
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

			condMap := indexFilterCond.(map[connor.FilterKey]any)
			for key, filterVal := range condMap {
				opKey := key.(*mapper.Operator)
				result = append(result, fieldFilterCond{
					op:   opKey.Operation,
					val:  filterVal,
					kind: f.indexedFields[i].Kind,
				})
				break
			}
			break
		}
		if !found {
			result = append(result, fieldFilterCond{op: opAny})
		}
	}
	return result
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
	res = res && (conditions[0].op != opEq || conditions[0].val != nil)

	// for the rest it must be _eq and val must be not nil
	for i := 1; i < len(conditions); i++ {
		res = res && (conditions[i].op == opEq && conditions[i].val != nil)
	}
	return res
}

func getCompareValsFunc[T cmp.Ordered](op string) func(T, T) bool {
	switch op {
	case opGt:
		return checkGT
	case opGe:
		return checkGE
	case opLt:
		return checkLT
	case opLe:
		return checkLE
	case opEq:
		return checkEQ
	case opNe:
		return checkNE
	}
	return nil
}

func checkGE[T cmp.Ordered](a, b T) bool { return a >= b }
func checkGT[T cmp.Ordered](a, b T) bool { return a > b }
func checkLE[T cmp.Ordered](a, b T) bool { return a <= b }
func checkLT[T cmp.Ordered](a, b T) bool { return a < b }
func checkEQ[T cmp.Ordered](a, b T) bool { return a == b }
func checkNE[T cmp.Ordered](a, b T) bool { return a != b }
