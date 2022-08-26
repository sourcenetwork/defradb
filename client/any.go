// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

import (
	"errors"
	"reflect"
)

// AnyToAnyList converts `any` to `[]any` if underlying type of `any` is a slice or an array (a list).
// This is because otherwise when doing a.([]any) where `a` is of type `any` we always get an error,
// it also returns an empty [] instead of preserving the elements of the underlying type `a`.
func AnyToAnyList(a any) ([]any, error) {
	switch value := reflect.ValueOf(a); value.Kind() {
	case reflect.Slice, reflect.Array:
		anyList := make([]any, value.Len(), value.Cap())
		for i := 0; i < value.Len(); i++ {
			anyList[i] = value.Index(i).Interface()
		}
		return anyList, nil
	default:
		return nil, errors.New("underlying type of `any` is not a list type.")
	}
}
