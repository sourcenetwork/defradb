// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package schema

import "github.com/sourcenetwork/defradb/errors"

const (
	errDuplicateField                string = "duplicate field"
	errFieldMissingRelation          string = "field missing associated relation"
	errRelationMissingField          string = "relation missing field"
	errAggregateTargetNotFound       string = "aggregate target not found"
	errSchemaTypeAlreadyExist        string = "schema type already exists"
	errMutationInputTypeAlreadyExist string = "mutation input type already exists"
	errObjectNotFoundDuringThunk     string = "object not found whilst executing fields thunk"
	errTypeNotFound                  string = "no type found for given name"
	errRelationNotFound              string = "no relation found"
	errNonNullForTypeNotSupported    string = "NonNull variants for type are not supported"
	errIndexMissingFields            string = "index missing fields"
	errIndexUnknownArgument          string = "index with unknown argument"
	errIndexInvalidArgument          string = "index with invalid argument"
	errIndexInvalidName              string = "index with invalid name"
	errPolicyUnknownArgument         string = "policy with unknown argument"
	errPolicyInvalidIDProp           string = "policy directive with invalid id property"
	errPolicyInvalidResourceProp     string = "policy directive with invalid resource property"
	errDefaultValueInvalid           string = "default value type must match field type"
	errDefaultValueNotAllowed        string = "default value is not allowed for this field type"
)

var (
	ErrDuplicateField                = errors.New(errDuplicateField)
	ErrFieldMissingRelation          = errors.New(errFieldMissingRelation)
	ErrRelationMissingField          = errors.New(errRelationMissingField)
	ErrAggregateTargetNotFound       = errors.New(errAggregateTargetNotFound)
	ErrSchemaTypeAlreadyExist        = errors.New(errSchemaTypeAlreadyExist)
	ErrMutationInputTypeAlreadyExist = errors.New(errMutationInputTypeAlreadyExist)
	ErrObjectNotFoundDuringThunk     = errors.New(errObjectNotFoundDuringThunk)
	ErrTypeNotFound                  = errors.New(errTypeNotFound)
	ErrRelationNotFound              = errors.New(errRelationNotFound)
	ErrNonNullForTypeNotSupported    = errors.New(errNonNullForTypeNotSupported)
	ErrRelationMutlipleTypes         = errors.New("relation type can only be either One or Many, not both")
	ErrRelationMissingTypes          = errors.New("relation is missing its defined types and fields")
	ErrRelationInvalidType           = errors.New("relation has an invalid type to be finalize")
	ErrMultipleRelationPrimaries     = errors.New("relation can only have a single field set as primary")
	// NonNull is the literal name of the GQL type, so we have to disable the linter
	//nolint:revive
	ErrNonNullNotSupported       = errors.New("NonNull fields are not currently supported")
	ErrIndexMissingFields        = errors.New(errIndexMissingFields)
	ErrIndexWithUnknownArg       = errors.New(errIndexUnknownArgument)
	ErrIndexWithInvalidArg       = errors.New(errIndexInvalidArgument)
	ErrPolicyWithUnknownArg      = errors.New(errPolicyUnknownArgument)
	ErrPolicyInvalidIDProp       = errors.New(errPolicyInvalidIDProp)
	ErrPolicyInvalidResourceProp = errors.New(errPolicyInvalidResourceProp)
)

func NewErrDuplicateField(objectName, fieldName string) error {
	return errors.New(
		errDuplicateField,
		errors.NewKV("Object", objectName),
		errors.NewKV("Field", fieldName),
	)
}

func NewErrIndexWithInvalidName(name string) error {
	return errors.New(errIndexInvalidName, errors.NewKV("Name", name))
}

func NewErrFieldMissingRelation(objectName, fieldName string, objectType string) error {
	return errors.New(
		errFieldMissingRelation,
		errors.NewKV("Object", objectName),
		errors.NewKV("Field", fieldName),
		errors.NewKV("ObjectType", objectType),
	)
}

func NewErrRelationMissingField(objectName, fieldName string) error {
	return errors.New(
		errRelationMissingField,
		errors.NewKV("Object", objectName),
		errors.NewKV("Field", fieldName),
	)
}

func NewErrAggregateTargetNotFound(objectName, target string) error {
	return errors.New(
		errAggregateTargetNotFound,
		errors.NewKV("Object", objectName),
		errors.NewKV("Target", target),
	)
}

func NewErrSchemaTypeAlreadyExist(name string) error {
	return errors.New(
		errSchemaTypeAlreadyExist,
		errors.NewKV("Name", name),
	)
}

func NewErrMutationInputTypeAlreadyExist(name string) error {
	return errors.New(
		errMutationInputTypeAlreadyExist,
		errors.NewKV("Name", name),
	)
}

func NewErrObjectNotFoundDuringThunk(object string) error {
	return errors.New(
		errObjectNotFoundDuringThunk,
		errors.NewKV("Object", object),
	)
}

func NewErrTypeNotFound(typeName string) error {
	return errors.New(
		errTypeNotFound,
		errors.NewKV("Type", typeName),
	)
}

func NewErrNonNullForTypeNotSupported(typeName string) error {
	return errors.New(
		errNonNullForTypeNotSupported,
		errors.NewKV("Type", typeName),
	)
}

func NewErrRelationNotFound(relationName string) error {
	return errors.New(
		errRelationNotFound,
		errors.NewKV("RelationName", relationName),
	)
}

func NewErrDefaultValueInvalid(expectedType string, actualType string) error {
	return errors.New(
		errDefaultValueInvalid,
		errors.NewKV("ExpectedType", expectedType),
		errors.NewKV("ActualType", actualType),
	)
}

func NewErrDefaultValueNotAllowed(fieldName, fieldType string) error {
	return errors.New(
		errDefaultValueNotAllowed,
		errors.NewKV("Name", fieldName),
		errors.NewKV("Type", fieldType),
	)
}
