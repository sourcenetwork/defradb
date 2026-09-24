// Copyright 2026 Democratized Data Foundation
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
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/crypto"
)

func TestAuthAudienceForURL(t *testing.T) {
	tests := []struct {
		name     string
		rawURL   string
		expected string
	}{
		{
			name:     "host and port without scheme",
			rawURL:   "127.0.0.1:9181",
			expected: "127.0.0.1:9181",
		},
		{
			name:     "http scheme is stripped from the audience",
			rawURL:   "http://127.0.0.1:9181",
			expected: "127.0.0.1:9181",
		},
		{
			name:     "https scheme is supported",
			rawURL:   "https://example.com",
			expected: "example.com",
		},
		{
			name:     "host is lower-cased to match the server check",
			rawURL:   "http://Example.COM:9181",
			expected: "example.com:9181",
		},
		{
			// Go's net/http does not strip default ports from the wire Host
			// header, so the audience must keep them to match the server check.
			name:     "default http port is preserved",
			rawURL:   "http://example.com:80",
			expected: "example.com:80",
		},
		{
			name:     "default https port is preserved",
			rawURL:   "https://example.com:443",
			expected: "example.com:443",
		},
		{
			name:     "ipv6 host is supported",
			rawURL:   "http://[::1]:9181",
			expected: "[::1]:9181",
		},
		{
			// A scheme-less host that merely begins with "http" must still get
			// a scheme prepended so it parses with a non-empty host.
			name:     "scheme-less host beginning with http",
			rawURL:   "httpbin.local:9181",
			expected: "httpbin.local:9181",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AuthAudienceForURL(tt.rawURL)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestAuthAudienceForURL_Errors(t *testing.T) {
	tests := []struct {
		name   string
		rawURL string
	}{
		{
			name:   "malformed url fails to parse",
			rawURL: "http://[::1]:invalid",
		},
		{
			name:   "url with no derivable host",
			rawURL: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AuthAudienceForURL(tt.rawURL)
			require.Error(t, err)
			assert.Empty(t, got)
		})
	}
}

// The audience the CLI mints must equal the Host the dialed client sends, so
// that the server's AuthMiddleware (which compares against the request Host)
// accepts the token. This pins the two derivations to the same source URL.
func TestAuthAudienceForURL_MatchesDialedClientHost(t *testing.T) {
	const rawURL = "0.0.0.0:19181"

	audience, err := AuthAudienceForURL(rawURL)
	require.NoError(t, err)

	client, err := newHttpClient(rawURL)
	require.NoError(t, err)

	assert.Equal(t, client.baseURL.Host, audience)
}

func TestClientDefaultIdentity(t *testing.T) {
	authorization := make(chan string, 1)
	server := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		authorization <- req.Header.Get(authHeaderName)
		rw.Header().Set("Content-Type", "application/json")
		if _, err := rw.Write([]byte(`{"id":1}`)); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	audience, err := AuthAudienceForURL(server.URL)
	require.NoError(t, err)
	ident, err := identity.Generate(crypto.KeyTypeEd25519)
	require.NoError(t, err)
	require.NoError(t, ident.UpdateToken(time.Hour, immutable.Some(audience), immutable.None[string]()))

	client, err := NewClient(
		server.URL,
		WithHTTPClient(server.Client()),
		WithIdentity(ident),
	)
	require.NoError(t, err)
	_, err = client.NewTxn(true)
	require.NoError(t, err)

	assert.Equal(t, authSchemaPrefix+ident.BearerToken(), <-authorization)
}
