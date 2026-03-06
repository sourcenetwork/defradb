// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package request

import (
	"github.com/sourcenetwork/defradb/errors"
)

const (
	errSelectOfNonGroupField    string = "cannot select a non-group-by field at group-level"
	errCursorMustContainQuery   string = "_cursor block must contain exactly one collection query"
	errMultipleQueriesInCursor  string = "_cursor block cannot contain multiple collection queries"
	errFirstMustBeNonNegative   string = "first must be non-negative"
	errLastMustBeNonNegative    string = "last must be non-negative"
	errForwardBackwardConflict  string = "forward parameters (first/after) cannot be combined with backward parameters (last/before)"
	errInvalidCursor            string = "invalid cursor"
)

// Errors returnable from this package.
//
// This list is incomplete and undefined errors may also be returned.
// Errors returned from this package may be tested against these errors with errors.Is.
var (
	ErrSelectOfNonGroupField   = errors.New(errSelectOfNonGroupField)
	ErrCursorMustContainQuery  = errors.New(errCursorMustContainQuery)
	ErrMultipleQueriesInCursor = errors.New(errMultipleQueriesInCursor)
	ErrFirstMustBeNonNegative  = errors.New(errFirstMustBeNonNegative)
	ErrLastMustBeNonNegative   = errors.New(errLastMustBeNonNegative)
	ErrForwardBackwardConflict = errors.New(errForwardBackwardConflict)
	ErrInvalidCursor           = errors.New(errInvalidCursor)
)

// NewErrSelectOfNonGroupField returns an error indicating that a non-group-by field
// was selected at group-level.
func NewErrSelectOfNonGroupField(name string) error {
	return errors.New(errSelectOfNonGroupField, errors.NewKV("Field", name))
}
