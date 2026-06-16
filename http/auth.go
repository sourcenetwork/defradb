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
	"net/http"
	"strings"

	"github.com/sourcenetwork/immutable"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	iIdentity "github.com/sourcenetwork/defradb/internal/identity"
)

const (
	// authHeaderName is the name of the authorization header.
	// This header should contain an ACP identity.
	authHeaderName = "Authorization"
	// authSchemaPrefix is the prefix added to the
	// authorization header value.
	authSchemaPrefix = "Bearer "
)

// AuthMiddleware authenticates an actor and sets their identity for all subsequent actions.
//
// A token is accepted when its audience matches the request host or any of the
// configured allowed origins, so a deployment behind a proxy or load balancer can
// authorize tokens minted for its public address rather than the host the server sees.
func AuthMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	// Normalize origins to host form once; the audience claim is a bare host
	// while a CORS origin may carry a scheme.
	allowedAudiences := audiencesForOrigins(allowedOrigins)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			token := strings.TrimPrefix(req.Header.Get(authHeaderName), authSchemaPrefix)
			if token == "" {
				next.ServeHTTP(rw, req)
				return
			}

			ident, err := acpIdentity.FromToken([]byte(token))
			if err != nil {
				http.Error(rw, "forbidden", http.StatusForbidden)
				return
			}

			audiences := append([]string{strings.ToLower(req.Host)}, allowedAudiences...)
			err = acpIdentity.VerifyAuthToken(ident, audiences...)
			if err != nil {
				http.Error(rw, "forbidden", http.StatusForbidden)
				return
			}

			ctx := iIdentity.WithContext(req.Context(), immutable.Some[acpIdentity.Identity](ident))
			next.ServeHTTP(rw, req.WithContext(ctx))
		})
	}
}

// audiencesForOrigins normalizes CORS origins into the audience (host) form an
// auth token would carry, so they can be compared against a token's audience
// claim. Origins that do not yield a host are skipped.
func audiencesForOrigins(origins []string) []string {
	audiences := make([]string, 0, len(origins))
	for _, origin := range origins {
		audience, err := AuthAudienceForURL(origin)
		if err != nil {
			continue
		}
		audiences = append(audiences, audience)
	}
	return audiences
}
