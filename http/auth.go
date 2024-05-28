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
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/lestrrat-go/jwx/v2/jwa"
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

var authTokenSignatureScheme = jwa.ES256K

// buildAuthToken returns a new jwt auth token with the subject and audience set
// to the given values. Default expiration and not before values will also be set.
func buildAuthToken(identity acpIdentity.Identity, audience string) (jwt.Token, error) {
	if identity.PublicKey == nil {
		return nil, ErrMissingIdentityPublicKey
	}
	subject := hex.EncodeToString(identity.PublicKey.SerializeCompressed())
	return jwt.NewBuilder().
		Subject(subject).
		Audience([]string{audience}).
		Expiration(time.Now().Add(15 * time.Minute)).
		NotBefore(time.Now()).
		Build()
}

// signAuthToken returns a signed jwt auth token that can be used to authenticate the
// actor identified by the given identity with a defraDB node identified by the given audience.
func signAuthToken(identity acpIdentity.Identity, token jwt.Token) ([]byte, error) {
	if identity.PrivateKey == nil {
		return nil, ErrMissingIdentityPrivateKey
	}
	return jwt.Sign(token, jwt.WithKey(authTokenSignatureScheme, identity.PrivateKey.ToECDSA()))
}

// buildAndSignAuthToken returns a signed jwt auth token that can be used to authenticate the
// actor identified by the given identity with a defraDB node identified by the given audience.
func buildAndSignAuthToken(identity acpIdentity.Identity, audience string) ([]byte, error) {
	token, err := buildAuthToken(identity, audience)
	if err != nil {
		return nil, err
	}
	return signAuthToken(identity, token)
}

// verifyAuthToken verifies that the jwt auth token is valid and that the signature
// matches the identity of the subject.
func verifyAuthToken(data []byte, audience string) (immutable.Option[acpIdentity.Identity], error) {
	token, err := jwt.Parse(data, jwt.WithVerify(false), jwt.WithAudience(audience))
	if err != nil {
		return immutable.None[acpIdentity.Identity](), err
	}
	subject, err := hex.DecodeString(token.Subject())
	if err != nil {
		return immutable.None[acpIdentity.Identity](), err
	}
	pubKey, err := secp256k1.ParsePubKey(subject)
	if err != nil {
		return immutable.None[acpIdentity.Identity](), err
	}
	_, err = jws.Verify(data, jws.WithKey(authTokenSignatureScheme, pubKey.ToECDSA()))
	if err != nil {
		return immutable.None[acpIdentity.Identity](), err
	}
	return acpIdentity.FromPublicKey(pubKey), nil
}

// AuthMiddleware authenticates an actor and sets their identity for all subsequent actions.
func AuthMiddleware(audience string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			token := strings.TrimPrefix(req.Header.Get(authHeaderName), authSchemaPrefix)
			if token == "" {
				next.ServeHTTP(rw, req)
				return
			}
			identity, err := verifyAuthToken([]byte(token), audience)
			if err != nil {
				http.Error(rw, "forbidden", http.StatusForbidden)
				return
			}
			ctx := db.SetContextIdentity(req.Context(), identity)
			next.ServeHTTP(rw, req.WithContext(ctx))
		})
	}
}
