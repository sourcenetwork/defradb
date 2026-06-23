// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package se

import (
	"github.com/sourcenetwork/defradb/errors"
)

const (
	errEmptyDocID                  = "DocID must not be empty"
	errFailedToGetEncIndexes       = "failed to get encrypted indexes"
	errUnsupportedIndexType        = "unsupported encrypted index type"
	errFailedToDeserializeBlock    = "failed to deserialize block"
	errFailedToGenerateSEArtifacts = "failed to generate SE artifacts"
	errCollectionNotFound          = "collection not found"
	errStoreSEArtifact             = "failed to store SE artifact"
	errGetCollectionIDForSE        = "failed to get collection ID for SE query"
	errCreateSEQueryIterator       = "failed to create SE query iterator"
)

func NewErrEmptyDocID(key string) error {
	return errors.New(errEmptyDocID, errors.NewKV("Key", key))
}

func NewErrFailedToGetEncryptedIndexes(inner error) error {
	return errors.Wrap(errFailedToGetEncIndexes, inner)
}

func NewErrUnsupportedIndexType(indexType string) error {
	return errors.New(errUnsupportedIndexType, errors.NewKV("Type", indexType))
}

func NewErrFailedToDeserializeBlock(inner error) error {
	return errors.Wrap(errFailedToDeserializeBlock, inner)
}

func NewErrFailedToGenerateSEArtifacts(inner error) error {
	return errors.Wrap(errFailedToGenerateSEArtifacts, inner)
}

func NewErrCollectionNotFound(collectionID string) error {
	return errors.New(errCollectionNotFound, errors.NewKV("CollectionID", collectionID))
}

func NewErrStoreSEArtifact(inner error, docID string, collectionID string) error {
	return errors.Wrap(errStoreSEArtifact, inner,
		errors.NewKV("DocID", docID), errors.NewKV("CollectionID", collectionID))
}

func NewErrGetCollectionIDForSE(inner error, collectionID string) error {
	return errors.Wrap(errGetCollectionIDForSE, inner, errors.NewKV("CollectionID", collectionID))
}

func NewErrCreateSEQueryIterator(inner error) error {
	return errors.Wrap(errCreateSEQueryIterator, inner)
}
