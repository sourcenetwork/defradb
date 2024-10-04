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

import "github.com/sourcenetwork/immutable"

// RemoveDuplicates removes duplicates from a slice of elements.
// Relative order of the elements is not preserved.
// Both runtime and space complexity are O(n).
func RemoveDuplicates[S ~[]E, E comparable](s S) S {
	sets := make(map[E]struct{})
	for i := len(s) - 1; i >= 0; i-- {
		if _, ok := sets[s[i]]; ok {
			swapLast(s, i)
			s = s[:len(s)-1]
		} else {
			sets[s[i]] = struct{}{}
		}
	}
	return s
}

// RemoveFirstIf removes the first element that satisfies the predicate.
// Relative order of the elements is not preserved, as the last element is swapped with the removed one.
func RemoveFirstIf[S ~[]E, E any](s S, predicate func(E) bool) (S, immutable.Option[E]) {
	for i := 0; i < len(s); i++ {
		if predicate(s[i]) {
			swapLast(s, i)
			lastInd := len(s) - 1
			return s[:lastInd], immutable.Some(s[lastInd])
		}
	}
	return s, immutable.None[E]()
}

func swap[T any](elements []T, i, j int) {
	elements[i], elements[j] = elements[j], elements[i]
}

func swapLast[T any](elements []T, i int) {
	swap(elements, i, len(elements)-1)
}
