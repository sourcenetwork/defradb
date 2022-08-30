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

// Option represents an item that may or may not have a value.
type Option[T any] struct {
	// If HasValue is true, this Option contains a value, if
	// it is false it contains no value.
	hasValue bool

	// The Value of this Option. Should be ignored if HasValue is false.
	Value T
}

// Some returns an `Option` of type `T` with the given value.
func Some[T any](value T) Option[T] {
	return Option[T]{
		hasValue: true,
		Value:    value,
	}
}

// Some returns an `Option` of type `T` with no value.
func None[T any]() Option[T] {
	return Option[T]{}
}

// HasValue returns a boolean indicating whether or not this optino contains a value. If
// it returns true, this Option contains a value, if it is false it contains no value.
func (o Option[T]) HasValue() bool {
	return o.hasValue
}
