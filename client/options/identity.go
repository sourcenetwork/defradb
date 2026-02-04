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
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/identity"
)

// OptionWithIdentity is an interface for options that provide and can set an identity.
// T is the concrete options type (for fluent API support).
type OptionWithIdentity[T any] interface {
	// GetIdentity returns the identity associated with this option, if any.
	GetIdentity() immutable.Option[identity.Identity]
	// SetIdentity sets the identity for this option and returns the option for chaining.
	SetIdentity(id identity.Identity) T
}

// WithIdentity sets the identity on an option if the identity is present.
// Returns the option for chaining.
func WithIdentity[T OptionWithIdentity[T]](opt T, ident immutable.Option[identity.Identity]) T {
	if ident.HasValue() {
		opt.SetIdentity(ident.Value())
	}
	return opt
}

// IdentityFrom extracts the identity from the first non-nil option in the slice.
// Returns an empty Option if opts is empty or the first option has no identity set.
//
// Example usage:
//
//	func (db *DB) SomeMethod(ctx context.Context, opts ...*options.SomeOptions) error {
//	    ident := options.IdentityFrom(opts...)
//	    // use ident...
//	}
func IdentityFrom[T OptionWithIdentity[T]](opts ...T) immutable.Option[identity.Identity] {
	for _, opt := range opts {
		// Check for nil using any type assertion since T is constrained to pointer types
		// that implement OptionWithIdentity
		if any(opt) != nil {
			return opt.GetIdentity()
		}
	}
	return immutable.None[identity.Identity]()
}
