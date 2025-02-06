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
	"cmp"
	"strings"
	"time"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/keys"
)

func executeValueMatchers(matchers []valueMatcher, fields []keys.IndexedField) (bool, error) {
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

type float32Matcher struct {
	value    float32
	evalFunc func(float32, float32) bool
}

func (m *float32Matcher) Match(value client.NormalValue) (bool, error) {
	if floatVal, ok := value.Float32(); ok {
		return m.evalFunc(floatVal, m.value), nil
	}
	if floatOptVal, ok := value.NillableFloat32(); ok {
		if !floatOptVal.HasValue() {
			return false, nil
		}
		return m.evalFunc(floatOptVal.Value(), m.value), nil
	}
	return false, NewErrUnexpectedTypeValue[float32](value)
}

type float64Matcher struct {
	value    float64
	evalFunc func(float64, float64) bool
}

func (m *float64Matcher) Match(value client.NormalValue) (bool, error) {
	if floatVal, ok := value.Float64(); ok {
		return m.evalFunc(floatVal, m.value), nil
	}
	if floatOptVal, ok := value.NillableFloat64(); ok {
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

type boolMatcher struct {
	value bool
	isEq  bool
}

func (m *boolMatcher) Match(value client.NormalValue) (bool, error) {
	boolVal, ok := value.Bool()
	if !ok {
		if boolOptVal, ok := value.NillableBool(); ok {
			boolVal = boolOptVal.Value()
		} else {
			return false, NewErrUnexpectedTypeValue[bool](value)
		}
	}
	return boolVal == m.value == m.isEq, nil
}

type nilMatcher struct {
	matchNil bool
}

func (m *nilMatcher) Match(value client.NormalValue) (bool, error) {
	return value.IsNil() == m.matchNil, nil
}

// compositeMatcher checks if the value satisfies any of the matchers
type compositeMatcher struct {
	matchers []valueMatcher
}

func newCompositeMatcher(matchers ...valueMatcher) *compositeMatcher {
	return &compositeMatcher{matchers: matchers}
}

func (m *compositeMatcher) Match(value client.NormalValue) (bool, error) {
	for i := range m.matchers {
		res, err := m.matchers[i].Match(value)
		if err != nil {
			return false, err
		}
		if res {
			return true, nil
		}
	}
	return false, nil
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

func (m *indexLikeMatcher) Match(val client.NormalValue) (bool, error) {
	strVal, ok := val.String()
	if !ok {
		if strOptVal, ok := val.NillableString(); ok {
			strVal = strOptVal.Value()
		} else if jsonVal, ok := val.JSON(); ok {
			strVal, ok = jsonVal.String()
			if !ok {
				return false, nil
			}
		} else {
			return false, NewErrUnexpectedTypeValue[string](val)
		}
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

type jsonComparingMatcher[T comparable] struct {
	value        T
	getValueFunc func(client.JSON) (T, bool)
	evalFunc     func(T, T) bool
}

func (m *jsonComparingMatcher[T]) Match(value client.NormalValue) (bool, error) {
	if jsonVal, ok := value.JSON(); ok {
		if val, ok := m.getValueFunc(jsonVal); ok {
			return m.evalFunc(val, m.value), nil
		}
		return false, nil
	}
	return false, NewErrUnexpectedTypeValue[float64](value)
}

// jsonTypeMatcher checks if the value is of the expected type
type jsonTypeMatcher[T comparable] struct {
	getValueFunc func(client.JSON) (T, bool)
	shouldMatch  bool
}

func (m *jsonTypeMatcher[T]) Match(value client.NormalValue) (bool, error) {
	if jsonVal, ok := value.JSON(); ok {
		_, ok := m.getValueFunc(jsonVal)
		return ok == m.shouldMatch, nil
	}
	return false, NewErrUnexpectedTypeValue[float64](value)
}

type jsonBoolMatcher struct {
	value bool
	isEq  bool
}

func (m *jsonBoolMatcher) Match(value client.NormalValue) (bool, error) {
	if jsonVal, ok := value.JSON(); ok {
		boolVal, ok := jsonVal.Bool()
		if ok {
			return boolVal == m.value == m.isEq, nil
		}
		return false, nil
	}
	return false, NewErrUnexpectedTypeValue[bool](value)
}

type jsonNullMatcher struct {
	matchNull bool
}

func (m *jsonNullMatcher) Match(value client.NormalValue) (bool, error) {
	if jsonVal, ok := value.JSON(); ok {
		return jsonVal.IsNull() == m.matchNull, nil
	}
	return false, NewErrUnexpectedTypeValue[any](value)
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
		return createComparingMatcher(condition), nil
	case opIn, opNin:
		inVals, err := client.ToArrayOfNormalValues(condition.val)
		if err != nil {
			return nil, err
		}
		return &indexInArrayMatcher{inValues: inVals, isIn: condition.op == opIn}, nil
	case opLike, opNlike, opILike, opNILike:
		strVal, err := extractStringFromNormalValue(condition.val)
		if err != nil {
			return nil, err
		}
		isLike := condition.op == opLike || condition.op == opILike
		isCaseInsensitive := condition.op == opILike || condition.op == opNILike
		return newLikeIndexCmp(strVal, isLike, isCaseInsensitive)
	case opAny:
		return &anyMatcher{}, nil
	}

	return nil, NewErrInvalidFilterOperator(condition.op)
}

func createComparingMatcher(condition *fieldFilterCond) valueMatcher {
	// JSON type needs a special handling if the op is _ne, because _ne should check also
	// difference of types
	if v, ok := condition.val.JSON(); ok {
		// we have a separate branch for null matcher because default matching behavior
		// is what we need: for filter `_ne: null` it will match all non-null values
		if v.IsNull() {
			return &jsonNullMatcher{matchNull: condition.op == opEq}
		}

		if condition.op != opNe {
			return createJSONComparingMatcher(v, condition.op)
		}

		// _ne filter on JSON fields should also accept values of different types
		var typeMatcher valueMatcher
		if _, ok := v.Number(); ok {
			typeMatcher = &jsonTypeMatcher[float64]{
				getValueFunc: func(j client.JSON) (float64, bool) { return j.Number() },
			}
		} else if _, ok := v.String(); ok {
			typeMatcher = &jsonTypeMatcher[string]{
				getValueFunc: func(j client.JSON) (string, bool) { return j.String() },
			}
		} else if _, ok := v.Bool(); ok {
			typeMatcher = &jsonTypeMatcher[bool]{
				getValueFunc: func(j client.JSON) (bool, bool) { return j.Bool() },
			}
		}
		return newCompositeMatcher(typeMatcher, createJSONComparingMatcher(v, condition.op))
	}

	matcher := createScalarComparingMatcher(condition)

	// for _ne filter on regular (non-JSON) fields the index should also accept nil values
	// there won't be `_ne: null` because nil check is done before this function is called
	if condition.op == opNe {
		matcher = newCompositeMatcher(&nilMatcher{matchNil: true}, matcher)
	}

	return matcher
}

// createJSONComparingMatcher creates a matcher for JSON values
func createJSONComparingMatcher(val client.JSON, op string) valueMatcher {
	if jsonVal, ok := val.Number(); ok {
		return &jsonComparingMatcher[float64]{
			value:        jsonVal,
			getValueFunc: func(j client.JSON) (float64, bool) { return j.Number() },
			evalFunc:     getCompareValsFunc[float64](op),
		}
	} else if jsonVal, ok := val.String(); ok {
		return &jsonComparingMatcher[string]{
			value:        jsonVal,
			getValueFunc: func(j client.JSON) (string, bool) { return j.String() },
			evalFunc:     getCompareValsFunc[string](op),
		}
	} else if jsonVal, ok := val.Bool(); ok {
		return &jsonBoolMatcher{value: jsonVal, isEq: op == opEq}
	}
	return nil
}

// createScalarComparingMatcher creates a matcher for scalar values (int, float, string, time, bool)
func createScalarComparingMatcher(condition *fieldFilterCond) valueMatcher {
	if v, ok := condition.val.Int(); ok {
		return &intMatcher{value: v, evalFunc: getCompareValsFunc[int64](condition.op)}
	} else if v, ok := condition.val.Float32(); ok {
		return &float32Matcher{value: v, evalFunc: getCompareValsFunc[float32](condition.op)}
	} else if v, ok := condition.val.Float64(); ok {
		return &float64Matcher{value: v, evalFunc: getCompareValsFunc[float64](condition.op)}
	} else if v, ok := condition.val.String(); ok {
		return &stringMatcher{value: v, evalFunc: getCompareValsFunc[string](condition.op)}
	} else if v, ok := condition.val.Time(); ok {
		return &timeMatcher{value: v, op: condition.op}
	} else if v, ok := condition.val.Bool(); ok {
		return &boolMatcher{value: v, isEq: condition.op == opEq}
	}
	return nil
}

func extractStringFromNormalValue(val client.NormalValue) (string, error) {
	strVal, ok := val.String()
	if !ok {
		if strOptVal, ok := val.NillableString(); ok {
			strVal = strOptVal.Value()
		} else if jsonVal, ok := val.JSON(); ok {
			strVal, ok = jsonVal.String()
			if !ok {
				return "", NewErrUnexpectedTypeValue[string](jsonVal)
			}
		} else {
			return "", NewErrUnexpectedTypeValue[string](val)
		}
	}
	return strVal, nil
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
