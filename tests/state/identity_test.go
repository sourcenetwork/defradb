// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package state

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tokenWithClaims builds a JWT-shaped string carrying the given claims. Only the
// payload is read, so the header and signature are placeholders.
func tokenWithClaims(t *testing.T, claims map[string]any) string {
	t.Helper()
	payload, err := json.Marshal(claims)
	require.NoError(t, err)
	return "header." + base64.RawURLEncoding.EncodeToString(payload) + ".signature"
}

func TestTokenHasAudience_WithMatchingAudience_ReturnsTrue(t *testing.T) {
	token := tokenWithClaims(t, map[string]any{"aud": "127.0.0.1:9181"})

	assert.True(t, TokenHasAudience(token, "127.0.0.1:9181"))
}

func TestTokenHasAudience_WithStaleAudience_ReturnsFalse(t *testing.T) {
	// An external node binds a new port every start. A token minted for the old
	// address still carries an audience, so checking only that one is present
	// leaves the stale token in place and the node answers 403.
	token := tokenWithClaims(t, map[string]any{"aud": "127.0.0.1:39267"})

	assert.False(t, TokenHasAudience(token, "127.0.0.1:45123"))
}

func TestTokenHasAudience_WithNoAudienceClaim_ReturnsFalse(t *testing.T) {
	token := tokenWithClaims(t, map[string]any{"sub": "did:key:z6Mk"})

	assert.False(t, TokenHasAudience(token, "127.0.0.1:9181"))
}

func TestTokenHasAudience_WithAudienceList_MatchesAnyEntry(t *testing.T) {
	// The audience claim is allowed to be a list.
	token := tokenWithClaims(t, map[string]any{
		"aud": []string{"127.0.0.1:9181", "127.0.0.1:45123"},
	})

	assert.True(t, TokenHasAudience(token, "127.0.0.1:45123"))
	assert.False(t, TokenHasAudience(token, "127.0.0.1:39267"))
}

func TestTokenHasAudience_WithMalformedToken_ReturnsFalse(t *testing.T) {
	assert.False(t, TokenHasAudience("", "127.0.0.1:9181"))
	assert.False(t, TokenHasAudience("not-a-jwt", "127.0.0.1:9181"))
	assert.False(t, TokenHasAudience("header.!!!not-base64!!!.sig", "127.0.0.1:9181"))
	assert.False(t, TokenHasAudience("header."+
		base64.RawURLEncoding.EncodeToString([]byte("not json"))+".sig", "127.0.0.1:9181"))
}
