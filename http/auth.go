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
	"github.com/sourcenetwork/defradb/errors"
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

// Auth rejection reason strings. These are returned as the 403 response body so
// callers can distinguish operator-actionable causes from token-integrity ones.
// They are a stable contract — automation may branch on them, so do not reword
// without considering downstream consumers.
const (
	// authErrAudienceMismatch is returned when the token's audience does not match
	// the request host. The operator likely minted the token for a different host.
	authErrAudienceMismatch = "auth token audience does not match the request host"
	// authErrMissingAudience is returned when the token carries no audience claim.
	authErrMissingAudience = "auth token is missing the audience claim"
	// authErrMissingKeyType is returned when the token lacks the key_type claim, or
	// it is not a string.
	authErrMissingKeyType = "auth token is missing or has an invalid key type claim"
	// authErrInvalidSubject is returned when the token's subject is not a valid
	// public key for the declared key type.
	authErrInvalidSubject = "auth token subject is not a valid public key"
	// authErrExpired is returned when the token's expiry (exp) is in the past.
	authErrExpired = "auth token has expired"
	// authErrNotYetValid is returned when the token's not-before (nbf) is in the future.
	authErrNotYetValid = "auth token is not yet valid"
	// authErrInvalidToken is returned for a malformed token or a token whose
	// signature does not verify. Signature failures are intentionally not
	// distinguished here, to avoid leaking information about key validity.
	authErrInvalidToken = "invalid auth token"
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
				http.Error(rw, authRejectionReason(err), http.StatusForbidden)
				return
			}

			audiences := append([]string{strings.ToLower(req.Host)}, allowedAudiences...)
			err = acpIdentity.VerifyAuthToken(ident, audiences...)
			if err != nil {
				http.Error(rw, authRejectionReason(err), http.StatusForbidden)
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

// authRejectionReason maps a token-construction or verification error to the
// documented 403 body. Any unrecognised cause — including signature failures —
// falls back to the generic invalid-token reason so no token-integrity detail is
// leaked.
func authRejectionReason(err error) string {
	switch {
	case errors.Is(err, acpIdentity.ErrMissingKeyType),
		errors.Is(err, acpIdentity.ErrInvalidKeyTypeClaimType):
		return authErrMissingKeyType
	case errors.Is(err, acpIdentity.ErrInvalidSubject):
		return authErrInvalidSubject
	case errors.Is(err, acpIdentity.ErrTokenExpired):
		return authErrExpired
	case errors.Is(err, acpIdentity.ErrTokenNotYetValid):
		return authErrNotYetValid
	case errors.Is(err, acpIdentity.ErrMissingAudience):
		return authErrMissingAudience
	case errors.Is(err, acpIdentity.ErrAudienceMismatch):
		return authErrAudienceMismatch
	default:
		return authErrInvalidToken
	}
}
