// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package utils

import (
	"reflect"

	"github.com/sourcenetwork/immutable/enumerable"
)

// NewOptions merges multiple option builders into a single options struct.
// It applies all functional options from all builders in the order they are provided.
// Returns nil if no valid options are provided.
//
// This follows the MongoDB Go driver pattern for option merging.
// Option builders implement enumerable.Enumerable, allowing iteration via Next()/Value().
//
// Example usage:
//
//	opts := options.NewOptions(
//	    options.GetCollections().SetIdentity(id),
//	    options.GetCollections().SetVersionID(vid),
//	)
func NewOptions[T any](opts ...enumerable.Enumerable[func(*T)]) *T {
	args := new(T)
	ApplyOptions(args, opts...)
	return args
}

// ApplyOptions applies all functional options onto the given target.
func ApplyOptions[T any](target *T, opts ...enumerable.Enumerable[func(*T)]) {
	for _, opt := range opts {
		if opt == nil || reflect.ValueOf(opt).IsNil() {
			continue
		}
		for {
			hasNext, err := opt.Next()
			if err != nil || !hasNext {
				break
			}
			setArgs, err := opt.Value()
			if err != nil {
				break
			}
			if setArgs != nil {
				setArgs(target)
			}
		}
		opt.Reset()
	}
}
