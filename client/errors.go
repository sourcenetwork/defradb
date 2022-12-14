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

import (
	"fmt"

	"github.com/sourcenetwork/defradb/errors"
)

const (
	errFieldNotExist         string = "The given field does not exist"
	errSelectOfNonGroupField string = "cannot select a non-group-by field at group-level"
	errUnexpectedType        string = "unexpected type"
)

// Errors returnable from this package.
//
// This list is incomplete and undefined errors may also be returned.
// Errors returned from this package may be tested against these errors with errors.Is.
var (
	ErrFieldNotExist         = errors.New(errFieldNotExist)
	ErrSelectOfNonGroupField = errors.New(errSelectOfNonGroupField)
	ErrUnexpectedType        = errors.New(errUnexpectedType)
	ErrFieldNotObject        = errors.New("trying to access field on a non object type")
	ErrValueTypeMismatch     = errors.New("value does not match indicated type")
	ErrIndexNotFound         = errors.New("no index found for given ID")
	ErrDocumentNotFound      = errors.New("no document for the given key exists")
	ErrInvalidUpdateTarget   = errors.New("the target document to update is of invalid type")
	ErrInvalidUpdater        = errors.New("the updater of a document is of invalid type")
	ErrInvalidDeleteTarget   = errors.New("the target document to delete is of invalid type")
)

func NewErrFieldNotExist(name string) error {
	return errors.New(errFieldNotExist, errors.NewKV("Name", name))
}

func NewErrSelectOfNonGroupField(name string) error {
	return errors.New(errSelectOfNonGroupField, errors.NewKV("Field", name))
}

func NewErrUnexpectedType[TExpected any](property string, actual any) error {
	var expected TExpected
	return errors.WithStack(
		ErrUnexpectedType,
		errors.NewKV("Property", property),
		errors.NewKV("Expected", fmt.Sprintf("%T", expected)),
		errors.NewKV("Actual", fmt.Sprintf("%T", actual)),
	)
}
