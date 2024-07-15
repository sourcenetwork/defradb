// Copyright 2024 Democratized Data Foundation
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

	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/sourcenetwork/immutable"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/internal/db"
)

const (
	// authHeaderName is the name of the authorization header.
	// This header should contain an ACP identity.
	authHeaderName = "Authorization"
	// authSchemaPrefix is the prefix added to the
	// authorization header value.
	authSchemaPrefix = "Bearer "
)

// verifyAuthToken verifies that the jwt auth token is valid and that the signature
// matches the identity of the subject.
func verifyAuthToken(identity acpIdentity.Identity, audience string) error {
	_, err := jwt.Parse([]byte(identity.BearerToken), jwt.WithVerify(false), jwt.WithAudience(audience))
	if err != nil {
		return err
	}

	_, err = jws.Verify(
		[]byte(identity.BearerToken),
		jws.WithKey(acpIdentity.BearerTokenSignatureScheme, identity.PublicKey.ToECDSA()),
	)
	if err != nil {
		return err
	}
	return nil
}

// AuthMiddleware authenticates an actor and sets their identity for all subsequent actions.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		token := strings.TrimPrefix(req.Header.Get(authHeaderName), authSchemaPrefix)
		if token == "" {
			next.ServeHTTP(rw, req)
			return
		}

		identity, err := acpIdentity.FromToken([]byte(token))
		if err != nil {
			http.Error(rw, "forbidden", http.StatusForbidden)
			return
		}

		err = verifyAuthToken(identity, strings.ToLower(req.Host))
		if err != nil {
			http.Error(rw, "forbidden", http.StatusForbidden)
			return
		}

		ctx := db.SetContextIdentity(req.Context(), immutable.Some(identity))
		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}
