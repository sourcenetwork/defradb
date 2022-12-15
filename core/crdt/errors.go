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
)

// Errors returnable from this package.
//
// This list is incomplete and undefined errors may also be returned.
// Errors returned from this package may be tested against these errors with errors.Is.
var (
	ErrFailedToGetPriority = errors.New(errFailedToGetPriority)
	ErrFailedToStoreValue  = errors.New(errFailedToStoreValue)
	ErrEncodingPriority    = errors.New("error encoding priority")
	ErrDecodingPriority    = errors.New("error decoding priority")
)

func NewErrFailedToGetPriority(inner error) error {
	return errors.Wrap(errFailedToGetPriority, inner)
}

func NewErrFailedToStoreValue(inner error) error {
	return errors.Wrap(errFailedToStoreValue, inner)
}
