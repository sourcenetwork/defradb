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
	"context"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/identity"
)

// OptionWithIdentity is an interface for options structs that provide identity access.
// This is used internally by HTTP client and C bindings for extracting identity from options.
type OptionWithIdentity[T any] interface {
	// GetIdentity returns the identity associated with this option, if any.
	GetIdentity() immutable.Option[identity.Identity]
}

// WithOptIdentity adds identity to context if present in the options struct.
// Returns the context with identity set, or the original context if no identity is present.
//
// This is used internally by the HTTP client to extract identity from options and add it to the request context.
func WithOptIdentity[T OptionWithIdentity[T]](ctx context.Context, opt T) context.Context {
	if ident := opt.GetIdentity(); ident.HasValue() {
		return identity.WithContext(ctx, ident)
	}
	return ctx
}
