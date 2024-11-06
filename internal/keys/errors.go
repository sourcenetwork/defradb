// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package keys

import (
	"github.com/sourcenetwork/defradb/errors"
)

const (
	errFailedToGetFieldIdOfKey string = "failed to get FieldID of Key"
	errInvalidFieldIndex       string = "invalid field index"
	errInvalidFieldValue       string = "invalid field value"
)

var (
	ErrFailedToGetFieldIdOfKey = errors.New(errFailedToGetFieldIdOfKey)
	ErrEmptyKey                = errors.New("received empty key string")
	ErrInvalidKey              = errors.New("invalid key string")
	ErrInvalidFieldIndex       = errors.New(errInvalidFieldIndex)
	ErrInvalidFieldValue       = errors.New(errInvalidFieldValue)
)

// NewErrFailedToGetFieldIdOfKey returns the error indicating failure to get FieldID of Key.
func NewErrFailedToGetFieldIdOfKey(inner error) error {
	return errors.Wrap(errFailedToGetFieldIdOfKey, inner)
}

// NewErrInvalidFieldIndex returns the error indicating invalid field index.
func NewErrInvalidFieldIndex(i int) error {
	return errors.New(errInvalidFieldIndex, errors.NewKV("index", i))
}

// NewErrInvalidFieldValue returns the error indicating invalid field value.
func NewErrInvalidFieldValue(reason string) error {
	return errors.New(errInvalidFieldValue, errors.NewKV("Reason", reason))
}
