// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package crdt

import (
	"github.com/sourcenetwork/defradb/errors"
)

const (
	errFailedToGetPriority string = "failed to get priority"
	errFailedToStoreValue  string = "failed to store value"
	errNegativeValue       string = "value cannot be negative"
)

// Errors returnable from this package.
//
// This list is incomplete and undefined errors may also be returned.
// Errors returned from this package may be tested against these errors with errors.Is.
var (
	ErrFailedToGetPriority = errors.New(errFailedToGetPriority)
	ErrFailedToStoreValue  = errors.New(errFailedToStoreValue)
	ErrNegativeValue       = errors.New(errNegativeValue)
	ErrEncodingPriority    = errors.New("error encoding priority")
	ErrDecodingPriority    = errors.New("error decoding priority")
	// ErrMismatchedMergeType - Tying to merge two ReplicatedData of different types
	ErrMismatchedMergeType = errors.New("given type to merge does not match source")
)

// NewErrFailedToGetPriority returns an error indicating that the priority could not be retrieved.
func NewErrFailedToGetPriority(inner error) error {
	return errors.Wrap(errFailedToGetPriority, inner)
}

// NewErrFailedToStoreValue returns an error indicating that the value could not be stored.
func NewErrFailedToStoreValue(inner error) error {
	return errors.Wrap(errFailedToStoreValue, inner)
}

func NewErrNegativeValue[T Incrementable](value T) error {
	return errors.New(errNegativeValue, errors.NewKV("Value", value))
}
