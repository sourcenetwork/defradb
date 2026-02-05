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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"

	"net/http"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
)

const (
	errFailedToLoadKeys             string = "failed to load given keys"
	errMethodIsNotImplemented       string = "the method is not implemented"
	errFailedToGetContext           string = "failed to get context"
	errPurgeRequestNonDeveloperMode string = "cannot purge database when development mode is disabled"
	errMissingRequiredParameter     string = "required parameter %s is missing"
	errCollectionNotFound           string = "collection not found"
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
)

func getErrorStatus(err error, defaultStatus int) int {
	var jsonSyntaxErr *json.SyntaxError
	var jsonUnmarshalErr *json.UnmarshalTypeError
	var hexInvalidByteErr hex.InvalidByteError
	switch {
	case errors.Is(err, client.ErrNotAuthorizedToPerformOperation):
		return http.StatusUnauthorized
	case errors.Is(err, client.ErrCollectionNotFound),
		errors.Is(err, client.ErrNotFound),
		errors.Is(err, client.ErrDocumentNotFoundOrNotAuthorized),
		errors.Is(err, ErrMigrationNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrNoListener),
		errors.Is(err, ErrNoEmail),
		errors.Is(err, ErrInvalidRequestBody),
		errors.Is(err, ErrStreamingNotSupported),
		errors.Is(err, ErrMissingRequest),
		errors.Is(err, ErrInvalidTransactionId),
		errors.Is(err, client.ErrInvalidDocIDVersion),
		errors.Is(err, client.ErrInvalidJSONPayload),
		errors.Is(err, ErrInvalidGraphQLRequest),
		errors.Is(err, io.EOF),
		errors.Is(err, io.ErrUnexpectedEOF),
		errors.As(err, &jsonSyntaxErr),
		errors.As(err, &jsonUnmarshalErr),
		errors.As(err, &hexInvalidByteErr),
		errors.Is(err, hex.ErrLength):
		return http.StatusBadRequest
	default:
		return defaultStatus
		
	}
}

func responseError(rw http.ResponseWriter, err error, defaultStatus int) {
	responseJSON(rw, getErrorStatus(err, defaultStatus), errorResponse{err})
}

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
