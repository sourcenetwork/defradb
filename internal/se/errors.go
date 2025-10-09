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
