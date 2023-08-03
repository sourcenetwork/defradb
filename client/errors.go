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
	errFieldNotExist        string = "The given field does not exist"
	errUnexpectedType       string = "unexpected type"
	errParsingFailed        string = "failed to parse argument"
	errUninitializeProperty string = "invalid state, required property is uninitialized"
	errMaxTxnRetries        string = "reached maximum transaction reties"
	errRelationOneSided     string = "relation must be defined on both schemas"
)

// Errors returnable from this package.
//
// This list is incomplete and undefined errors may also be returned.
// Errors returned from this package may be tested against these errors with errors.Is.
var (
	ErrFieldNotExist        = errors.New(errFieldNotExist)
	ErrUnexpectedType       = errors.New(errUnexpectedType)
	ErrParsingFailed        = errors.New(errParsingFailed)
	ErrUninitializeProperty = errors.New(errUninitializeProperty)
	ErrFieldNotObject       = errors.New("trying to access field on a non object type")
	ErrValueTypeMismatch    = errors.New("value does not match indicated type")
	ErrIndexNotFound        = errors.New("no index found for given ID")
	ErrDocumentNotFound     = errors.New("no document for the given key exists")
	ErrInvalidUpdateTarget  = errors.New("the target document to update is of invalid type")
	ErrInvalidUpdater       = errors.New("the updater of a document is of invalid type")
	ErrInvalidDeleteTarget  = errors.New("the target document to delete is of invalid type")
	ErrMalformedDocKey      = errors.New("malformed DocKey, missing either version or cid")
	ErrInvalidDocKeyVersion = errors.New("invalid DocKey version")
	ErrMaxTxnRetries        = errors.New(errMaxTxnRetries)
	ErrRelationOneSided     = errors.New(errRelationOneSided)
)

// NewErrFieldNotExist returns an error indicating that the given field does not exist.
func NewErrFieldNotExist(name string) error {
	return errors.New(errFieldNotExist, errors.NewKV("Name", name))
}

// NewErrFieldIndexNotExist returns an error indicating that a field does not exist at the
// given location.
func NewErrFieldIndexNotExist(index int) error {
	return errors.New(errFieldNotExist, errors.NewKV("Index", index))
}

// NewErrUnexpectedType returns an error indicating that the given value is of an unexpected type.
func NewErrUnexpectedType[TExpected any](property string, actual any) error {
	var expected TExpected
	return errors.WithStack(
		ErrUnexpectedType,
		errors.NewKV("Property", property),
		errors.NewKV("Expected", fmt.Sprintf("%T", expected)),
		errors.NewKV("Actual", fmt.Sprintf("%T", actual)),
	)
}

// NewErrUnhandledType returns an error indicating that the given value is of
// a type that is not handled.
func NewErrUnhandledType(property string, actual any) error {
	return errors.WithStack(
		ErrUnexpectedType,
		errors.NewKV("Property", property),
		errors.NewKV("Actual", fmt.Sprintf("%T", actual)),
	)
}

// NewErrParsingFailed returns an error indicating that the given argument could not be parsed.
func NewErrParsingFailed(inner error, argumentName string) error {
	return errors.Wrap(errParsingFailed, inner, errors.NewKV("Argument", argumentName))
}

// NewErrUninitializeProperty returns an error indicating that the required property
// is uninitialized.
func NewErrUninitializeProperty(host string, propertyName string) error {
	return errors.New(
		errUninitializeProperty,
		errors.NewKV("Host", host),
		errors.NewKV("PropertyName", propertyName),
	)
}

// NewErrFieldIndexNotExist returns an error indicating that a field does not exist at the
// given location.
func NewErrMaxTxnRetries(inner error) error {
	return errors.Wrap(errMaxTxnRetries, inner)
}

func NewErrRelationOneSided(fieldName string, typeName string) error {
	return errors.New(
		errRelationOneSided,
		errors.NewKV("Field", fieldName),
		errors.NewKV("Type", typeName),
	)
}
