// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cli

import (
	"strings"

	"github.com/sourcenetwork/defradb/errors"
)

const (
	errMissingArg                  string = "missing argument"
	errMissingArgs                 string = "missing arguments"
	errTooManyArgs                 string = "too many arguments"
	errEmptyStdin                  string = "empty stdin"
	errEmptyFile                   string = "empty file"
	errFailedToReadFile            string = "failed to read file"
	errFailedToReadStdin           string = "failed to read stdin"
	errFailedToCreateRPCClient     string = "failed to create RPC client"
	errFailedToAddReplicator       string = "failed to add replicator, request failed"
	errFailedToJoinEndpoint        string = "failed to join endpoint"
	errFailedToSendRequest         string = "failed to send request"
	errFailedToReadResponseBody    string = "failed to read response body"
	errFailedToCloseResponseBody   string = "failed to close response body"
	errFailedToStatStdOut          string = "failed to stat stdout"
	errFailedToHandleGQLErrors     string = "failed to handle GraphQL errors"
	errFailedToPrettyPrintResponse string = "failed to pretty print response"
	errFailedToUnmarshalResponse   string = "failed to unmarshal response"
	errFailedParsePeerID           string = "failed to parse PeerID"
	errFailedToMarshalData         string = "failed to marshal data"
	errInvalidArgumentLength       string = "invalid argument length"
)

// Errors returnable from this package.
//
// This list is incomplete and undefined errors may also be returned.
// Errors returned from this package may be tested against these errors with errors.Is.
var (
	ErrMissingArg                  = errors.New(errMissingArg)
	ErrMissingArgs                 = errors.New(errMissingArgs)
	ErrTooManyArgs                 = errors.New(errTooManyArgs)
	ErrEmptyFile                   = errors.New(errEmptyFile)
	ErrEmptyStdin                  = errors.New(errEmptyStdin)
	ErrFailedToReadFile            = errors.New(errFailedToReadFile)
	ErrFailedToReadStdin           = errors.New(errFailedToReadStdin)
	ErrFailedToCreateRPCClient     = errors.New(errFailedToCreateRPCClient)
	ErrFailedToAddReplicator       = errors.New(errFailedToAddReplicator)
	ErrFailedToJoinEndpoint        = errors.New(errFailedToJoinEndpoint)
	ErrFailedToSendRequest         = errors.New(errFailedToSendRequest)
	ErrFailedToReadResponseBody    = errors.New(errFailedToReadResponseBody)
	ErrFailedToStatStdOut          = errors.New(errFailedToStatStdOut)
	ErrFailedToHandleGQLErrors     = errors.New(errFailedToHandleGQLErrors)
	ErrFailedToPrettyPrintResponse = errors.New(errFailedToPrettyPrintResponse)
	ErrFailedToUnmarshalResponse   = errors.New(errFailedToUnmarshalResponse)
	ErrFailedParsePeerID           = errors.New(errFailedParsePeerID)
	ErrInvalidExportFormat         = errors.New("invalid export format")
	ErrInvalidArgumentLength       = errors.New(errInvalidArgumentLength)
)

func NewErrMissingArg(name string) error {
	return errors.New(errMissingArg, errors.NewKV("Name", name))
}

func NewErrMissingArgs(names []string) error {
	return errors.New(errMissingArgs, errors.NewKV("Required", strings.Join(names, ", ")))
}

func NewErrTooManyArgs(max, actual int) error {
	return errors.New(errTooManyArgs, errors.NewKV("Max", max), errors.NewKV("Actual", actual))
}

func NewFailedToReadFile(inner error) error {
	return errors.Wrap(errFailedToReadFile, inner)
}

func NewFailedToReadStdin(inner error) error {
	return errors.Wrap(errFailedToReadStdin, inner)
}

func NewErrFailedToCreateRPCClient(inner error) error {
	return errors.Wrap(errFailedToCreateRPCClient, inner)
}

func NewErrFailedToAddReplicator(inner error) error {
	return errors.Wrap(errFailedToAddReplicator, inner)
}

func NewErrFailedToJoinEndpoint(inner error) error {
	return errors.Wrap(errFailedToJoinEndpoint, inner)
}

func NewErrFailedToSendRequest(inner error) error {
	return errors.Wrap(errFailedToSendRequest, inner)
}

func NewErrFailedToReadResponseBody(inner error) error {
	return errors.Wrap(errFailedToReadResponseBody, inner)
}

func NewErrFailedToCloseResponseBody(closeErr, other error) error {
	if other != nil {
		return errors.Wrap(errFailedToCloseResponseBody, closeErr, errors.NewKV("Other error", other))
	}
	return errors.Wrap(errFailedToCloseResponseBody, closeErr)
}

func NewErrFailedToStatStdOut(inner error) error {
	return errors.Wrap(errFailedToStatStdOut, inner)
}

func NewErrFailedToHandleGQLErrors(inner error) error {
	return errors.Wrap(errFailedToHandleGQLErrors, inner)
}

func NewErrFailedToPrettyPrintResponse(inner error) error {
	return errors.Wrap(errFailedToPrettyPrintResponse, inner)
}

func NewErrFailedToUnmarshalResponse(inner error) error {
	return errors.Wrap(errFailedToUnmarshalResponse, inner)
}

func NewErrFailedParsePeerID(inner error) error {
	return errors.Wrap(errFailedParsePeerID, inner)
}

// NewFailedToMarshalData returns an error indicating that a there was a problem with mashalling.
func NewFailedToMarshalData(inner error) error {
	return errors.Wrap(errFailedToMarshalData, inner)
}

// NewErrInvalidArgumentLength returns an error indicating an incorrect number of arguments.
func NewErrInvalidArgumentLength(inner error, expected int) error {
	return errors.Wrap(errInvalidArgumentLength, inner, errors.NewKV("Expected", expected))
}
