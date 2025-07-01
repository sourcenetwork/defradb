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
func AuthMiddleware(next http.Handler) http.Handler {
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

		err = acpIdentity.VerifyAuthToken(ident, strings.ToLower(req.Host))
		if err != nil {
			http.Error(rw, "forbidden", http.StatusForbidden)
			return
		}

		identity, ok := ident.(acpIdentity.Identity)
		if !ok {
			http.Error(rw, "forbidden", http.StatusForbidden)
			return
		}
		ctx := acpIdentity.WithContext(req.Context(), immutable.Some(identity))
		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}
