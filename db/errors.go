// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"github.com/sourcenetwork/defradb/errors"
)

const (
	errFailedToGetHeads              string = "failed to get document heads"
	errFailedToCreateCollectionQuery string = "failed to create collection prefix query"
	errFailedToGetCollection         string = "failed to get collection"
	errDocVerification               string = "the document verification failed"
	errAddingP2PCollection           string = "cannot add collection ID"
	errRemovingP2PCollection         string = "cannot remove collection ID"
)

var (
	ErrFailedToGetHeads              = errors.New(errFailedToGetHeads)
	ErrFailedToCreateCollectionQuery = errors.New(errFailedToCreateCollectionQuery)
	ErrFailedToGetCollection         = errors.New(errFailedToGetCollection)
	// ErrDocVerification occurs when a documents contents fail the verification during a Create()
	// call against the supplied Document Key.
	ErrDocVerification         = errors.New(errDocVerification)
	ErrSubscriptionsNotAllowed = errors.New("server does not accept subscriptions")
	ErrDeleteTargetEmpty       = errors.New("the doc delete targeter cannot be empty")
	ErrDeleteEmpty             = errors.New("the doc delete cannot be empty")
	ErrUpdateTargetEmpty       = errors.New("the doc update targeter cannot be empty")
	ErrUpdateEmpty             = errors.New("the doc update cannot be empty")
	ErrInvalidMergeValueType   = errors.New(
		"the type of value in the merge patch doesn't match the schema",
	)
	ErrMissingDocFieldToUpdate  = errors.New("missing document field to update")
	ErrDocMissingKey            = errors.New("document is missing key")
	ErrMergeSubTypeNotSupported = errors.New("merge doesn't support sub types yet")
	ErrInvalidFilter            = errors.New("invalid filter")
	ErrInvalidOpPath            = errors.New("invalid patch op path")
	ErrDocumentAlreadyExists    = errors.New("a document with the given dockey already exists")
	ErrUnknownCRDTArgument      = errors.New("invalid CRDT arguments")
	ErrUnknownCRDT              = errors.New("unknown crdt")
	ErrSchemaFirstFieldDocKey   = errors.New("collection schema first field must be a DocKey")
	ErrCollectionAlreadyExists  = errors.New("collection already exists")
	ErrCollectionNameEmpty      = errors.New("collection name can't be empty")
	ErrSchemaIdEmpty            = errors.New("schema ID can't be empty")
	ErrSchemaVersionIdEmpty     = errors.New("schema version ID can't be empty")
	ErrKeyEmpty                 = errors.New("key cannot be empty")
	ErrAddingP2PCollection      = errors.New(errAddingP2PCollection)
	ErrRemovingP2PCollection    = errors.New(errRemovingP2PCollection)
)

// NewErrFailedToGetHeads returns a new error indicating that the heads of a document
// could not be obtained.
func NewErrFailedToGetHeads(inner error) error {
	return errors.Wrap(errFailedToGetHeads, inner)
}

// NewErrFailedToCreateCollectionQuery returns a new error indicating that the query
// to create a collection failed.
func NewErrFailedToCreateCollectionQuery(inner error) error {
	return errors.Wrap(errFailedToCreateCollectionQuery, inner)
}

// NewErrFailedToGetCollection returns a new error indicating that the collection could not be obtained.
func NewErrFailedToGetCollection(name string, inner error) error {
	return errors.Wrap(errFailedToGetCollection, inner, errors.NewKV("Name", name))
}

// NewErrDocVerification returns a new error indicating that the document verification failed.
func NewErrDocVerification(expected string, actual string) error {
	return errors.New(
		errDocVerification,
		errors.NewKV("Expected", expected),
		errors.NewKV("Actual", actual),
	)
}

// NewErrAddingP2PCollection returns a new error indicating that adding a collection ID to the
// persisted list of P2P collection IDs was not successful.
func NewErrAddingP2PCollection(inner error) error {
	return errors.Wrap(errAddingP2PCollection, inner)
}

// NewErrRemovingP2PCollection returns a new error indicating that removing a collection ID to the
// persisted list of P2P collection IDs was not successful.
func NewErrRemovingP2PCollection(inner error) error {
	return errors.Wrap(errRemovingP2PCollection, inner)
}
