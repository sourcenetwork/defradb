// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package identity

import (
	"context"

	"github.com/sourcenetwork/immutable"
)

// identityContextKey is the key type for ACP identity context values.
type identityContextKey struct{}

// FromContext returns the identity from the given context.
//
// If an identity does not exist `NoIdentity` is returned.
func FromContext(ctx context.Context) immutable.Option[Identity] {
	identity, ok := ctx.Value(identityContextKey{}).(Identity)
	if ok {
		return immutable.Some(identity)
	}
	return None
}

// WithContext returns a new context with the identity value set.
//
// This will overwrite any previously set identity value.
func WithContext(ctx context.Context, identity immutable.Option[Identity]) context.Context {
	if identity.HasValue() {
		return context.WithValue(ctx, identityContextKey{}, identity.Value())
	}
	return context.WithValue(ctx, identityContextKey{}, nil)
}
