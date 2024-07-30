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
	errSelectOfNonGroupField string = "cannot select a non-group-by field at group-level"
)

// Errors returnable from this package.
//
// This list is incomplete and undefined errors may also be returned.
// Errors returned from this package may be tested against these errors with errors.Is.
var (
	ErrSelectOfNonGroupField  = errors.New(errSelectOfNonGroupField)
	ErrMissingOperationName   = errors.New("request with multiple operations must have an operationName")
	ErrMissingQueryOrMutation = errors.New("request is missing query or mutation operation statements")
)

// NewErrSelectOfNonGroupField returns an error indicating that a non-group-by field
// was selected at group-level.
func NewErrSelectOfNonGroupField(name string) error {
	return errors.New(errSelectOfNonGroupField, errors.NewKV("Field", name))
}
