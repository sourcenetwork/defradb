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

// AuthMiddleware rejects every token-verify failure with the same bare "forbidden"
// body. The tests below pin that opacity so a future fix that categorises the cause
// can't silently preserve it. They also lock in that the rejection is deterministic
// per (token, Host) pair — not a startup race, as the bug report assumed.
func TestAuth_TokenAudienceMismatch_ReturnsBareForbidden(t *testing.T) {
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
			require.Equal(t, "forbidden", body)
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
// claim to exist, not just to match. Without this case the suite couldn't tell
// "wrong audience" from "no audience" if those ever diverge in future categorisation.
func TestAuth_TokenWithoutAudience_ReturnsBareForbidden(t *testing.T) {
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
	require.Equal(t, "forbidden", body)
}

// Covers the FromToken-failure branch, which the audience cases don't reach
// (they all parse cleanly). With this in place every distinguishable auth-failure
// cause is shown to collapse to the same bare "forbidden".
func TestAuth_MalformedToken_ReturnsBareForbidden(t *testing.T) {
	cdb := setupDatabase(t)

	status, body := doAuthedPost(
		t,
		cdb,
		"http://localhost:9181/api/v1/collections",
		"not-a-jwt",
		`type Author { name: String }`,
	)

	require.Equal(t, http.StatusForbidden, status)
	require.Equal(t, "forbidden", body)
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
// the check.
func TestAuth_TokenAudienceMatchesNeitherHostNorOrigin_ReturnsForbidden(t *testing.T) {
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
	require.Equal(t, "forbidden", body)
}
