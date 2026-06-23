// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package http

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/immutable"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/crypto"
)

// Pass an empty audience to omit the `aud` claim entirely.
func bearerTokenForAudience(t *testing.T, audience string) string {
	t.Helper()

	full, err := acpIdentity.Generate(crypto.KeyTypeSecp256k1)
	require.NoError(t, err)

	aud := immutable.None[string]()
	if audience != "" {
		aud = immutable.Some(audience)
	}

	token, err := full.NewToken(time.Hour, aud, immutable.None[string]())
	require.NoError(t, err)

	return string(token)
}

// tokenWithInvalidSubject returns a validly-signed token whose subject is not a
// valid public key, so FromToken rejects it with ErrInvalidSubject.
func tokenWithInvalidSubject(t *testing.T, audience string) string {
	t.Helper()

	full, err := acpIdentity.Generate(crypto.KeyTypeSecp256k1)
	require.NoError(t, err)

	now := time.Now()
	tok, err := jwt.NewBuilder().
		Subject("not-a-public-key").
		Audience([]string{audience}).
		Expiration(now.Add(time.Hour)).
		NotBefore(now).
		IssuedAt(now).
		Build()
	require.NoError(t, err)
	require.NoError(t, tok.Set(acpIdentity.KeyTypeClaim, string(crypto.KeyTypeSecp256k1)))

	privKey := full.PrivateKey().Underlying()
	secpPrivKey, ok := privKey.(*secp256k1.PrivateKey)
	require.True(t, ok)
	signed, err := jwt.Sign(tok, jwt.WithKey(jwa.ES256K, secpPrivKey.ToECDSA()))
	require.NoError(t, err)

	return string(signed)
}

// tamperedSignatureToken returns a token that parses and carries the given
// audience but whose signature no longer verifies against the signer's key. The
// subject (public key) is left intact so FromToken still derives the right key;
// only the signature segment is corrupted, so verification fails at jws.Verify.
func tamperedSignatureToken(t *testing.T, audience string) string {
	t.Helper()

	token := bearerTokenForAudience(t, audience)

	parts := strings.Split(token, ".")
	require.Len(t, parts, 3, "expected a three-segment JWT")

	// Mutate a character mid-segment so the decoded signature bytes actually change
	// (flipping the final base64url character can leave them unchanged). The segment
	// still decodes, but the signature no longer verifies.
	sig := []byte(parts[2])
	require.Greater(t, len(sig), 1)
	i := len(sig) / 2
	if sig[i] == 'A' {
		sig[i] = 'B'
	} else {
		sig[i] = 'A'
	}
	parts[2] = string(sig)

	return strings.Join(parts, ".")
}

func doAuthedPost(t *testing.T, cdb DB, url, token, body string) (int, string) {
	t.Helper()
	return doAuthedPostWithOrigins(t, cdb, nil, url, token, body)
}

// doAuthedPostWithOrigins builds a handler whose AuthMiddleware is configured with
// the given allowed origins, so the audience allow-list behaviour can be exercised.
func doAuthedPostWithOrigins(t *testing.T, cdb DB, allowedOrigins []string, url, token, body string) (int, string) {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(body))
	req.Header.Set(authHeaderName, authSchemaPrefix+token)

	rec := httptest.NewRecorder()
	var nodeOpts *options.NodeOptions
	if allowedOrigins != nil {
		nodeOpts = &options.NodeOptions{
			HTTP: options.NodeHTTPOptions{AllowedOrigins: allowedOrigins},
		}
	}
	handler, err := NewHandler(cdb, nodeOpts)
	require.NoError(t, err)
	handler.ServeHTTP(rec, req)

	res := rec.Result()
	raw, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	return res.StatusCode, strings.TrimSpace(string(raw))
}

// AuthMiddleware rejects each distinct token-verify failure with a descriptive 403
// body identifying the operator-actionable cause. The tests below pin those reason
// strings (the stable contract) and lock in that the rejection is deterministic per
// (token, Host) pair — not a startup race, as the bug report assumed.
func TestAuth_TokenAudienceMismatch_ReturnsMismatchReason(t *testing.T) {
	cdb := setupDatabase(t)

	// Token minted for audience "localhost:9181" — the typical provisioning host.
	token := bearerTokenForAudience(t, "localhost:9181")

	collectionSDL := `type Author { name: String }`
	policyYAML := "name: test\ndescription: a policy\n"

	testCases := []struct {
		name string
		url  string
		body string
	}{
		{
			name: "POST /api/v1/collections",
			url:  "http://127.0.0.1:9181/api/v1/collections",
			body: collectionSDL,
		},
		{
			name: "POST /api/v1/acp/document/policy",
			url:  "http://127.0.0.1:9181/api/v1/acp/document/policy",
			body: policyYAML,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			status, body := doAuthedPost(t, cdb, tc.url, token, tc.body)

			require.Equal(t, http.StatusForbidden, status)
			require.Equal(t, authErrAudienceMismatch, body)
		})
	}
}

// Counter-test for the mismatch cases: a matching Host must not produce the bare
// 403 — otherwise the assertions above would also pass on a middleware that rejects
// everything. The handler past the middleware may still 4xx for unrelated reasons,
// so we only rule out the middleware's specific rejection signature.
func TestAuth_TokenAudienceMatch_PassesMiddleware(t *testing.T) {
	cdb := setupDatabase(t)

	token := bearerTokenForAudience(t, "localhost:9181")

	status, body := doAuthedPost(
		t,
		cdb,
		"http://localhost:9181/api/v1/collections",
		token,
		`type Author { name: String }`,
	)

	require.False(t, status == http.StatusForbidden && body == "forbidden",
		"matching Host/audience must pass AuthMiddleware (got %d %q)", status, body)
}

