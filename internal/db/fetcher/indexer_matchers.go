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
			intVal, ok := value.Int()
			if !ok {
				if intOptVal, ok := value.NillableInt(); ok {
					intVal = intOptVal.Value()
				} else {
					return false, NewErrUnexpectedTypeValue[bool](value)
				}
			}
			boolVal = intVal != 0
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

// invertedMatcher inverts the result of the inner matcher.
type invertedMatcher struct {
	matcher valueMatcher
}

func (m *invertedMatcher) Match(val client.NormalValue) (bool, error) {
	res, err := m.matcher.Match(val)
	if err != nil {
		return false, err
	}
	return !res, nil
}

type jsonMatcher struct {
	value    float64
	evalFunc func(float64, float64) bool
}

func (m *jsonMatcher) Match(value client.NormalValue) (bool, error) {
	if jsonVal, ok := value.JSON(); ok {
		if floatVal, ok := jsonVal.Number(); ok {
			return m.evalFunc(floatVal, m.value), nil
		}
	}
	return false, NewErrUnexpectedTypeValue[float64](value)
}

func createValueMatcher(condition *fieldFilterCond) (valueMatcher, error) {
	if condition.op == "" {
		return &anyMatcher{}, nil
	}

	// TODO: test json null
	if condition.val.IsNil() {
		return &nilMatcher{matchNil: condition.op == opEq}, nil
	}

	switch condition.op {
	case opEq, opGt, opGe, opLt, opLe, opNe:
		if v, ok := condition.val.Int(); ok {
			return &intMatcher{value: v, evalFunc: getCompareValsFunc[int64](condition.op)}, nil
		}
		if v, ok := condition.val.Float(); ok {
			return &floatMatcher{value: v, evalFunc: getCompareValsFunc[float64](condition.op)}, nil
		}
		if v, ok := condition.val.String(); ok {
			return &stringMatcher{value: v, evalFunc: getCompareValsFunc[string](condition.op)}, nil
		}
		if v, ok := condition.val.Time(); ok {
			return &timeMatcher{value: v, op: condition.op}, nil
		}
		if v, ok := condition.val.Bool(); ok {
			return &boolMatcher{value: v, isEq: condition.op == opEq}, nil
		}
		if v, ok := condition.val.JSON(); ok {
			if jsonVal, ok := v.Number(); ok {
				return &jsonMatcher{value: jsonVal, evalFunc: getCompareValsFunc[float64](condition.op)}, nil
			}
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
