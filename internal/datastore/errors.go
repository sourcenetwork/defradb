// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package datastore

import (
	"github.com/sourcenetwork/defradb/errors"
)

const (
	errInvalidStoredValue    string = "invalid stored value"
	errStoreBlock            string = "failed to store block"
	errCheckBlockExists      string = "failed to check if block exists"
	errCheckBlockMergeStatus string = "failed to check block merge status"
	errMarkBlockAsMerged     string = "failed to mark block as merged"
	errDeserializePrefix     string = "failed to deserialize prefix query result"
	errFetchKeysForPrefix    string = "failed to fetch keys for prefix"
)

// Errors returnable from this package.
//
// This list is incomplete and undefined errors may also be returned.
// Errors returned from this package may be tested against these errors with errors.Is.
var (
	// ErrHashMismatch is an error returned when the hash of a block is different than expected.
	ErrHashMismatch = errors.New("block in storage has different hash than requested")
)

// NewErrInvalidStoredValue returns a new error indicating that the stored
// value in the database is invalid.
func NewErrInvalidStoredValue(inner error) error {
	return errors.Wrap(errInvalidStoredValue, inner)
}

// NewErrStoreBlock returns a new error indicating that a block could not be stored.
func NewErrStoreBlock(inner error) error {
	return errors.Wrap(errStoreBlock, inner)
}

// NewErrCheckBlockExists returns a new error indicating that checking block existence failed.
func NewErrCheckBlockExists(inner error) error {
	return errors.Wrap(errCheckBlockExists, inner)
}

// NewErrCheckBlockMergeStatus returns a new error indicating that checking block merge status failed.
func NewErrCheckBlockMergeStatus(inner error) error {
	return errors.Wrap(errCheckBlockMergeStatus, inner)
}

// NewErrMarkBlockAsMerged returns a new error indicating that marking a block as merged failed.
func NewErrMarkBlockAsMerged(inner error) error {
	return errors.Wrap(errMarkBlockAsMerged, inner)
}

// NewErrDeserializePrefix returns a new error indicating that deserializing a prefix query result failed.
func NewErrDeserializePrefix(inner error) error {
	return errors.Wrap(errDeserializePrefix, inner)
}

// NewErrFetchKeysForPrefix returns a new error indicating that fetching keys for a prefix failed.
func NewErrFetchKeysForPrefix(inner error) error {
	return errors.Wrap(errFetchKeysForPrefix, inner)
}
