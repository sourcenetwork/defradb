// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

import "github.com/sourcenetwork/defradb/errors"

const (
	errFieldNotExist         string = "The given field does not exist"
	errSelectOfNonGroupField string = "cannot select a non-group-by field at group-level"
)

// Errors returnable from this package.
//
// This list is incomplete and undefined errors may also be returned.
// Errors returned from this package may be tested against these errors with errors.Is.
var (
	ErrFieldNotExist         = errors.New(errFieldNotExist)
	ErrSelectOfNonGroupField = errors.New(errSelectOfNonGroupField)
	ErrFieldNotObject        = errors.New("Trying to access field on a non object type")
	ErrValueTypeMismatch     = errors.New("Value does not match indicated type")
	ErrIndexNotFound         = errors.New("No index found for given ID")
	ErrDocumentNotFound      = errors.New("No document for the given key exists")
	ErrInvalidUpdateTarget   = errors.New("The target document to update is of invalid type")
	ErrInvalidUpdater        = errors.New("The updater of a document is of invalid type")
	ErrInvalidDeleteTarget   = errors.New("The target document to delete is of invalid type")
)

func NewErrFieldNotExist(name string) error {
	return errors.New(errFieldNotExist, errors.NewKV("Name", name))
}

func NewErrSelectOfNonGroupField(name string) error {
	return errors.New(errSelectOfNonGroupField, errors.NewKV("Field", name))
}
