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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AuthAudienceForURL(tt.rawURL)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
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
