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
	errInvalidStoredValue string = "invalid stored value"
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
