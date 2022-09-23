// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package enumerable

type enumerableTake[T any] struct {
	source Enumerable[T]
	limit  int64
	count  int64
}

// Take creates an `Enumerable` from the given `Enumerable` and limit. The returned
// `Enumerable` will restrict the maximum number of items yielded to the given limit.
func Take[T any](source Enumerable[T], limit int64) Enumerable[T] {
	return &enumerableTake[T]{
		source: source,
		limit:  limit,
	}
}

func (s *enumerableTake[T]) Next() (bool, error) {
	if s.count == s.limit {
		return false, nil
	}
	s.count += 1
	return s.source.Next()
}

func (s *enumerableTake[T]) Value() T {
	return s.source.Value()
}

func (s *enumerableTake[T]) Reset() {
	s.count = 0
	s.source.Reset()
}
