// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package coreblock

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/crypto"
	iIdentity "github.com/sourcenetwork/defradb/internal/identity"
	"github.com/sourcenetwork/immutable"
)

func TestEnabledSigningFromContext_WithFullIdentity_ReturnsIdentity(t *testing.T) {
	ident, err := identity.Generate(crypto.KeyTypeSecp256k1)
	require.NoError(t, err)

	ctx := context.Background()
	ctx = iIdentity.WithContext(ctx, immutable.Some[identity.Identity](ident))
	ctx = ContextWithEnabledSigning(ctx)

	enabled, fullIdent := EnabledSigningFromContext(ctx)
	require.True(t, enabled)
	require.True(t, fullIdent.HasValue())
	require.NotNil(t, fullIdent.Value().PrivateKey())
}

func TestEnabledSigningFromContext_WithTokenIdentityNilPrivateKey_ReturnsNone(t *testing.T) {
	// Generate an identity and create a bearer token from it.
	ident, err := identity.Generate(crypto.KeyTypeSecp256k1)
	require.NoError(t, err)
	err = ident.UpdateToken(time.Hour, immutable.Some("test"), immutable.None[string]())
	require.NoError(t, err)

	// Parse the token back, this creates a FullIdentity with nil PrivateKey.
	// This is exactly what happens on the server side when a client sends a
	// request with a bearer token (e.g. --no-keyring with --identity).
	tokenIdent, err := identity.FromToken([]byte(ident.BearerToken()))
	require.NoError(t, err)

	ctx := context.Background()
	ctx = iIdentity.WithContext(ctx, immutable.Some[identity.Identity](tokenIdent))
	ctx = ContextWithEnabledSigning(ctx)

	enabled, fullIdent := EnabledSigningFromContext(ctx)
	require.True(t, enabled)

	// If extractFullIdentity returns Some(fullIdent) with nil PrivateKey, that can
	// lead to a panic in signBlock. After the fix, it returns None so signing is skipped.
	require.False(t, fullIdent.HasValue())
}

func TestEnabledSigningFromContext_WithNoIdentity_ReturnsNone(t *testing.T) {
	ctx := context.Background()
	ctx = ContextWithEnabledSigning(ctx)

	enabled, fullIdent := EnabledSigningFromContext(ctx)
	require.True(t, enabled)
	require.False(t, fullIdent.HasValue())
}

func TestEnabledSigningFromContext_WithSigningDisabled_ReturnsFalse(t *testing.T) {
	ctx := context.Background()

	enabled, fullIdent := EnabledSigningFromContext(ctx)
	require.False(t, enabled)
	require.False(t, fullIdent.HasValue())
}
