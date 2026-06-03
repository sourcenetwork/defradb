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

	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/require"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/crypto"
)

// bearerTokenForAudience mints a signed bearer token whose `aud` claim is set to
// the given audience (or no audience claim at all when audience is empty).
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

// doAuthedPost issues a POST with a bearer token and returns the status code and
// the (trimmed) response body.
func doAuthedPost(t *testing.T, cdb DB, url, token, body string) (int, string) {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(body))
	req.Header.Set(authHeaderName, authSchemaPrefix+token)

	rec := httptest.NewRecorder()
	handler, err := NewHandler(cdb)
	require.NoError(t, err)
	handler.ServeHTTP(rec, req)

	res := rec.Result()
	raw, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	return res.StatusCode, strings.TrimSpace(string(raw))
}

// TestAuth_TokenAudienceMismatch_ReturnsBareForbidden reproduces the bug report
// "Policy/schema POST races node startup: returns 403 Forbidden if hit too early".
//
// The report claims the 403 is a startup warm-up race (ACP middleware / SourceHub
// grant cache not yet warm). It is not: AuthMiddleware verifies the bearer token's
// `aud` claim against the request Host (see VerifyAuthToken, which is called with
// strings.ToLower(req.Host)). When the token's audience does not match the Host the
// request is rejected in the middleware with a bare "forbidden" body, before any
// handler or ACP check runs.
//
// This is deterministic, not timing-dependent: the same token always fails against
// a mismatched Host and always passes against a matching Host. A retry "succeeding
// a second later" only works if the retry happens to use the correct Host/audience.
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
			// Host is 127.0.0.1:9181 but the token's audience is localhost:9181.
			status, body := doAuthedPost(t, cdb, tc.url, token, tc.body)

			require.Equal(t, http.StatusForbidden, status,
				"audience/Host mismatch must be rejected by AuthMiddleware")
			// The body is the bare string "forbidden" with no detail — exactly what
			// the bug report observed and complained about (improvement #1).
			require.Equal(t, "forbidden", body,
				"middleware rejection returns a bare \"forbidden\" string, not a JSON error")
		})
	}
}

// TestAuth_TokenAudienceMatch_PassesMiddleware shows the same token is accepted
// when the request Host matches the token audience — i.e. the rejection is purely a
// function of Host/audience agreement, not of how recently the node started.
//
// We only assert that the response is NOT the middleware's bare 403 "forbidden":
// once past AuthMiddleware the request reaches the handler, which may legitimately
// return other statuses (e.g. 400 for the deliberately-minimal policy body here).
func TestAuth_TokenAudienceMatch_PassesMiddleware(t *testing.T) {
	cdb := setupDatabase(t)

	token := bearerTokenForAudience(t, "localhost:9181")

	// Host matches the token audience.
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

// TestAuth_TokenWithoutAudience_ReturnsBareForbidden documents the second failure
// mode: a token minted with NO audience claim is also rejected against any Host,
// because VerifyAuthToken requires the "aud" claim to be present and to match.
func TestAuth_TokenWithoutAudience_ReturnsBareForbidden(t *testing.T) {
	cdb := setupDatabase(t)

	token := bearerTokenForAudience(t, "") // no aud claim

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
