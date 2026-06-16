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
	// authErrInvalidToken is returned for a malformed token or a token whose
	// signature does not verify. Signature failures are intentionally not
	// distinguished here, to avoid leaking information about key validity.
	authErrInvalidToken = "invalid auth token"
)

// AuthMiddleware authenticates an actor and sets their identity for all subsequent actions.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		token := strings.TrimPrefix(req.Header.Get(authHeaderName), authSchemaPrefix)
		if token == "" {
			next.ServeHTTP(rw, req)
			return
		}

		ident, err := acpIdentity.FromToken([]byte(token))
		if err != nil {
			http.Error(rw, authErrInvalidToken, http.StatusForbidden)
			return
		}

		err = acpIdentity.VerifyAuthToken(ident, strings.ToLower(req.Host))
		if err != nil {
			http.Error(rw, authRejectionReason(err), http.StatusForbidden)
			return
		}

		ctx := iIdentity.WithContext(req.Context(), immutable.Some[acpIdentity.Identity](ident))
		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}

// authRejectionReason maps a VerifyAuthToken error to the documented 403 body.
// Any unrecognised cause (including signature failures) falls back to the generic
// invalid-token reason so no token-integrity detail is leaked.
func authRejectionReason(err error) string {
	switch {
	case errors.Is(err, acpIdentity.ErrAudienceMismatch):
		return authErrAudienceMismatch
	case errors.Is(err, acpIdentity.ErrMissingAudience):
		return authErrMissingAudience
	default:
		return authErrInvalidToken
	}
}
