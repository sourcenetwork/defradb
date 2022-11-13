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

type enumerableConcat[T any] struct {
	sources            []Enumerable[T]
	currentSourceIndex int
}

// Concat takes zero to many source `Ènumerable`s and stacks them on top
// of each other, resulting in one enumerable that will iterate through all
// the values in all of the given sources.
func Concat[T any](sources ...Enumerable[T]) Enumerable[T] {
	return &enumerableConcat[T]{
		sources:            sources,
		currentSourceIndex: 0,
	}
}

func (s *enumerableConcat[T]) Next() (bool, error) {
	for {
		if s.currentSourceIndex >= len(s.sources) {
			return false, nil
		}

		currentSource := s.sources[s.currentSourceIndex]
		hasValue, err := currentSource.Next()
		if err != nil {
			return false, nil
		}
		if hasValue {
			return true, nil
		}

		s.currentSourceIndex += 1
	}
}

func (s *enumerableConcat[T]) Value() T {
	return s.sources[s.currentSourceIndex].Value()
}

func (s *enumerableConcat[T]) Reset() {
	s.currentSourceIndex = 0
	for _, source := range s.sources {
		source.Reset()
	}
}
