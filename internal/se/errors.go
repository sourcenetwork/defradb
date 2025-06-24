// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package se

import (
	"github.com/sourcenetwork/defradb/errors"
)

const (
	errEmptyDocID            = "DocID must not be empty"
	errFailedToGetEncIndexes = "failed to get encrypted indexes"
	errFailedToGetFieldValue = "failed to get field value"
	errUnsupportedIndexType  = "unsupported encrypted index type"
)

func NewErrEmptyDocID(key string) error {
	return errors.New(errEmptyDocID, errors.NewKV("Key", key))
}

func NewErrFailedToGetEncryptedIndexes(inner error) error {
	return errors.Wrap(errFailedToGetEncIndexes, inner)
}

func NewErrFailedToGetFieldValue(fieldName string, inner error) error {
	return errors.Wrap(errFailedToGetFieldValue, inner, errors.NewKV("FieldName", fieldName))
}

func NewErrUnsupportedIndexType(indexType string) error {
	return errors.New(errUnsupportedIndexType, errors.NewKV("Type", indexType))
}