// A missing `aud` claim is its own failure mode — VerifyAuthToken requires the
// claim to exist, not just to match — and now reports a distinct reason from a
// wrong audience.
func TestAuth_TokenWithoutAudience_ReturnsMissingAudienceReason(t *testing.T) {
	cdb := setupDatabase(t)

	token := bearerTokenForAudience(t, "")

	status, body := doAuthedPost(
		t,
		cdb,
		"http://localhost:9181/api/v1/collections",
		token,
		`type Author { name: String }`,
	)

	require.Equal(t, http.StatusForbidden, status)
	require.Equal(t, authErrMissingAudience, body)
}

// Covers the FromToken-failure branch, which the audience cases don't reach
// (they all parse cleanly). A malformed token reports the generic invalid-token
// reason — the same bucket as a signature failure, so the two are not
// distinguishable to a caller.
func TestAuth_MalformedToken_ReturnsInvalidTokenReason(t *testing.T) {
	cdb := setupDatabase(t)

	status, body := doAuthedPost(
		t,
		cdb,
		"http://localhost:9181/api/v1/collections",
		"not-a-jwt",
		`type Author { name: String }`,
	)

	require.Equal(t, http.StatusForbidden, status)
	require.Equal(t, authErrInvalidToken, body)
}

// A token whose signature does not verify must report the generic invalid-token
// reason, never a signature-specific message — distinguishing it would let an
// attacker probe key validity. We forge a token with a valid audience but signed
// by a different key than the one its DID/public key implies.
func TestAuth_InvalidSignature_ReturnsInvalidTokenReason(t *testing.T) {
	cdb := setupDatabase(t)

	token := tamperedSignatureToken(t, "localhost:9181")

	status, body := doAuthedPost(
		t,
		cdb,
		"http://localhost:9181/api/v1/collections",
		token,
		`type Author { name: String }`,
	)

	require.Equal(t, http.StatusForbidden, status)
	require.Equal(t, authErrInvalidToken, body)
}

// An expired token reports its own cause rather than the generic reason, so an
// operator can see they need to mint a fresh token.
func TestAuth_ExpiredToken_ReturnsExpiredReason(t *testing.T) {
	cdb := setupDatabase(t)

	full, err := acpIdentity.Generate(crypto.KeyTypeSecp256k1)
	require.NoError(t, err)
	tokenBytes, err := full.NewToken(-time.Hour, immutable.Some("localhost:9181"), immutable.None[string]())
	require.NoError(t, err)

	status, body := doAuthedPost(
		t,
		cdb,
		"http://localhost:9181/api/v1/collections",
		string(tokenBytes),
		`type Author { name: String }`,
	)

	require.Equal(t, http.StatusForbidden, status)
	require.Equal(t, authErrExpired, body)
}

// A token whose subject is not a valid public key reports the invalid-subject
// cause, so an operator can see they put the wrong value in the `sub` claim.
func TestAuth_InvalidSubject_ReturnsInvalidSubjectReason(t *testing.T) {
	cdb := setupDatabase(t)

	token := tokenWithInvalidSubject(t, "localhost:9181")

	status, body := doAuthedPost(
		t,
		cdb,
		"http://localhost:9181/api/v1/collections",
		token,
		`type Author { name: String }`,
	)

	require.Equal(t, http.StatusForbidden, status)
	require.Equal(t, authErrInvalidSubject, body)
}

// A token whose audience does not match the request host is accepted when the
// audience matches one of the configured allowed origins. This lets a server
// behind a proxy authorize tokens minted for its public address even though the
// host it directly sees differs.
func TestAuth_TokenAudienceMatchesAllowedOrigin_PassesMiddleware(t *testing.T) {
	cdb := setupDatabase(t)

	// Token minted for the public address, while the request reaches the server
	// on a different host (e.g. behind a proxy).
	token := bearerTokenForAudience(t, "public.example.com")

	status, body := doAuthedPostWithOrigins(
		t,
		cdb,
		[]string{"http://public.example.com"},
		"http://127.0.0.1:9181/api/v1/collections",
		token,
		`type Author { name: String }`,
	)

	require.False(t, status == http.StatusForbidden && body == "forbidden",
		"audience matching an allowed origin must pass AuthMiddleware (got %d %q)", status, body)
}

// A token whose audience matches neither the request host nor any allowed origin
// is still rejected — the allow-list widens the accepted set, it does not disable
// the check — and reports the audience-mismatch reason.
func TestAuth_TokenAudienceMatchesNeitherHostNorOrigin_ReturnsMismatchReason(t *testing.T) {
	cdb := setupDatabase(t)

	token := bearerTokenForAudience(t, "evil.example.com")

	status, body := doAuthedPostWithOrigins(
		t,
		cdb,
		[]string{"http://public.example.com"},
		"http://127.0.0.1:9181/api/v1/collections",
		token,
		`type Author { name: String }`,
	)

	require.Equal(t, http.StatusForbidden, status)
	require.Equal(t, authErrAudienceMismatch, body)
}
