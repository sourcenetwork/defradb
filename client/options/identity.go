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

// OptionWithIdentity is an interface for options structs that provide identity access.
// This is implemented by the options structs (not builders) to provide read access to the Identity field.
type OptionWithIdentity[T any] interface {
	// GetIdentity returns the identity associated with this option, if any.
	GetIdentity() immutable.Option[identity.Identity]
}

// BuilderWithIdentity is an interface for option builders that can set identity.
// T is the options type, B is the builder type (for fluent API support).
type BuilderWithIdentity[T any, B any] interface {
	Lister[T]
	// SetIdentity sets the identity for this option and returns the builder for chaining.
	SetIdentity(id identity.Identity) B
}

// WithIdentity sets the identity on a builder if the identity is present.
// Returns the builder for chaining.
func WithIdentity[T any, B BuilderWithIdentity[T, B]](builder B, ident immutable.Option[identity.Identity]) B {
	if ident.HasValue() {
		return builder.SetIdentity(ident.Value())
	}
	return builder
}

// IdentityFrom extracts the identity from option builders by merging them and reading the identity.
// Returns an empty Option if opts is empty or no identity is set.
//
// Example usage:
//
//	func (db *DB) SomeMethod(ctx context.Context, opts ...options.Lister[options.SomeOptions]) error {
//	    ident := options.IdentityFrom[options.SomeOptions](opts...)
//	    // use ident...
//	}
func IdentityFrom[T any](opts ...Lister[T]) immutable.Option[identity.Identity] {
	merged := NewOptions(opts...)
	if withIdentity, ok := any(merged).(interface {
		GetIdentity() immutable.Option[identity.Identity]
	}); ok {
		return withIdentity.GetIdentity()
	}
	return immutable.None[identity.Identity]()
}
