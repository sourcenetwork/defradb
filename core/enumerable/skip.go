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

type enumerableSkip[T any] struct {
	source Enumerable[T]
	offset uint64
	count  uint64
}

// Skip creates an `Enumerable` from the given `Enumerable` and offset. The returned
// `Enumerable` will skip through items until the number of items yielded from source
// excedes the give offset.
func Skip[T any](source Enumerable[T], offset uint64) Enumerable[T] {
	return &enumerableSkip[T]{
		source: source,
		offset: offset,
	}
}

func (s *enumerableSkip[T]) Next() (bool, error) {
	for s.count < s.offset {
		s.count += 1
		hasNext, err := s.source.Next()
		if !hasNext || err != nil {
			return hasNext, err
		}
	}
	s.count += 1
	return s.source.Next()
}

func (s *enumerableSkip[T]) Value() T {
	return s.source.Value()
}

func (s *enumerableSkip[T]) Reset() {
	s.count = 0
	s.source.Reset()
}
