// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package options

import (
	"github.com/sourcenetwork/immutable/enumerable"
)

// Lister is a type alias for enumerable.Enumerable.
// This allows option builders to be passed as enumerable collections,
// and also allows users to pass raw slices via enumerable.New().
//
// Option builders implement enumerable.Enumerable via Next(), Value(), and Reset() methods.
// This provides flexibility for advanced users to define custom option functions
// without using the builder pattern.
type Lister[T any] = enumerable.Enumerable[func(*T)]

// enumerableBuilder provides a reusable implementation of enumerable.Enumerable
// for option builders. All builders can embed this struct to gain enumerable
// capabilities without reimplementing Next(), Value(), and Reset().
type enumerableBuilder[T any] struct {
	opts  []func(*T)
	index int
}

// Next implements enumerable.Enumerable interface.
func (b *enumerableBuilder[T]) Next() (bool, error) {
	if b.index < len(b.opts) {
		b.index++
		return true, nil
	}
	return false, nil
}

// Value implements enumerable.Enumerable interface.
func (b *enumerableBuilder[T]) Value() (func(*T), error) {
	if b.index > 0 && b.index <= len(b.opts) {
		return b.opts[b.index-1], nil
	}
	return nil, nil
}

// Reset implements enumerable.Enumerable interface.
func (b *enumerableBuilder[T]) Reset() {
	b.index = 0
}

// append adds a functional option to the builder.
func (b *enumerableBuilder[T]) append(fn func(*T)) {
	b.opts = append(b.opts, fn)
}
