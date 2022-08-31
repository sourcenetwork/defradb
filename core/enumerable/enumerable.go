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

// Enumerable represents a set of elements that can be iterated through
// multiple times.
//
// The enumerable may be a composite of multiple actions that will be lazily
// executed upon iteration, allowing the enumerable to be constructed out of a
// complex set of instructions that can be evaluated in a single iteration of the
// underlying set.
type Enumerable[T any] interface {
	// Next attempts to evaluate the next item in the enumeration - allowing its
	// exposure via the `Value()` function.
	//
	// It will return false if it has reached the end of the enumerable, and/or an
	// error if one was generated during evaluation.
	Next() (bool, error)

	// Value returns the current item in the enumeration. It does not progress the
	// enumeration, and should be a simple getter.
	//
	// If the previous Next call did not return true, or Next has never been called
	// the behaviour and return value of this function is undefined.
	Value() T

	// Reset resets the enumerable, allowing for re-iteration.
	Reset()
}

type enumerableSlice[T any] struct {
	source       []T
	currentIndex int
	maxIndex     int
}

// New creates an `Enumerable` from the given slice.
func New[T any](source []T) Enumerable[T] {
	return &enumerableSlice[T]{
		source:       source,
		currentIndex: -1,
		maxIndex:     len(source) - 1,
	}
}

func (s *enumerableSlice[T]) Next() (bool, error) {
	if s.currentIndex == s.maxIndex {
		return false, nil
	}
	s.currentIndex += 1
	return true, nil
}

func (s *enumerableSlice[T]) Value() T {
	return s.source[s.currentIndex]
}

func (s *enumerableSlice[T]) Reset() {
	s.currentIndex = -1
}

// ForEach iterates over the given source `Enumerable` performing the given
// action on each item. It resets the source `Enumerable` on completion.
func ForEach[T any](source Enumerable[T], action func(item T)) error {
	for {
		hasNext, err := source.Next()
		if err != nil {
			return err
		}
		if !hasNext {
			break
		}
		item := source.Value()
		action(item)
	}
	source.Reset()
	return nil
}

// OnEach iterates over the given source `Enumerable` performing the given
// action for each item yielded. It resets the source `Enumerable` on completion.
func OnEach[T any](source Enumerable[T], action func()) error {
	for {
		hasNext, err := source.Next()
		if err != nil {
			return err
		}
		if !hasNext {
			break
		}
		action()
	}
	source.Reset()
	return nil
}
