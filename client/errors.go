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

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/errors"
)

const (
	errFieldNotExist                       string = "the given field does not exist"
	errUnexpectedType                      string = "unexpected type"
	errParsingFailed                       string = "failed to parse argument"
	errUninitializeProperty                string = "invalid state, required property is uninitialized"
	errMaxTxnRetries                       string = "reached maximum transaction reties"
	errCollectionNotFound                  string = "collection not found"
	errUnknownCRDT                         string = "unknown crdt"
	errCRDTKindMismatch                    string = "CRDT type %s can't be assigned to field kind %s"
	errInvalidCRDTType                     string = "CRDT type not supported"
	errFailedToUnmarshalCollection         string = "failed to unmarshal collection json"
	errOperationNotPermittedOnNamelessCols string = "operation not permitted on nameless collection"
	errInvalidJSONPayload                  string = "invalid JSON payload"
	errCanNotNormalizeValue                string = "can not normalize value"
	errCanNotTurnNormalValueIntoArray      string = "can not turn normal value into array"
	errCanNotMakeNormalNilFromFieldKind    string = "can not make normal nil from field kind"
	errFailedToParseKind                   string = "failed to parse kind"
	errCannotSetRelationFromSecondarySide  string = "cannot set relation from secondary side"
	errArraySizeMismatch                   string = "array size mismatch"
	errInvalidTypeForEmbedding             string = "invalid type for vector embedding"
	errInvalidTypeForEmbeddingGeneration   string = "invalid field type for vector embedding generation"
	errEmptyFieldNameForEmbedding          string = "embedding FieldName cannot be empty"
	errEmptyFieldsForEmbedding             string = "embedding Fields cannot be empty"
	errEmptyProviderForEmbedding           string = "embedding Provider cannot be empty"
	errEmptyModelForEmbedding              string = "embedding Model cannot be empty"
	errUnknownEmbeddingProvider            string = "unknown embedding provider"
	errEmbeddingFieldEmbedding             string = "embedding fields cannot refer to self or another embedding field"
)

var (
	errNotFound string = corekv.ErrNotFound.Error()
)

