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

// BuilderWithIdentity is an interface for option builders that can set identity.
// T is the options type, B is the builder type (for fluent API support).
type BuilderWithIdentity[T any, B any] interface {
	Enumerable[T]
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
