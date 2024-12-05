// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package base

import (
	"bytes"
	"strings"
	"time"
)

// Compare compares two values of a Document field, and determines
// which is greater.
// returns -1 if a < b
// returns 0 if a == b
// returns 1 if a > b.
//
// The only possible values for a and b is a concrete field type
// and they are always the same type as each other.
// @todo: Handle list/slice/array fields
func Compare(a, b any) int {
	if a == nil || b == nil {
		return compareNil(a, b)
	}

	switch v := a.(type) {
	case bool:
		return compareBool(v, b.(bool))
	case int:
		return compareInt(int64(v), int64(b.(int)))
	case int64:
		return compareInt(v, b.(int64))
	case uint64:
		return compareUint(v, b.(uint64))
	case float64:
		return compareFloat(v, b.(float64))
	case time.Time:
		return compareTime(v, b.(time.Time))
	case string:
		return compareString(v, b.(string))
	case []byte:
		return compareBytes(v, b.([]byte))
	default:
		return 0
	}
}

func compareNil(a, b any) int {
	if a == nil && b == nil {
		return 0
	} else if b == nil && a != nil { // a > b (1 > nil)
		return 1
	}
	return -1
}

func compareBool(a, b bool) int {
	if a == b {
		return 0
	} else if a && !b { // a > b (true > false)
		return 1
	}
	return -1
}
func compareInt(a, b int64) int {
	if a == b {
		return 0
	} else if a > b {
		return 1
	}
	return -1
}
func compareUint(a, b uint64) int {
	if a == b {
		return 0
	} else if a > b {
		return 1
	}
	return -1
}
func compareFloat(a, b float64) int {
	if a == b {
		return 0
	} else if a > b {
		return 1
	}
	return -1
}
func compareTime(a, b time.Time) int {
	if a.Equal(b) {
		return 0
	} else if a.After(b) {
		return 1
	}
	return -1
}
func compareString(a, b string) int {
	return strings.Compare(a, b)
}
func compareBytes(a, b []byte) int {
	return bytes.Compare(a, b)
}