// Errors returnable from this package.
//
// This list is incomplete and undefined errors may also be returned.
// Errors returned from this package may be tested against these errors with errors.Is.
var (
	ErrFieldNotExist                        = errors.New(errFieldNotExist)
	ErrUnexpectedType                       = errors.New(errUnexpectedType)
	ErrFailedToUnmarshalCollection          = errors.New(errFailedToUnmarshalCollection)
	ErrOperationNotPermittedOnNamelessCols  = errors.New(errOperationNotPermittedOnNamelessCols)
	ErrFieldNotObject                       = errors.New("trying to access field on a non object type")
	ErrValueTypeMismatch                    = errors.New("value does not match indicated type")
	ErrDocumentNotFoundOrNotAuthorized      = errors.New("document not found or not authorized to access")
	ErrACPOperationButACPNotAvailable       = errors.New("operation requires ACP, but ACP not available")
	ErrACPOperationButCollectionHasNoPolicy = errors.New("operation requires ACP, but collection has no policy")
	ErrInvalidUpdateTarget                  = errors.New("the target document to update is of invalid type")
	ErrInvalidUpdater                       = errors.New("the updater of a document is of invalid type")
	ErrInvalidDeleteTarget                  = errors.New("the target document to delete is of invalid type")
	ErrMalformedDocID                       = errors.New("malformed document ID, missing either version or cid")
	ErrInvalidDocIDVersion                  = errors.New("invalid document ID version")
	ErrInvalidJSONPayload                   = errors.New(errInvalidJSONPayload)
	ErrCanNotNormalizeValue                 = errors.New(errCanNotNormalizeValue)
	ErrCanNotTurnNormalValueIntoArray       = errors.New(errCanNotTurnNormalValueIntoArray)
	ErrCanNotMakeNormalNilFromFieldKind     = errors.New(errCanNotMakeNormalNilFromFieldKind)
	ErrCollectionNotFound                   = errors.New(errCollectionNotFound)
	ErrFailedToParseKind                    = errors.New(errFailedToParseKind)
	ErrArraySizeMismatch                    = errors.New(errArraySizeMismatch)
	ErrInvalidTypeForEmbedding              = errors.New(errInvalidTypeForEmbedding)
	ErrInvalidTypeForEmbeddingGeneration    = errors.New(errInvalidTypeForEmbeddingGeneration)
	ErrEmptyFieldNameForEmbedding           = errors.New(errEmptyFieldNameForEmbedding)
	ErrEmptyFieldsForEmbedding              = errors.New(errEmptyFieldsForEmbedding)
	ErrEmptyProviderForEmbedding            = errors.New(errEmptyProviderForEmbedding)
	ErrEmptyModelForEmbedding               = errors.New(errEmptyModelForEmbedding)
	ErrUnknownEmbeddingProvider             = errors.New(errUnknownEmbeddingProvider)
	ErrEmbeddingFieldEmbedding              = errors.New(errEmbeddingFieldEmbedding)
	ErrNotFound                             = errors.New(errNotFound)
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

// NewCanNotNormalizeValue returns an error indicating that the given value can not be normalized.
func NewCanNotNormalizeValue(val any) error {
	return errors.New(errCanNotNormalizeValue, errors.NewKV("Value", val))
}

// NewCanNotTurnNormalValueIntoArray returns an error indicating that the given value can not be
// turned into an array.
func NewCanNotTurnNormalValueIntoArray(val any) error {
	return errors.New(errCanNotTurnNormalValueIntoArray, errors.NewKV("Value", val))
}

// NewCanNotMakeNormalNilFromFieldKind returns an error indicating that a normal nil value can not be
// created from the given field kind.
func NewCanNotMakeNormalNilFromFieldKind(kind FieldKind) error {
	return errors.New(errCanNotMakeNormalNilFromFieldKind, errors.NewKV("Kind", kind))
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

func NewErrCollectionNotFoundForSchemaVersion(schemaVersionID string) error {
	return errors.New(
		errCollectionNotFound,
		errors.NewKV("SchemaVersionID", schemaVersionID),
	)
}

func NewErrCollectionNotFoundForSchema(schemaRoot string) error {
	return errors.New(
		errCollectionNotFound,
		errors.NewKV("SchemaRoot", schemaRoot),
	)
}

func NewErrUnknownCRDT(cType CType) error {
	return errors.New(
		errUnknownCRDT,
		errors.NewKV("Type", cType),
	)
}

func NewErrInvalidCRDTType(name, crdtType string) error {
	return errors.New(
		errInvalidCRDTType,
		errors.NewKV("Name", name),
		errors.NewKV("CRDTType", crdtType),
	)
}

func NewErrCRDTKindMismatch(cType, kind string) error {
	return errors.New(fmt.Sprintf(errCRDTKindMismatch, cType, kind))
}

func NewErrInvalidJSONPayload(payload any) error {
	return errors.New(errInvalidJSONPayload, errors.NewKV("Payload", payload))
}

func NewErrFailedToParseKind(kind []byte) error {
	return errors.New(
		errCRDTKindMismatch,
		errors.NewKV("Kind", kind),
	)
}

// ReviveError attempts to return a client specific error from
// the given message. If no matching error is found the message
// is wrapped in a new anonymous error type.
func ReviveError(message string) error {
	switch message {
	case ErrDocumentNotFoundOrNotAuthorized.Error():
		return ErrDocumentNotFoundOrNotAuthorized
	case corekv.ErrTxnConflict.Error():
		return corekv.ErrTxnConflict
	default:
		return fmt.Errorf("%s", message)
	}
}

func NewErrCannotSetRelationFromSecondarySide(name string) error {
	return errors.New(errCannotSetRelationFromSecondarySide, errors.NewKV("Name", name))
}

func NewErrArraySizeMismatch[T any](array []T, expected int) error {
	return errors.New(errArraySizeMismatch, errors.NewKV("Actual", len(array)), errors.NewKV("Expected", expected))
}

func NewErrInvalidTypeForEmbedding(actual FieldKind) error {
	return errors.New(errInvalidTypeForEmbedding, errors.NewKV("Actual", actual.String()))
}

func NewErrFieldForEmbeddingGenerationDoesNotExist(fieldName string) error {
	return errors.New(errFieldNotExist, errors.NewKV("Embedding generation field", fieldName))
}

func NewErrVectorFieldDoesNotExist(fieldName string) error {
	return errors.New(errFieldNotExist, errors.NewKV("Vector field", fieldName))
}

func NewErrInvalidTypeForEmbeddingGeneration(actual FieldKind) error {
	return errors.New(errInvalidTypeForEmbeddingGeneration, errors.NewKV("Actual", actual.String()))
}

func NewErrUnknownEmbeddingProvider(provider string) error {
	return errors.New(errUnknownEmbeddingProvider, errors.NewKV("Provider", provider))
}

func NewErrEmbeddingFieldEmbedding(fieldName string) error {
	return errors.New(errEmbeddingFieldEmbedding, errors.NewKV("Field", fieldName))
}

func NewErrNotFound(kv errors.KV) error {
	return errors.New(errNotFound, kv)
}
