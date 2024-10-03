// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package slice

// RemoveDuplicates removes duplicates from a slice of elements in-place.
// The algorithm is not stable, and the order of the elements may change.
// A new slice is returned just to update the length of the given slice.
func RemoveDuplicates[T comparable](elements []T) []T {
	sets := make(map[T]struct{})
	for i := len(elements) - 1; i >= 0; i-- {
		if _, ok := sets[elements[i]]; ok {
			elements[i] = elements[len(elements)-1]
			elements = elements[:len(elements)-1]
		} else {
			sets[elements[i]] = struct{}{}
		}
	}
	return elements
}
