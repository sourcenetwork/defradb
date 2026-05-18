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
	"net/http"

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/acp"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/internal/db/p2p"
)

const (
	errFailedToLoadKeys         string = "failed to load given keys"
	errMethodIsNotImplemented   string = "the method is not implemented"
	errFailedToGetContext       string = "failed to get context"
	errMissingRequiredParameter string = "required parameter %s is missing"
	errCollectionNotFound       string = "collection not found"
)

// Errors returnable from this package.
//
// This list is incomplete. Undefined errors may also be returned.
// Errors returned from this package may be tested against these errors with errors.Is.
var (
	ErrNoListener                   = errors.New("cannot serve with no listener")
	ErrNoEmail                      = errors.New("email address must be specified for tls with autocert")
	ErrInvalidRequestBody           = errors.New("invalid request body")
	ErrStreamingNotSupported        = errors.New("streaming not supported")
	ErrMigrationNotFound            = errors.New("migration not found")
	ErrMissingRequest               = errors.New("missing request")
	ErrInvalidTransactionId         = errors.New("invalid transaction id")
	ErrP2PDisabled                  = errors.New("p2p network is disabled")
	ErrMethodIsNotImplemented       = errors.New(errMethodIsNotImplemented)
	ErrMissingIdentity              = errors.New("required identity is missing")
	ErrInvalidSubscriptionTransport = errors.New("invalid subscription transport")
	ErrInvalidGraphQLRequest        = errors.New("invalid graphql request")
	ErrTransactionNotFound          = errors.New("transaction not found")
)

type errorResponse struct {
	Error error `json:"error"`
}

func (e errorResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{"error": e.Error.Error()})
}

func (e *errorResponse) UnmarshalJSON(data []byte) error {
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return err
	}
	if msg, ok := out["error"].(string); ok {
		e.Error = client.ReviveError(msg)
	} else {
		e.Error = fmt.Errorf("%s", out)
	}
	return nil
}

func NewErrFailedToGetContext(contextType string) error {
	return errors.New(
		errFailedToGetContext,
		errors.NewKV("ContextType", contextType),
	)
}

func NewErrFailedToLoadKeys(inner error, publicKeyPath, privateKeyPath string) error {
	return errors.Wrap(
		errFailedToLoadKeys,
		inner,
		errors.NewKV("PublicKeyPath", publicKeyPath),
		errors.NewKV("PrivateKeyPath", privateKeyPath),
	)
}

// NewErrMissingRequiredParameter creates a new error for a missing required parameter
func NewErrMissingRequiredParameter(paramName string) error {
	return errors.New(fmt.Sprintf(errMissingRequiredParameter, paramName))
}

func NewErrCollectionNotFound(collectionName string) error {
	return errors.New(
		errCollectionNotFound,
		errors.NewKV("CollectionName", collectionName),
	)
}

// httpStatusFromError maps known error types to appropriate HTTP status codes.
func httpStatusFromError(err error) int {
	// 401 Unauthorized
	if errors.Is(err, client.ErrNotAuthorizedToPerformOperation) {
		return http.StatusUnauthorized
	}

	// 403 Forbidden
	if errors.Is(err, client.ErrOperationRequiresDeveloperMode) ||
		errors.Is(err, client.ErrCanNotDoThisNACOpWithNACIsDisabled) ||
		errors.Is(err, db.ErrMissingPermission) ||
		errors.Is(err, acp.ErrResourceIsMissingRequiredPermission) {
		return http.StatusForbidden
	}

	// 404 Not Found
	if errors.Is(err, client.ErrDocumentNotFoundOrNotAuthorized) ||
		errors.Is(err, client.ErrCollectionNotFound) ||
		errors.Is(err, db.ErrDocIDNotFound) ||
		errors.Is(err, db.ErrIndexWithNameDoesNotExists) ||
		errors.Is(err, db.ErrEncryptedIndexDoesNotExist) ||
		errors.Is(err, db.ErrCollectionRootNotFound) ||
		errors.Is(err, db.ErrLensCIDNotFound) ||
		errors.Is(err, p2p.ErrReplicatorNotFound) ||
		errors.Is(err, acp.ErrPolicyDoesNotExistWithACP) ||
		errors.Is(err, acp.ErrResourceDoesNotExistOnTargetPolicy) {
		return http.StatusNotFound
	}

	// 409 Conflict
	if errors.Is(err, db.ErrCollectionAlreadyExists) ||
		errors.Is(err, db.ErrDocumentAlreadyExists) ||
		errors.Is(err, db.ErrIndexWithNameAlreadyExists) ||
		errors.Is(err, db.ErrEncryptedIndexAlreadyExists) ||
		errors.Is(err, db.ErrReplicatorExists) ||
		errors.Is(err, db.ErrMultipleActiveCollectionVersions) ||
		errors.Is(err, corekv.ErrTxnConflict) {
		return http.StatusConflict
	}

	// 422 Unprocessable Entity
	if errors.Is(err, db.ErrCanNotHavePolicyWithoutACP) ||
		errors.Is(err, db.ErrMaterializedViewAndACPNotSupported) ||
		errors.Is(err, db.ErrColNotMaterialized) ||
		errors.Is(err, db.ErrColMutatingIsBranchable) ||
		errors.Is(err, db.ErrP2PColHasPolicy) ||
		errors.Is(err, db.ErrReplicatorColHasPolicy) ||
		errors.Is(err, db.ErrCollectionNameMutated) ||
		errors.Is(err, db.ErrCannotDeleteOldVersion) ||
		errors.Is(err, db.ErrMigrationBetweenNonAdjacentVersions) ||
		errors.Is(err, db.ErrNACIsAlreadyDisabled) ||
		errors.Is(err, db.ErrNACIsAlreadyEnabled) ||
		errors.Is(err, client.ErrACPOperationButACPNotAvailable) {
		return http.StatusUnprocessableEntity
	}

	// 503 Service Unavailable
	if errors.Is(err, ErrP2PDisabled) {
		return http.StatusServiceUnavailable
	}

	// 400 Bad Request (default)
	return http.StatusBadRequest
}
