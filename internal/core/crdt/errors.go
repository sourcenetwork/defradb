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
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
)

const (
	errFailedToGetPriority    string = "failed to get priority"
	errFailedToStoreValue     string = "failed to store value"
	errNegativeValue          string = "value cannot be negative"
	errUnsupportedCounterType string = "unsupported counter type. Valid types are int64 and float64"
	errGetRegisterStatus      string = "failed to get LWW register status"
	errGetRegisterValue       string = "failed to get current LWW register value"
	errDeleteRegisterVal      string = "failed to delete LWW register value"
	errSerializeLWWValue      string = "failed to serialize LWW field value"
	errCheckCounterExists     string = "failed to check if counter exists"
	errGenerateCounterNonce   string = "failed to generate counter nonce"
	errDecodeCounterValue     string = "failed to decode counter value"
	errGetCurrentCounterValue string = "failed to get current counter value"
	errGetCounterStatus       string = "failed to get counter status"
	errIncrementCounter       string = "failed to increment counter"
	errSetDocAsDeleted        string = "failed to set document as deleted"
	errGetDocMarker           string = "failed to get document marker"
	errSetDocVersion          string = "failed to set document version"
	errCreateDeleteIter       string = "failed to create iterator for document deletion"
	errSetDeletedFlag         string = "failed to set deleted flag on field"
	errDeleteFieldValue       string = "failed to delete field value"
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
	ErrMismatchedMergeType    = errors.New("given type to merge does not match source")
	ErrUnsupportedCounterType = errors.New(errUnsupportedCounterType)
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

func NewErrUnsupportedCounterType(valueType client.ScalarKind) error {
	return errors.New(errUnsupportedCounterType, errors.NewKV("Type", valueType))
}

func NewErrGetRegisterStatus(inner error, docID string, field string) error {
	return errors.Wrap(errGetRegisterStatus, inner,
		errors.NewKV("DocID", docID), errors.NewKV("Field", field))
}

func NewErrGetRegisterValue(inner error, docID string, field string) error {
	return errors.Wrap(errGetRegisterValue, inner,
		errors.NewKV("DocID", docID), errors.NewKV("Field", field))
}

func NewErrDeleteRegisterValue(inner error, docID string, field string) error {
	return errors.Wrap(errDeleteRegisterVal, inner,
		errors.NewKV("DocID", docID), errors.NewKV("Field", field))
}

func NewErrSerializeLWWValue(inner error, field string) error {
	return errors.Wrap(errSerializeLWWValue, inner, errors.NewKV("Field", field))
}

func NewErrCheckCounterExists(inner error, docID string, field string) error {
	return errors.Wrap(errCheckCounterExists, inner,
		errors.NewKV("DocID", docID), errors.NewKV("Field", field))
}

func NewErrGenerateCounterNonce(inner error) error {
	return errors.Wrap(errGenerateCounterNonce, inner)
}

func NewErrDecodeCounterValue(inner error) error {
	return errors.Wrap(errDecodeCounterValue, inner)
}

func NewErrGetCurrentCounterValue(inner error) error {
	return errors.Wrap(errGetCurrentCounterValue, inner)
}

func NewErrGetCounterStatus(inner error, docID string, field string) error {
	return errors.Wrap(errGetCounterStatus, inner,
		errors.NewKV("DocID", docID), errors.NewKV("Field", field))
}

func NewErrIncrementCounter(inner error, docID string, field string, kind string) error {
	return errors.Wrap(errIncrementCounter, inner,
		errors.NewKV("DocID", docID), errors.NewKV("Field", field), errors.NewKV("Kind", kind))
}

func NewErrSetDocAsDeleted(inner error, docID string) error {
	return errors.Wrap(errSetDocAsDeleted, inner, errors.NewKV("DocID", docID))
}

func NewErrGetDocMarker(inner error, docID string) error {
	return errors.Wrap(errGetDocMarker, inner, errors.NewKV("DocID", docID))
}

func NewErrSetDocVersion(inner error, docID string) error {
	return errors.Wrap(errSetDocVersion, inner, errors.NewKV("DocID", docID))
}

func NewErrCreateDeleteIter(inner error, docID string) error {
	return errors.Wrap(errCreateDeleteIter, inner, errors.NewKV("DocID", docID))
}

func NewErrSetDeletedFlag(inner error, docID string, key string) error {
	return errors.Wrap(errSetDeletedFlag, inner, errors.NewKV("DocID", docID), errors.NewKV("Key", key))
}

func NewErrDeleteFieldValue(inner error, docID string, key string) error {
	return errors.Wrap(errDeleteFieldValue, inner, errors.NewKV("DocID", docID), errors.NewKV("Key", key))
}
