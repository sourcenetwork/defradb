// Copyright 2023 Democratized Data Foundation
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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sourcenetwork/immutable"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore/badger/v4"
)

// Using Basic right now, but this will soon change to 'Bearer' as acp authentication
// gets implemented: https://github.com/sourcenetwork/defradb/issues/2017
const authSchemaPrefix = "Basic "

// Name of authorization header
const authHeaderName = "Authorization"

func requestJSON(req *http.Request, out any) error {
	data, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}

func responseJSON(rw http.ResponseWriter, status int, out any) {
	rw.Header().Add("Content-Type", "application/json")
	rw.WriteHeader(status)
	json.NewEncoder(rw).Encode(out) //nolint:errcheck
}

func parseError(msg any) error {
	switch msg {
	case client.ErrDocumentNotFoundOrNotAuthorized.Error():
		return client.ErrDocumentNotFoundOrNotAuthorized
	case badger.ErrTxnConflict.Error():
		return badger.ErrTxnConflict
	default:
		return fmt.Errorf("%s", msg)
	}
}

// addIdentityToAuthHeader adds the identity to auth header as it must always exist.
func addIdentityToAuthHeader(req *http.Request, identity string) {
	// Create a bearer that will get added to authorization header.
	bearerWithIdentity := authSchemaPrefix + identity

	// Add the authorization header with the bearer containing identity.
	req.Header.Add(authHeaderName, bearerWithIdentity)
}

// addIdentityToAuthHeaderIfExists adds the identity to auth header if it exsits, otherwise does nothing.
func addIdentityToAuthHeaderIfExists(req *http.Request, identity immutable.Option[string]) {
	// Do nothing if there is no identity to add.
	if !identity.HasValue() {
		return
	}
	addIdentityToAuthHeader(req, identity.Value())
}

// getIdentityFromAuthHeader tries to get the identity from the auth header, if it is found
// with the expecte auth schema then it is returned, otherwise no identity is returned.
func getIdentityFromAuthHeader(req *http.Request) immutable.Option[string] {
	authHeader := req.Header.Get(authHeaderName)
	if authHeader == "" {
		return acpIdentity.NoIdentity
	}

	identity := strings.TrimPrefix(authHeader, authSchemaPrefix)
	// If expected schema prefix was not found, or empty, then assume no identity.
	if identity == authHeader || identity == "" {
		return acpIdentity.NoIdentity
	}

	return acpIdentity.NewIdentity(identity)
}
