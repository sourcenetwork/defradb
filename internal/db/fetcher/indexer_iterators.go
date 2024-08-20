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
	"time"

	ds "github.com/ipfs/go-datastore"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/internal/connor"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"

	"github.com/ipfs/go-datastore/query"
)

const (
	opEq     = "_eq"
	opGt     = "_gt"
	opGe     = "_ge"
	opLt     = "_lt"
	opLe     = "_le"
	opNe     = "_ne"
	opIn     = "_in"
	opNin    = "_nin"
	opLike   = "_like"
	opNlike  = "_nlike"
	opILike  = "_ilike"
	opNILike = "_nilike"
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
	indexedFields []client.FieldDefinition
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
		doesMatch, err := executeValueMatchers(iter.matchers, res.key.Fields)
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
	inValues     []client.NormalValue
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

	switch fieldIter := iter.indexIterator.(type) {
	case *eqPrefixIndexIterator:
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

		didMatch, err := executeValueMatchers(iter.matchers, res.key.Fields)

		if didMatch {
			return res, err
		}
	}
}

// checks if the value satisfies the condition
type valueMatcher interface {
	Match(client.NormalValue) (bool, error)
}

type intMatcher struct {
	value    int64
	evalFunc func(int64, int64) bool
}

func (m *intMatcher) Match(value client.NormalValue) (bool, error) {
	if intVal, ok := value.Int(); ok {
		return m.evalFunc(intVal, m.value), nil
	}
	if intOptVal, ok := value.NillableInt(); ok {
		if !intOptVal.HasValue() {
			return false, nil
		}
		return m.evalFunc(intOptVal.Value(), m.value), nil
	}
	return false, NewErrUnexpectedTypeValue[int64](value)
}

type floatMatcher struct {
	value    float64
	evalFunc func(float64, float64) bool
}

func (m *floatMatcher) Match(value client.NormalValue) (bool, error) {
	if floatVal, ok := value.Float(); ok {
		return m.evalFunc(floatVal, m.value), nil
	}
	if floatOptVal, ok := value.NillableFloat(); ok {
		if !floatOptVal.HasValue() {
			return false, nil
		}
		return m.evalFunc(floatOptVal.Value(), m.value), nil
	}
	return false, NewErrUnexpectedTypeValue[float64](value)
}

type stringMatcher struct {
	value    string
	evalFunc func(string, string) bool
}

func (m *stringMatcher) Match(value client.NormalValue) (bool, error) {
	if strVal, ok := value.String(); ok {
		return m.evalFunc(strVal, m.value), nil
	}
	if strOptVal, ok := value.NillableString(); ok {
		if !strOptVal.HasValue() {
			return false, nil
		}
		return m.evalFunc(strOptVal.Value(), m.value), nil
	}
	return false, NewErrUnexpectedTypeValue[string](value)
}

type timeMatcher struct {
	op    string
	value time.Time
}

func (m *timeMatcher) Match(value client.NormalValue) (bool, error) {
	timeVal, ok := value.Time()
	if !ok {
		if timeOptVal, ok := value.NillableTime(); ok {
			timeVal = timeOptVal.Value()
		} else {
			return false, NewErrUnexpectedTypeValue[time.Time](value)
		}
	}
	switch m.op {
	case opEq:
		return timeVal.Equal(m.value), nil
	case opGt:
		return timeVal.After(m.value), nil
	case opGe:
		return !timeVal.Before(m.value), nil
	case opLt:
		return timeVal.Before(m.value), nil
	case opLe:
		return !timeVal.After(m.value), nil
	case opNe:
		return !timeVal.Equal(m.value), nil
	}
	return false, NewErrInvalidFilterOperator(m.op)
}

type nilMatcher struct {
	matchNil bool
}

func (m *nilMatcher) Match(value client.NormalValue) (bool, error) {
	return value.IsNil() == m.matchNil, nil
}

// checks if the index value is or is not in the given array
type indexInArrayMatcher struct {
	inValues []client.NormalValue
	isIn     bool
}

func (m *indexInArrayMatcher) Match(value client.NormalValue) (bool, error) {
	for _, inVal := range m.inValues {
		if inVal.Unwrap() == value.Unwrap() {
			return m.isIn, nil
		}
	}
	return !m.isIn, nil
}

// checks if the index value satisfies the LIKE condition
type indexLikeMatcher struct {
	hasPrefix         bool
	hasSuffix         bool
	startAndEnd       []string
	isLike            bool
	isCaseInsensitive bool
	value             string
}

