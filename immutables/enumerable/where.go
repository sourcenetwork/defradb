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

type enumerableWhere[T any] struct {
	source    Enumerable[T]
	predicate func(T) (bool, error)
}

// Where creates an `Enumerable` from the given `Enumerable` and predicate. Items in the
// source `Enumerable` must return true when passed into the predicate in order to be yielded
// from the returned `Enumerable`.
func Where[T any](source Enumerable[T], predicate func(T) (bool, error)) Enumerable[T] {
	return &enumerableWhere[T]{
		source:    source,
		predicate: predicate,
	}
}

func (s *enumerableWhere[T]) Next() (bool, error) {
	for {
		hasNext, err := s.source.Next()
		if !hasNext || err != nil {
			return hasNext, err
		}

		value := s.source.Value()
		if passes, err := s.predicate(value); passes || err != nil {
			return passes, err
		}
	}
}

func (s *enumerableWhere[T]) Value() T {
	return s.source.Value()
}

func (s *enumerableWhere[T]) Reset() {
	s.source.Reset()
}