func newLikeIndexCmp(filterValue string, isLike bool, isCaseInsensitive bool) (*indexLikeMatcher, error) {
	matcher := &indexLikeMatcher{
		isLike:            isLike,
		isCaseInsensitive: isCaseInsensitive,
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
	if isCaseInsensitive {
		matcher.value = strings.ToLower(filterValue)
	} else {
		matcher.value = filterValue
	}

	return matcher, nil
}

func (m *indexLikeMatcher) Match(value client.NormalValue) (bool, error) {
	strVal, ok := value.String()
	if !ok {
		strOptVal, ok := value.NillableString()
		if !ok {
			return false, NewErrUnexpectedTypeValue[string](value)
		}
		if !strOptVal.HasValue() {
			return false, nil
		}
		strVal = strOptVal.Value()
	}
	if m.isCaseInsensitive {
		strVal = strings.ToLower(strVal)
	}

	return m.doesMatch(strVal) == m.isLike, nil
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

func (m *anyMatcher) Match(client.NormalValue) (bool, error) { return true, nil }

// newPrefixIndexIterator creates a new eqPrefixIndexIterator for fetching indexed data.
// It can modify the input matchers slice.
func (f *IndexFetcher) newPrefixIndexIterator(
	fieldConditions []fieldFilterCond,
	matchers []valueMatcher,
) (*eqPrefixIndexIterator, error) {
	keyFieldValues := make([]client.NormalValue, 0, len(fieldConditions))
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

	key := f.newIndexDataStoreKeyWithValues(keyFieldValues)

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
		indexKey.Fields = []core.IndexedField{{Descending: f.indexDesc.Fields[0].Descending}}

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

func (f *IndexFetcher) newIndexDataStoreKeyWithValues(values []client.NormalValue) core.IndexDataStoreKey {
	fields := make([]core.IndexedField, len(values))
	for i := range values {
		fields[i].Value = values[i]
		fields[i].Descending = f.indexDesc.Fields[i].Descending
	}
	return core.NewIndexDataStoreKey(f.col.ID(), f.indexDesc.ID, fields)
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

	switch fieldConditions[0].op {
	case opEq:
		if isUniqueFetchByFullKey(&f.indexDesc, fieldConditions) {
			keyFieldValues := make([]client.NormalValue, len(fieldConditions))
			for i := range fieldConditions {
				keyFieldValues[i] = fieldConditions[i].val
			}

			key := f.newIndexDataStoreKeyWithValues(keyFieldValues)

			return &eqSingleIndexIterator{
				indexKey: key,
				execInfo: &f.execInfo,
			}, nil
		} else {
			return f.newPrefixIndexIterator(fieldConditions, matchers)
		}
	case opIn:
		return f.newInIndexIterator(fieldConditions, matchers)
	case opGt, opGe, opLt, opLe, opNe, opNin, opLike, opNlike, opILike, opNILike:
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

	if condition.val.IsNil() {
		return &nilMatcher{matchNil: condition.op == opEq}, nil
	}

	switch condition.op {
	case opEq, opGt, opGe, opLt, opLe, opNe:
		if v, ok := condition.val.Int(); ok {
			return &intMatcher{value: v, evalFunc: getCompareValsFunc[int64](condition.op)}, nil
		}
		if v, ok := condition.val.NillableInt(); ok {
			return &intMatcher{value: v.Value(), evalFunc: getCompareValsFunc[int64](condition.op)}, nil
		}
		if v, ok := condition.val.Float(); ok {
			return &floatMatcher{value: v, evalFunc: getCompareValsFunc[float64](condition.op)}, nil
		}
		if v, ok := condition.val.NillableFloat(); ok {
			return &floatMatcher{value: v.Value(), evalFunc: getCompareValsFunc[float64](condition.op)}, nil
		}
		if v, ok := condition.val.String(); ok {
			return &stringMatcher{value: v, evalFunc: getCompareValsFunc[string](condition.op)}, nil
		}
		if v, ok := condition.val.NillableString(); ok {
			return &stringMatcher{value: v.Value(), evalFunc: getCompareValsFunc[string](condition.op)}, nil
		}
		if v, ok := condition.val.Time(); ok {
			return &timeMatcher{value: v, op: condition.op}, nil
		}
		if v, ok := condition.val.NillableTime(); ok {
			return &timeMatcher{value: v.Value(), op: condition.op}, nil
		}
	case opIn, opNin:
		inVals, err := client.ToArrayOfNormalValues(condition.val)
		if err != nil {
			return nil, err
		}
		return &indexInArrayMatcher{inValues: inVals, isIn: condition.op == opIn}, nil
	case opLike, opNlike, opILike, opNILike:
		strVal, ok := condition.val.String()
		if !ok {
			strOptVal, ok := condition.val.NillableString()
			if !ok {
				return nil, NewErrUnexpectedTypeValue[string](condition.val)
			}
			strVal = strOptVal.Value()
		}
		isLike := condition.op == opLike || condition.op == opILike
		isCaseInsensitive := condition.op == opILike || condition.op == opNILike
		return newLikeIndexCmp(strVal, isLike, isCaseInsensitive)
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
	val  client.NormalValue
	kind client.FieldKind
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

			condMap := indexFilterCond.(map[connor.FilterKey]any)
			for key, filterVal := range condMap {
				opKey := key.(*mapper.Operator)
				var normalVal client.NormalValue
				var err error
				if filterVal == nil {
					normalVal, err = client.NewNormalNil(f.indexedFields[i].Kind)
				} else {
					normalVal, err = client.NewNormalValue(filterVal)
				}
				if err != nil {
					return nil, err
				}
				result = append(result, fieldFilterCond{
					op:   opKey.Operation,
					val:  normalVal,
					kind: f.indexedFields[i].Kind,
				})
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
